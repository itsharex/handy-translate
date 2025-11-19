package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"time"

	"handy-translate/config"
	"handy-translate/history"
	"handy-translate/window/screenshot"
	"handy-translate/window/toolbar"
	"handy-translate/window/translate"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed frontend/public/appicon.png
var iconlogo []byte

var app *application.App

var fromLang, toLang = "auto", "zh"

var projectName = "handy-translate"

func main() {
	app = application.New(application.Options{
		Name: projectName,
		Services: []application.Service{
			application.NewService(&App{}),
		},
		Icon: iconlogo,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.wails.handy-translate",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				log.Printf("Second instance launched with args: %v", data.Args)
				log.Printf("Working directory: %s", data.WorkingDir)
				log.Printf("Additional data: %v", data.AdditionalData)
			},
			// Optional: Pass additional data to second instance
			AdditionalData: map[string]string{
				"launchtime": time.Now().String(),
			},
		},
	})

	toolbar.NewWindow(app)

	translate.NewWindow(app)

	screenshot.NewWindow(app)

	app.Event.On("translateLang", func(event *application.CustomEvent) {
		app.Logger.Info("translateType", slog.Any("event", event))

		if dataSlice, ok := event.Data.([]interface{}); ok {
			if len(dataSlice) >= 2 {
				fromLang = fmt.Sprintf("%v", dataSlice[0])
				toLang = fmt.Sprintf("%v", dataSlice[1])
				app.Logger.Info("translateLang",
					slog.String("fromLang", fromLang),
					slog.String("toLang", toLang))
			}
		}
	})

	app.Event.On("toolbarMode", func(event *application.CustomEvent) {
		app.Logger.Info("toolbarMode", slog.Any("event", event))
		if mode, ok := event.Data.(string); ok {
			SetToolbarMode(mode)
			app.Logger.Info("toolbarMode 已更新", slog.String("mode", mode))
			// 推送模式更新到前端
			app.Event.Emit("toolbarModeUpdated", mode)

			// 如果有当前的 queryText，自动重新处理
			currentQueryText := getCurrentQueryText()
			if currentQueryText != "" {
				app.Logger.Info("模式切换，重新处理当前查询",
					slog.String("queryText", currentQueryText),
					slog.String("mode", mode))

				// 发送 query 事件让前端准备
				app.Event.Emit("query", currentQueryText)

				// 根据新模式自动处理
				if mode == toolbar.ExplainMode {
					// 解释模式
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
						slog.Info("模式切换后自动解释", slog.String("templateID", templateID))
						explainRes := processExplain(currentQueryText, templateID)
						slog.Info("模式切换后流式解释完成", slog.Int("len", len(explainRes)))
					}
				} else {
					// 翻译模式
					translateRes := processTranslate(currentQueryText)
					slog.Info("模式切换后流式翻译完成", slog.Int("len", len(translateRes)))
				}
			}
		}
	})

	// 系统托盘
	systemTray := app.SystemTray.New()
	myMenu := app.Menu.New()

	myMenu.Add("翻译").OnClick(func(ctx *application.Context) {
		if translate.Window == nil {
			log.Printf("错误: translate.Window 为 nil")
			return
		}
		log.Printf("显示翻译窗口")
		// 使用 Center() 和 Show() 显示窗口
		translate.Window.Center()
		translate.Window.Show()
		// 确保窗口获得焦点
		translate.Window.Focus()
		log.Printf("翻译窗口已调用 Show() 和 Focus()")
	})

	myMenu.Add("截图").OnClick(func(ctx *application.Context) {
		screenshot.ScreenshotFullScreen()
		base64Image := screenshot.ScreenshotFullScreen()
		app.Event.Emit("screenshotBase64", base64Image)
	})

	myMenu.Add("退出").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	systemTray.SetMenu(myMenu)
	systemTray.SetIcon(iconlogo)

	systemTray.OnClick(func() {
		toolbar.Window.Show()
	})

	// 初始化文件和鼠标事件
	config.Init(projectName)

	// 从配置读取工具栏模式
	if config.Data.ToolbarMode != "" {
		SetToolbarMode(config.Data.ToolbarMode)
		app.Logger.Info("从配置读取工具栏模式", slog.String("mode", config.Data.ToolbarMode))
	} else {
		// 如果配置中没有模式，使用默认值并保存
		SetToolbarMode("translate")
		app.Logger.Info("使用默认工具栏模式", slog.String("mode", "translate"))
	}

	// 初始化历史记录服务
	history.GlobalHistoryService = history.NewHistoryService()

	go processHook()

	err := app.Run()
	if err != nil {
		// 报错退出程序
		panic(err)
	}
}
