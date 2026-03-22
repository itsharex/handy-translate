// Package service 提供应用层业务编排（门面模式）。
// Translator 封装翻译/解释的完整流程，包括流式/非流式选择、事件发射、历史保存。
package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"handy-translate/config"
	"handy-translate/history"
	"handy-translate/internal/event"
	"handy-translate/internal/translate"
	"handy-translate/logger"
)

// Translator 翻译业务门面（Facade 模式）。
// 封装底层翻译提供者的调用，统一处理流式/非流式分支、事件发射、历史保存等。
type Translator struct {
	registry  *translate.Registry
	config    *config.Config
	eventBus  *event.Bus
	history   *history.HistoryService
	wordCache *WordCache
}

// NewTranslator 创建翻译服务门面。
func NewTranslator(
	registry *translate.Registry,
	cfg *config.Config,
	eventBus *event.Bus,
	historySvc *history.HistoryService,
	wordCache *WordCache,
) *Translator {
	return &Translator{
		registry:  registry,
		config:    cfg,
		eventBus:  eventBus,
		history:   historySvc,
		wordCache: wordCache,
	}
}

// Translate 执行翻译（自动选择流式/非流式）。
// 对于流式翻译，会通过 EventBus 发射 result_stream 事件；
// 对于非流式翻译，直接返回结果。
func (t *Translator) Translate(ctx context.Context, queryText, fromLang, toLang string) string {
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		slog.Error("获取翻译提供者失败", slog.String("error", err.Error()))
		return ""
	}

	// 尝试流式翻译
	if sp, ok := provider.(translate.StreamProvider); ok {
		return t.translateStream(ctx, sp, queryText, fromLang, toLang)
	}

	// 非流式翻译
	return t.translateNormal(ctx, provider, queryText, fromLang, toLang)
}

// TranslateStream 纯流式翻译（不返回结果，通过事件通知前端）。
func (t *Translator) TranslateStream(ctx context.Context, queryText, fromLang, toLang string) {
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		slog.Error("获取翻译提供者失败", slog.String("error", err.Error()))
		return
	}

	if sp, ok := provider.(translate.StreamProvider); ok {
		t.translateStream(ctx, sp, queryText, fromLang, toLang)
	} else {
		// 回退到普通翻译
		logger.LogStreamNotSupported(provider.Name())
		res := t.translateNormal(ctx, provider, queryText, fromLang, toLang)
		t.eventBus.EmitResult(res)
	}
}

// TranslateMeanings 翻译释义（支持流式）。
func (t *Translator) TranslateMeanings(ctx context.Context, queryText, fromLang, toLang string) string {
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		slog.Error("获取翻译提供者失败", slog.String("error", err.Error()))
		return ""
	}

	if sp, ok := provider.(translate.StreamProvider); ok {
		logger.LogTranslateStart(provider.Name(), fromLang, toLang, len(queryText))

		var streamResult strings.Builder
		chunkCount := 0
		startTime := time.Now()

		err := sp.TranslateStream(ctx, translate.TranslateRequest{
			Text: queryText, SourceLang: fromLang, TargetLang: toLang,
		}, func(chunk string) {
			streamResult.WriteString(chunk)
			chunkCount++
			logger.LogChunkReceived(chunkCount, len(chunk), streamResult.Len())
			t.eventBus.EmitResultMeaningsStream(streamResult.String())
		})

		if err != nil {
			elapsed := time.Since(startTime)
			logger.LogTranslateError(provider.Name(), err, chunkCount, elapsed)
			t.eventBus.EmitStreamError(err.Error())
			return ""
		}

		t.eventBus.EmitStreamDone()
		resultStr := streamResult.String()
		elapsed := time.Since(startTime)
		logger.LogTranslateSuccess(provider.Name(), len(queryText), len(resultStr), chunkCount, elapsed)

		t.saveTranslateHistory(queryText, resultStr, fromLang, toLang)
		return resultStr
	}

	// 非流式
	return t.translateNormal(ctx, provider, queryText, fromLang, toLang)
}

// Explain 流式解释（支持模板选择）。
func (t *Translator) Explain(ctx context.Context, queryText, templateID string) string {
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		slog.Error("获取翻译提供者失败", slog.String("error", err.Error()))
		return ""
	}

	sp, ok := provider.(translate.StreamProvider)
	if !ok {
		logger.LogStreamExplainNotSupported(provider.Name(), templateID)
		return ""
	}

	logger.LogExplainStart(provider.Name(), templateID, len(queryText))

	var streamResult strings.Builder
	chunkCount := 0
	startTime := time.Now()

	err = sp.ExplainStream(ctx, queryText, templateID, func(chunk string) {
		streamResult.WriteString(chunk)
		chunkCount++
		logger.LogChunkReceived(chunkCount, len(chunk), streamResult.Len())
		t.eventBus.EmitResultStream(streamResult.String())
	})

	if err != nil {
		elapsed := time.Since(startTime)
		logger.LogExplainError(provider.Name(), templateID, err, chunkCount, elapsed)
		t.eventBus.EmitStreamError(err.Error())
		return ""
	}

	t.eventBus.EmitStreamDone()
	resultStr := streamResult.String()
	elapsed := time.Since(startTime)
	logger.LogExplainSuccess(provider.Name(), templateID, len(queryText), len(resultStr), chunkCount, elapsed)

	t.saveExplainHistory(queryText, resultStr, templateID)
	return resultStr
}

// QueryWord 使用 LLM 查询单词详情（音标、词性、释义、例句——全部一次返回）。
// 通过 word_query_result 事件发送结果。优先使用文件缓存。
func (t *Translator) QueryWord(ctx context.Context, word string) {
	// 1. 缓存命中 → 延迟 100ms 再发（确保前端先处理 query 事件，避免竞态）
	if t.wordCache != nil {
		if cached, ok := t.wordCache.Get(word); ok {
			slog.Info("📦 缓存命中", slog.String("word", word))
			time.Sleep(100 * time.Millisecond) // 等前端处理 query 事件
			t.eventBus.EmitWordQueryResult(cached)
			return
		}
	}

	// 2. 缓存未命中 → 调用 LLM
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		slog.Error("获取翻译提供者失败", slog.String("error", err.Error()))
		return
	}

	sp, ok := provider.(translate.StreamProvider)
	if !ok {
		result, err := provider.Translate(ctx, translate.TranslateRequest{
			Text: word, SourceLang: "en", TargetLang: "zh",
		})
		if err != nil {
			slog.Error("QueryWord 翻译失败", slog.Any("err", err))
			return
		}
		t.eventBus.EmitWordQueryResult(strings.Join(result, "\n"))
		return
	}

	prompt := buildWordQueryPrompt(word)
	var streamResult strings.Builder

	err = sp.ExplainStream(ctx, prompt, "", func(chunk string) {
		streamResult.WriteString(chunk)
	})

	if err != nil {
		slog.Error("QueryWord 失败", slog.String("word", word), slog.Any("err", err))
		return
	}

	jsonResult := streamResult.String()
	slog.Info("QueryWord 完成", slog.String("word", word), slog.Int("result_length", len(jsonResult)))

	// 3. 保存到缓存
	if t.wordCache != nil {
		t.wordCache.Set(word, jsonResult)
	}

	t.eventBus.EmitWordQueryResult(jsonResult)
}

// WordCacheGet 检查单词缓存是否命中。
func (t *Translator) WordCacheGet(word string) (string, bool) {
	if t.wordCache != nil {
		return t.wordCache.Get(word)
	}
	return "", false
}

// IsStreamSupported 检查当前翻译服务是否支持流式。
func (t *Translator) IsStreamSupported() bool {
	provider, err := t.registry.GetFromConfig(t.config.TranslateWay, t.config)
	if err != nil {
		return false
	}
	_, ok := provider.(translate.StreamProvider)
	return ok
}

// ──────────────────────────────────────────────
// 内部方法
// ──────────────────────────────────────────────

func (t *Translator) translateStream(ctx context.Context, sp translate.StreamProvider, queryText, fromLang, toLang string) string {
	logger.LogTranslateStart(sp.Name(), fromLang, toLang, len(queryText))

	var streamResult strings.Builder
	chunkCount := 0
	startTime := time.Now()

	err := sp.TranslateStream(ctx, translate.TranslateRequest{
		Text: queryText, SourceLang: fromLang, TargetLang: toLang,
	}, func(chunk string) {
		streamResult.WriteString(chunk)
		chunkCount++
		logger.LogChunkReceived(chunkCount, len(chunk), streamResult.Len())
		t.eventBus.EmitResultStream(streamResult.String())
	})

	if err != nil {
		elapsed := time.Since(startTime)
		logger.LogTranslateError(sp.Name(), err, chunkCount, elapsed)
		t.eventBus.EmitStreamError(err.Error())
		return ""
	}

	t.eventBus.EmitStreamDone()
	resultStr := streamResult.String()
	elapsed := time.Since(startTime)
	logger.LogTranslateSuccess(sp.Name(), len(queryText), len(resultStr), chunkCount, elapsed)

	t.saveTranslateHistory(queryText, resultStr, fromLang, toLang)
	return resultStr
}

func (t *Translator) translateNormal(ctx context.Context, provider translate.Provider, queryText, fromLang, toLang string) string {
	logger.LogNormalTranslateStart(provider.Name(), fromLang, toLang)

	result, err := provider.Translate(ctx, translate.TranslateRequest{
		Text: queryText, SourceLang: fromLang, TargetLang: toLang,
	})
	if err != nil {
		logger.LogNormalTranslateError(provider.Name(), err)
		return ""
	}

	logger.LogNormalTranslateSuccess(provider.Name(), len(result))
	translateRes := strings.Join(result, "\n")

	t.saveTranslateHistory(queryText, translateRes, fromLang, toLang)
	return translateRes
}

func (t *Translator) saveTranslateHistory(queryText, result, fromLang, toLang string) {
	if t.config.History.Enabled && t.history != nil {
		go t.history.SaveTranslateRecord(queryText, result, fromLang, toLang)
	}
}

func (t *Translator) saveExplainHistory(queryText, result, templateID string) {
	if t.config.History.Enabled && t.history != nil {
		go t.history.SaveExplainRecord(queryText, result, templateID)
	}
}

// buildWordQueryPrompt 构造单词查询的 LLM 提示词。
func buildWordQueryPrompt(word string) string {
	return `查询英文单词"` + word + `"，严格按 JSON 返回，无其他内容：
{"word":"` + word + `","phonetic":"IPA音标","translation":"最常用的中文翻译(2-3个词)","meanings":[{"partOfSpeech":"词性英文","definitions":[{"definition":"英文释义","definitionZh":"中文释义","example":"英文例句","exampleZh":"中文翻译"}]}]}
要求：每个词性给1个最常用释义，含例句，只返回JSON`
}
