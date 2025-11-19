package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"handy-translate/config"
	"handy-translate/history"
	"handy-translate/logger"
	"handy-translate/os_api/windows"
	"handy-translate/translate_service"
	"handy-translate/utils"
	"handy-translate/window/screenshot"
	"handy-translate/window/toolbar"
	"handy-translate/window/translate"

	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var toolbarIsShowing bool = false // 标记工具栏是否已经显示

// 和js绑定的go方法集合
type AppInterface interface {
	Show(windowName string)
	Hide(windowName string)
	ToolBarShow(height float64)
	SetToolBarPinned(pinned bool)
	GetToolBarPinned() bool
}

// App is a service
type App struct{}

// truncateText 截断文本用于日志显示，超长文本显示前N个字符后跟...
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	// 如果是中文字符，需要特殊处理以避免切割中文字符
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// GetToolbarMode 获取工具栏模式
func GetToolbarMode() string {
	return config.Data.ToolbarMode
}

// SetToolbarMode 设置工具栏模式
func SetToolbarMode(mode string) {
	config.Data.ToolbarMode = mode
	if err := config.Save(); err != nil {
		slog.Error("Failed to save config", slog.String("error", err.Error()))
	}
	slog.Info("SetToolbarMode", slog.String("mode", mode))
}

// GetToolbarMode 获取工具栏模式（供前端调用）
func (a *App) GetToolbarMode() string {
	return GetToolbarMode()
}

// MyFetch URl
func (a *App) MyFetch(URL string, content map[string]interface{}) interface{} {
	return utils.MyFetch(URL, content)
}

// Translate 翻译逻辑
func (a *App) Translate(queryText, fromLang, toLang string) string {
	app.Logger.Info("🔍 Translate 开始",
		slog.String("text_preview", truncateText(queryText, 50)),
		slog.String("from", fromLang),
		slog.String("to", toLang))

	res := processTranslate(queryText)
	return res
}

// TranslateMeanings 翻译逻辑
func (a *App) TranslateMeanings(queryText, fromLang, toLang string) string {
	app.Logger.Info("🔍 TranslateMeanings 开始",
		slog.String("text_preview", truncateText(queryText, 50)),
		slog.String("from", fromLang),
		slog.String("to", toLang))

	translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

	// 检查是否支持流式输出
	if streamTranslate, ok := translateWay.(translate_service.StreamTranslate); ok {
		// 支持流式输出
		logger.LogTranslateStart(translateWay.GetName(), fromLang, toLang, len(queryText))

		var streamResult strings.Builder
		chunkCount := 0
		startTime := time.Now()

		err := streamTranslate.PostQueryStream(queryText, fromLang, toLang, func(chunk string) {
			streamResult.WriteString(chunk)
			chunkCount++
			totalLen := len(streamResult.String())

			// 每次收到数据块时发送事件到前端
			logger.LogChunkReceived(chunkCount, len(chunk), totalLen)
			app.Event.Emit("result_meanings_stream", chunk)
		})
		if err != nil {
			elapsed := time.Since(startTime)
			logger.LogTranslateError(translateWay.GetName(), err, chunkCount, elapsed)
			app.Event.Emit("result_stream_error", err.Error())
			return ""
		}

		// 发送完成事件
		app.Event.Emit("result_stream_done", "done")

		resultStr := streamResult.String()
		elapsed := time.Since(startTime)

		logger.LogTranslateSuccess(translateWay.GetName(), len(queryText), len(resultStr), chunkCount, elapsed)

		return resultStr
	}

	// 不支持流式，使用普通翻译
	logger.LogNormalTranslateStart(translateWay.GetName(), fromLang, toLang)

	result, err := translateWay.PostQuery(queryText, fromLang, toLang)
	if err != nil {
		logger.LogNormalTranslateError(translateWay.GetName(), err)
	}

	logger.LogNormalTranslateSuccess(translateWay.GetName(), len(result))

	translateRes := strings.Join(result, "\n")

	// 保存翻译历史记录
	if config.Data.History.Enabled {
		go history.GlobalHistoryService.SaveTranslateRecord(queryText, translateRes, fromLang, toLang)
	}

	return translateRes
}

// TranslateStream 流式翻译逻辑（仅支持 DeepSeek）
func (a *App) TranslateStream(queryText, fromLang, toLang string) {
	app.Logger.Info("🔍 TranslateStream 开始",
		slog.String("text_preview", truncateText(queryText, 50)),
		slog.String("from", fromLang),
		slog.String("to", toLang))

	translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

	// 检查是否支持流式输出
	if streamTranslate, ok := translateWay.(translate_service.StreamTranslate); ok {
		// 支持流式输出
		logger.LogTranslateStart(translateWay.GetName(), fromLang, toLang, len(queryText))

		chunkCount := 0
		startTime := time.Now()

		err := streamTranslate.PostQueryStream(queryText, fromLang, toLang, func(chunk string) {
			chunkCount++

			// 每次收到数据块时发送事件到前端
			slog.Debug("📥 接收流式数据块",
				slog.Int("chunk_count", chunkCount),
				slog.Int("chunk_size", len(chunk)))
			app.Event.Emit("result_stream", chunk)
		})

		if err != nil {
			elapsed := time.Since(startTime)
			logger.LogTranslateError(translateWay.GetName(), err, chunkCount, elapsed)
			// 发送错误事件
			app.Event.Emit("result_stream_error", err.Error())
		} else {
			// 发送完成事件
			elapsed := time.Since(startTime)
			logger.LogTranslateSuccess(translateWay.GetName(), len(queryText), 0, chunkCount, elapsed)
			app.Event.Emit("result_stream_done", "done")
		}
	} else {
		// 不支持流式输出，使用普通翻译
		logger.LogStreamNotSupported(translateWay.GetName())
		res := processTranslate(queryText)
		app.Event.Emit("result", res)
	}
}

// GetTranslateMap 获取所有翻译配置
func (a *App) GetTranslateMap() string {
	translateList := config.Data.Translate
	bTranslate, err := json.Marshal(translateList)
	if err != nil {
		logrus.WithError(err).Error("Marshal")
	}
	return string(bTranslate)
}

// SetTranslateWay 设置当前翻译服务
func (a *App) SetTranslateWay(translateWay string) {
	config.Data.TranslateWay = translateWay
	translate_service.SetQueryText("")
	if err := config.Save(); err != nil {
		slog.Error("Failed to save config", slog.String("error", err.Error()))
	}
	slog.Info("SetTranslateList", slog.Any("config.Data.Translate", config.Data.Translate))
}

// GetTranslateWay 获取当前翻译的服务
func (a *App) GetTranslateWay() string {
	return config.Data.TranslateWay
}

// GetExplainTemplates 获取所有解释模板
func (a *App) GetExplainTemplates() string {
	templates := make(map[string]map[string]interface{})

	// 如果配置为空，返回空结果
	if len(config.Data.ExplainTemplates.Templates) == 0 {
		return "{}"
	}

	// 构建返回数据
	for id, template := range config.Data.ExplainTemplates.Templates {
		templates[id] = map[string]interface{}{
			"id":          id,
			"name":        template.Name,
			"description": template.Description,
		}
	}

	result := map[string]interface{}{
		"default_template": config.Data.ExplainTemplates.DefaultTemplate,
		"templates":        templates,
	}

	b, err := json.Marshal(result)
	if err != nil {
		logrus.WithError(err).Error("Marshal ExplainTemplates")
		return "{}"
	}
	return string(b)
}

// SetDefaultExplainTemplate 设置默认解释模板
func (a *App) SetDefaultExplainTemplate(templateID string) {
	config.Data.ExplainTemplates.DefaultTemplate = templateID
	if err := config.Save(); err != nil {
		slog.Error("Failed to save config", slog.String("error", err.Error()))
	}
	slog.Info("SetDefaultExplainTemplate", slog.String("templateID", templateID))
}

// Show 通过名字控制窗口事件
func (a *App) Show(windowName string) {
	var win *application.WebviewWindow
	switch windowName {
	case screenshot.WindowName:
		win = screenshot.Window
	case translate.WindowName:
		win = translate.Window
	case toolbar.WindowName:
		win = toolbar.Window
	}

	// 检查窗口是否存在
	if win == nil {
		app.Logger.Error("Show: 窗口不存在", slog.String("windowName", windowName))
		return
	}

	win.Center()
	win.Show()
}

// Hide 通过名字控制窗口事件
func (a *App) Hide(windowName string) {
	var win *application.WebviewWindow
	switch windowName {
	case screenshot.WindowName:
		win = screenshot.Window
	case translate.WindowName:
		win = translate.Window
	case toolbar.WindowName:
		win = toolbar.Window
	}

	// 检查窗口是否存在
	if win == nil {
		app.Logger.Error("Hide: 窗口不存在", slog.String("windowName", windowName))
		return
	}

	win.Hide()
}

// ToolBarShow 显示工具弹窗，控制大小，布局, 前端调用，传递文本高度
func (a *App) ToolBarShow(height float64) {
	// 40 + 55 窗口空白区域+翻译的图标区域
	height = height + 35 + 54
	app.Logger.Info("ToolBarShow", slog.Float64("height", height), slog.Bool("isShowing", toolbarIsShowing))

	h := min(int(height), toolbar.QueryResultHeight+500)

	if h == 0 {
		h = toolbar.QueryResultHeight
	}
	toolbar.QueryResultHeight = h
	processToolbarShow()
}

// SetToolBarPinned 设置工具栏固定状态
func (a *App) SetToolBarPinned(pinned bool) {
	toolbar.IsPinned = pinned
	app.Logger.Info("SetToolBarPinned", slog.Bool("pinned", pinned))
}

// GetToolBarPinned 获取工具栏固定状态
func (a *App) GetToolBarPinned() bool {
	app.Logger.Info("GetToolBarPinned", slog.Bool("pinned", toolbar.IsPinned))
	return toolbar.IsPinned
}

func processToolbarShow() {
	height := toolbar.QueryResultHeight
	w := toolbar.Window
	w.SetSize(w.Width(), height)

	// 如果窗口已经显示，只调整大小，不改变位置
	if toolbarIsShowing {
		// slog.Info("工具栏已显示，仅调整大小", slog.Int("height", height))
		// 窗口已经显示，不需要重新定位
		return
	}

	xval := 0
	yval := 0

	if runtime.GOOS == "windows" {
		pos := windows.GetCursorPos()
		xval, yval = int(pos.X), int(pos.Y) // 处理获取坐标不正确，采用windows原生api
	} else {
		slog.Error("仅支持Windows平台", slog.String("platform", runtime.GOOS))
		return
	}

	sc, _ := w.GetScreen()
	// 计算屏幕任务多出的高度，防止弹出框超出屏幕外面
	c := int(float64(sc.Size.Height) * 0.1)

	// 计算左边对应的窗体高度是否超出屏幕外，超出则需要重新计算y轴坐标，防止弹出框超出屏幕外面
	if yval+height+c >= sc.Size.Height {
		gap := yval + height + c - sc.Size.Height
		slog.Info("窗口初始定位（超出屏幕）", slog.Int("gap", gap), slog.Int("x", xval+10), slog.Int("y", yval-gap))
		w.SetPosition(xval+10, yval-gap)
	} else {
		slog.Info("窗口初始定位（正常）", slog.Int("x", xval+10), slog.Int("y", yval+10))
		w.SetPosition(xval+10, yval+10)
	}

	// 显示窗口
	if runtime.GOOS == "windows" {
		// Windows 平台：尝试使用原生 API 显示窗口（更可靠）
		win := windows.FindWindow(toolbar.WindowName)
		if win != nil {
			slog.Info("使用 Windows 原生 API 显示工具栏")
			win.ShowForWindows()
		} else {
			// 找不到窗口时使用 Wails API 作为后备
			slog.Warn("无法通过 FindWindow 找到工具栏，使用 Wails Show() 方法")
			toolbar.Window.Show()
		}
	} else {
		// 非 Windows 平台：使用 Wails API
		toolbar.Window.Show()
	}

	// 标记窗口已显示
	toolbarIsShowing = true
}

// ResetToolbarState 重置工具栏状态（在窗口隐藏时调用）
func ResetToolbarState() {
	toolbarIsShowing = false
	slog.Info("工具栏状态已重置")
}

// CaptureSelectedScreen 截取选中的区域
func (a *App) CaptureSelectedScreen(startX, startY, width, height float64) {
	croppedImg := screenshot.CaptureSelectedScreen(int(startX), int(startY), int(width), int(height))
	if croppedImg == nil {
		return
	}

	var buf bytes.Buffer
	err := png.Encode(&buf, croppedImg)
	if err != nil {
		slog.Error("png.Encode", slog.Any("err", err))
		return
	}

	filename := "screenshot.png" // 保存的文件名
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())

	err = saveBase64Image(base64String, filename)
	if err != nil {
		slog.Error("saveBase64Image", slog.Any("err", err))
	}

	// OCR解析文本
	queryText := ExecOCR(".\\RapidOCR-json.exe", filename)

	// 重置工具栏状态，准备新的翻译
	ResetToolbarState()

	// 检查是否使用了流式翻译
	translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

	// 无论是流式还是普通翻译，都先发送 query 事件让前端准备
	sendQueryText(queryText)

	if _, ok := translateWay.(translate_service.StreamTranslate); ok {
		// 流式翻译：开始流式翻译（会发送 result_stream 事件）
		translateRes := processTranslate(queryText)
		slog.Info("截图OCR流式翻译完成，结果长度，模式", slog.Int("len", len(translateRes)), slog.String("mode", GetToolbarMode()))
	} else {
		// 普通翻译：翻译后发送完整结果
		translateRes := processTranslate(queryText)
		sendResult(translateRes, "")
	}
}

// 翻译处理
func processTranslate(queryText string) string {
	translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

	// 检查是否支持流式输出
	if streamTranslate, ok := translateWay.(translate_service.StreamTranslate); ok {
		// 支持流式输出
		logger.LogTranslateStart(translateWay.GetName(), fromLang, toLang, len(queryText))

		var streamResult strings.Builder
		chunkCount := 0
		startTime := time.Now()

		err := streamTranslate.PostQueryStream(queryText, fromLang, toLang, func(chunk string) {
			streamResult.WriteString(chunk)
			chunkCount++
			totalLen := len(streamResult.String())

			// 每次收到数据块时发送事件到前端
			logger.LogChunkReceived(chunkCount, len(chunk), totalLen)
			app.Event.Emit("result_stream", chunk)
		})
		if err != nil {
			elapsed := time.Since(startTime)
			logger.LogTranslateError(translateWay.GetName(), err, chunkCount, elapsed)
			app.Event.Emit("result_stream_error", err.Error())
			return ""
		}

		// 发送完成事件
		app.Event.Emit("result_stream_done", "done")

		resultStr := streamResult.String()
		elapsed := time.Since(startTime)

		logger.LogTranslateSuccess(translateWay.GetName(), len(queryText), len(resultStr), chunkCount, elapsed)

		// 保存翻译历史记录
		if config.Data.History.Enabled {
			go history.GlobalHistoryService.SaveTranslateRecord(queryText, resultStr, fromLang, toLang)
		}

		return resultStr
	}

	// 不支持流式，使用普通翻译
	logger.LogNormalTranslateStart(translateWay.GetName(), fromLang, toLang)

	result, err := translateWay.PostQuery(queryText, fromLang, toLang)
	if err != nil {
		logger.LogNormalTranslateError(translateWay.GetName(), err)
	}

	logger.LogNormalTranslateSuccess(translateWay.GetName(), len(result))

	translateRes := strings.Join(result, "\n")

	// 保存翻译历史记录
	if config.Data.History.Enabled {
		go history.GlobalHistoryService.SaveTranslateRecord(queryText, translateRes, fromLang, toLang)
	}

	return translateRes
} // 解释处理（支持模板选择）
func processExplain(queryText, templateID string) string {
	translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

	// 检查是否支持流式输出
	if streamTranslate, ok := translateWay.(translate_service.StreamTranslate); ok {
		// 支持流式输出
		logger.LogExplainStart(translateWay.GetName(), templateID, len(queryText))

		var streamResult strings.Builder
		chunkCount := 0
		startTime := time.Now()

		err := streamTranslate.PostExplainStream(queryText, templateID, func(chunk string) {
			streamResult.WriteString(chunk)
			chunkCount++
			totalLen := len(streamResult.String())

			// 每次收到数据块时发送事件到前端
			logger.LogChunkReceived(chunkCount, len(chunk), totalLen)
			app.Event.Emit("result_stream", chunk)
			time.Sleep(time.Millisecond * 20) // 控制发送速度，防止前端卡顿
		})
		if err != nil {
			elapsed := time.Since(startTime)
			logger.LogExplainError(translateWay.GetName(), templateID, err, chunkCount, elapsed)
			app.Event.Emit("result_stream_error", err.Error())
			return ""
		}

		// 发送完成事件
		app.Event.Emit("result_stream_done", "done")

		resultStr := streamResult.String()
		elapsed := time.Since(startTime)

		logger.LogExplainSuccess(translateWay.GetName(), templateID, len(queryText), len(resultStr), chunkCount, elapsed)

		// 保存解释历史记录
		if config.Data.History.Enabled {
			go history.GlobalHistoryService.SaveExplainRecord(queryText, resultStr, templateID)
		}

		return resultStr
	}

	// 不支持流式
	logger.LogStreamExplainNotSupported(translateWay.GetName(), templateID)
	return ""
}

func sendQueryText(queryText string) {
	app.Event.Emit("query", queryText)
}

func sendResult(result, explains string) {
	app.Event.Emit("result", result)
	app.Event.Emit("explains", explains)
}

// 监听处理鼠标事件
func processHook() {
	// TODO 工厂设计模式
	if runtime.GOOS == "windows" {
		go windows.WindowsHook()
	}

	for msg := range windows.HookChan {
		switch msg {
		case "mouse":
			result, ok := app.Clipboard.Text()
			if !ok {
				app.Logger.Error("Failed to get clipboard text")
			}

			queryText := result

			app.Logger.Info("processHook GetQueryText",
				slog.String("queryText", queryText),
				slog.String("fromLang", fromLang),
				slog.String("toLang", toLang))

			// 当工具栏已固定时，不重置或重新定位窗口，直接在现有窗口渲染数据
			if !toolbar.IsPinned {
				ResetToolbarState()
				processToolbarShow()
			}

			if queryText != translate_service.GetQueryText() && queryText != "" {
				translate_service.SetQueryText(queryText)

				// 检查是否使用了流式翻译
				translateWay := translate_service.GetTranslateWay(config.Data.TranslateWay)

				// 无论是流式还是普通翻译，都先发送 query 事件让前端准备
				sendQueryText(queryText)

				// 根据工具栏模式选择翻译或解释
				mode := GetToolbarMode()
				switch mode {
				case toolbar.ExplainMode:
					// 解释模式：后端自动处理，使用默认模板
					templateID := config.Data.ExplainTemplates.DefaultTemplate
					if templateID == "" {
						// 如果没有默认模板，尝试使用第一个模板
						for id := range config.Data.ExplainTemplates.Templates {
							templateID = id
							break
						}
					}
					if templateID == "" {
						slog.Error("解释模式但未找到模板")
						app.Event.Emit("result_stream_error", "未找到解释模板")
					} else {
						slog.Info("解释模式，自动处理", slog.String("templateID", templateID))
						// 自动调用解释处理
						if _, ok := translateWay.(translate_service.StreamTranslate); ok {
							// 流式解释：开始流式解释（会发送 result_stream 事件）
							explainRes := processExplain(queryText, templateID)
							slog.Info("流式解释完成，结果长度", slog.Int("len", len(explainRes)))
						}
					}
				default:
					// 翻译模式（默认）
					if _, ok := translateWay.(translate_service.StreamTranslate); ok {
						// 流式翻译：开始流式翻译（会发送 result_stream 事件）
						translateRes := processTranslate(queryText)
						slog.Info("流式翻译完成，结果长度", slog.Int("len", len(translateRes)))
					}
				}
			}
		case "screenshot":
			screenshot.ScreenshotFullScreen()
			base64Image := screenshot.ScreenshotFullScreen()
			app.Event.Emit("screenshotBase64", base64Image)
		default:
			app.Logger.Error("processHook", slog.String("msg", msg))
		}
	}
}
