// Package translate 定义翻译服务提供者的接口（策略模式）。
package translate

import (
	"context"
)

// TranslateRequest 翻译请求参数。
type TranslateRequest struct {
	Text       string
	SourceLang string
	TargetLang string
}

// Provider 翻译服务提供者接口（策略模式）。
// 每个翻译服务（百度、有道、DeepSeek 等）只需实现此接口。
type Provider interface {
	// Name 返回提供者名称标识（如 "deepseek"、"baidu"）。
	Name() string
	// Translate 执行翻译，返回翻译结果。
	Translate(ctx context.Context, req TranslateRequest) ([]string, error)
}

// StreamProvider 支持流式输出的翻译服务提供者接口。
// 扩展 Provider，增加流式翻译和流式解释能力。
type StreamProvider interface {
	Provider
	// TranslateStream 流式翻译，通过 onChunk 回调返回每个数据块。
	TranslateStream(ctx context.Context, req TranslateRequest, onChunk func(chunk string)) error
	// ExplainStream 流式解释，通过 onChunk 回调返回每个数据块。
	ExplainStream(ctx context.Context, text, templateID string, onChunk func(chunk string)) error
}
