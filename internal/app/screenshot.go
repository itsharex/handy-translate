// Package app 完整的截图+OCR+翻译流程。
package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"log/slog"

	"handy-translate/config"
	"handy-translate/internal/event"
	"handy-translate/internal/service"
	"handy-translate/internal/window"
	"handy-translate/window/screenshot"
)

// HandleCaptureSelectedScreen 截取选中区域 → OCR → 翻译。
func (a *Application) HandleCaptureSelectedScreen(startX, startY, width, height float64) {
	croppedImg := screenshot.CaptureSelectedScreen(int(startX), int(startY), int(width), int(height))
	if croppedImg == nil {
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, croppedImg); err != nil {
		slog.Error("png.Encode", slog.Any("err", err))
		return
	}

	filename := "screenshot.png"
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())

	if err := service.SaveBase64Image(base64String, filename); err != nil {
		slog.Error("saveBase64Image", slog.Any("err", err))
		return
	}

	// OCR 解析文本
	queryText := a.OCR.Recognize(filename)
	a.State.SetCurrentQuery(queryText)
	a.WindowMgr.ResetToolbarState()

	// 发送查询文本
	a.EventBus.EmitQuery(queryText)

	// 根据模式翻译或解释
	mode := a.State.GetToolbarMode()
	ctx := context.Background()
	fl, tl := a.State.GetLangs()

	if a.Translator.IsStreamSupported() {
		switch mode {
		case window.ExplainMode:
			templateID := config.Data.ExplainTemplates.DefaultTemplate
			if templateID == "" {
				for id := range config.Data.ExplainTemplates.Templates {
					templateID = id
					break
				}
			}
			if templateID != "" {
				a.Translator.Explain(ctx, queryText, templateID)
			}
		default:
			a.Translator.Translate(ctx, queryText, fl, tl)
		}
	} else {
		result := a.Translator.Translate(ctx, queryText, fl, tl)
		a.EventBus.EmitResult(result)
		a.EventBus.Emit(event.Explains, "")
	}
}

// HandleScreenshotEvent 处理截图快捷键事件。
func (a *Application) HandleScreenshotEvent() {
	base64Image := screenshot.ScreenshotFullScreen()
	a.EventBus.EmitScreenshotBase64(base64Image)
}
