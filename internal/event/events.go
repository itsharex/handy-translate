// Package event 集中管理前后端事件通信（观察者模式）。
// 所有事件名作为常量统一定义，避免硬编码字符串散落各处。
package event

import (
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ──────────────────────────────────────────────
// 事件名称常量
// ──────────────────────────────────────────────

const (
	// 翻译相关
	Query                = "query"                  // 发送查询文本到前端
	Result               = "result"                 // 发送翻译结果（非流式）
	ResultStream         = "result_stream"           // 流式翻译数据块
	ResultMeaningsStream = "result_meanings_stream"  // 流式释义数据块
	ResultStreamDone     = "result_stream_done"      // 流式传输完成
	ResultStreamError    = "result_stream_error"     // 流式传输错误
	WordQueryResult      = "word_query_result"       // 单词查询结果（JSON）
	Explains             = "explains"                // 释义结果

	// 语言设置
	TranslateLang = "translateLang" // 前端设置翻译语言

	// 工具栏模式
	ToolbarMode        = "toolbarMode"        // 前端切换工具栏模式
	ToolbarModeUpdated = "toolbarModeUpdated" // 通知前端模式已更新

	// 截图
	ScreenshotBase64 = "screenshotBase64" // 全屏截图 base64 数据

	// 工具栏固定状态
	ToolbarPinnedUpdated    = "toolbarPinnedUpdated"    // 通知前端固定状态已更新
	ToolbarSlideDirection   = "toolbarSlideDirection"   // 通知前端弹窗滑入方向
)

// ──────────────────────────────────────────────
// EventBus — 对 Wails App 事件系统的封装
// ──────────────────────────────────────────────

// Emitter 抽象事件发射接口，方便测试时 Mock。
type Emitter interface {
	EmitEvent(eventName string, data ...interface{})
}

// Bus 封装 Wails 的事件发射和监听，提供类型安全的事件操作。
type Bus struct {
	app *application.App
}

// NewBus 创建 EventBus 实例。
func NewBus(app *application.App) *Bus {
	return &Bus{app: app}
}

// Emit 发射一个事件（兼容 Wails v3 的 app.Event.Emit 签名）。
func (b *Bus) Emit(eventName string, data ...interface{}) {
	if len(data) > 0 {
		b.app.Event.Emit(eventName, data[0])
	} else {
		b.app.Event.Emit(eventName)
	}
}

// On 监听一个事件。
func (b *Bus) On(eventName string, handler func(event *application.CustomEvent)) {
	b.app.Event.On(eventName, handler)
}

// EmitEvent 实现 Emitter 接口。
func (b *Bus) EmitEvent(eventName string, data ...interface{}) {
	b.Emit(eventName, data...)
}

// ──────────────────────────────────────────────
// 便捷的发射方法
// ──────────────────────────────────────────────

// EmitQuery 发送查询文本到前端。
func (b *Bus) EmitQuery(queryText string) {
	b.Emit(Query, queryText)
}

// EmitResult 发送完整翻译结果。
func (b *Bus) EmitResult(result string) {
	b.Emit(Result, result)
}

// EmitResultStream 发送流式翻译数据（完整累积文本）。
func (b *Bus) EmitResultStream(fullText string) {
	b.Emit(ResultStream, fullText)
}

// EmitResultMeaningsStream 发送流式释义数据。
func (b *Bus) EmitResultMeaningsStream(fullText string) {
	b.Emit(ResultMeaningsStream, fullText)
}

// EmitStreamDone 通知前端流式传输完成。
func (b *Bus) EmitStreamDone() {
	b.Emit(ResultStreamDone, "done")
}

// EmitStreamError 通知前端流式传输出错。
func (b *Bus) EmitStreamError(errMsg string) {
	slog.Error("流式传输错误", slog.String("error", errMsg))
	b.Emit(ResultStreamError, errMsg)
}

// EmitWordQueryResult 发送单词查询结果。
func (b *Bus) EmitWordQueryResult(jsonResult string) {
	b.Emit(WordQueryResult, jsonResult)
}

// EmitExplains 发送解释结果。
func (b *Bus) EmitExplains(explains string) {
	b.Emit(Explains, explains)
}

// EmitToolbarModeUpdated 通知前端工具栏模式已更新。
func (b *Bus) EmitToolbarModeUpdated(mode string) {
	b.Emit(ToolbarModeUpdated, mode)
}

// EmitScreenshotBase64 发送截图数据。
func (b *Bus) EmitScreenshotBase64(base64Image string) {
	b.Emit(ScreenshotBase64, base64Image)
}

// EmitToolbarPinnedUpdated 通知前端工具栏固定状态已更新。
func (b *Bus) EmitToolbarPinnedUpdated(pinned bool) {
	b.Emit(ToolbarPinnedUpdated, pinned)
}

// EmitToolbarSlideDirection 通知前端弹窗滑入方向。
// direction: "right-bottom", "right-top", "left-bottom", "left-top"
func (b *Bus) EmitToolbarSlideDirection(direction string) {
	b.Emit(ToolbarSlideDirection, direction)
}
