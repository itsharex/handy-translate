package deepseek

import (
	"context"
	"log"
	"sync"

	"handy-translate/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const (
	Way              = "deepseek"
	TranslatePrompts = `You are a professional translator.
			"Please translate the following text accurately and naturally.
			"Keep the original meaning, tone, and formatting.
			"Do not explain or add anything else.
			"If the text is Chinese, translate to English.
			"If the text is English, translate to Chinese.
			"Text:{{.text}}`
)

var (
	once      sync.Once
	llm       *openai.LLM
	llmErr    error
	llmErrMsg string
)

type Deepseek struct {
	config.Translate
}

type TranslationPayload struct {
	Source    []string `json:"source"`
	TransType string   `json:"trans_type"`
	RequestID string   `json:"request_id"`
	Detect    bool     `json:"detect"`
}

type TranslationResponse struct {
	Target []string `json:"target"`
}

func (c *Deepseek) GetName() string {
	return Way
}

func (c *Deepseek) GetLLM() (*openai.LLM, error) {
	once.Do(func() {
		llm, llmErr = openai.New(
			openai.WithToken(config.Data.Translate[Way].Key),
			openai.WithModel("deepseek-chat"),
			openai.WithBaseURL("https://api.deepseek.com"),
		)
		if llmErr != nil {
			llmErrMsg = "failed to initialize DeepSeek LLM: " + llmErr.Error()
			log.Printf("ERROR: %s", llmErrMsg)
		}
	})
	return llm, llmErr
}

func (c *Deepseek) PostQuery(query, fromLang, toLang string) ([]string, error) {
	llm, err := c.GetLLM()
	if err != nil {
		return nil, err
	}

	// 定义模板
	promptTemplate := prompts.NewPromptTemplate(
		TranslatePrompts,
		[]string{"text"},
	)

	// 构建输入
	promptValue, err := promptTemplate.Format(map[string]any{
		"text": query,
	})
	if err != nil {
		return nil, err
	}

	// 调用 LLM
	resp, err := llms.GenerateFromSinglePrompt(context.Background(), llm, promptValue)
	if err != nil {
		return nil, err
	}

	return []string{resp}, nil
}

// PostQueryStream 流式翻译
func (c *Deepseek) PostQueryStream(query, fromLang, toLang string, callback func(chunk string)) error {
	llm, err := c.GetLLM()
	if err != nil {
		return err
	}

	// 定义模板
	promptTemplate := prompts.NewPromptTemplate(TranslatePrompts,
		[]string{"text"},
	)

	// 构建输入
	promptValue, err := promptTemplate.Format(map[string]any{
		"text": query,
	})
	if err != nil {
		return err
	}

	// 流式调用 LLM
	ctx := context.Background()
	_, err = llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextPart(promptValue),
			},
			Role: llms.ChatMessageTypeHuman,
		},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 每次接收到数据块时调用回调函数
		if len(chunk) > 0 {
			callback(string(chunk))
		}
		return nil
	}))

	return err
}

// PostExplainStream 流式术语解释（支持模板选择）
func (c *Deepseek) PostExplainStream(query, templateID string, callback func(chunk string)) error {
	llm, err := c.GetLLM()
	if err != nil {
		return err
	}

	var promptValue string

	if templateID == "" {
		// 没有模板 ID 时，直接使用 query 作为完整提示词（如 QueryWord 场景）
		promptValue = query
	} else {
		// 获取模板内容
		templateStr := c.getTemplate(templateID)

		// 定义术语解释模板
		promptTemplate := prompts.NewPromptTemplate(
			templateStr,
			[]string{"text"},
		)

		// 构建输入
		promptValue, err = promptTemplate.Format(map[string]any{
			"text": query,
		})
		if err != nil {
			return err
		}
	}

	// 流式调用 LLM
	ctx := context.Background()
	_, err = llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextPart(promptValue),
			},
			Role: llms.ChatMessageTypeHuman,
		},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 每次接收到数据块时调用回调函数
		if len(chunk) > 0 {
			callback(string(chunk))
		}
		return nil
	}))

	return err
}

// getTemplate 获取提示词模板，委托到共用模板查找逻辑。
func (c *Deepseek) getTemplate(templateID string) string {
	return config.FindTemplate(&config.Data.ExplainTemplates, templateID)
}

