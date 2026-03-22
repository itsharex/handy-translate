package main

import (
	"embed"
	_ "embed"
	"log"
	"log/slog"
	"time"

	"handy-translate/config"
	"handy-translate/history"
	internalApp "handy-translate/internal/app"
	"handy-translate/internal/event"
	"handy-translate/internal/service"
	"handy-translate/internal/translate"
	"handy-translate/internal/window"
	screenshotWin "handy-translate/window/screenshot"
	"handy-translate/window/toolbar"
	translateWin "handy-translate/window/translate"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed frontend/public/appicon.png
var iconlogo []byte

var projectName = "handy-translate"

func main() {
	// ──────────────────────────────────────────
	// 1. 创建 Wails 应用（Binding 稍后注入）
	// ──────────────────────────────────────────
	// Binding 需要在 Application 组装完成后填充
	var binding internalApp.Binding

	wailsApp := application.New(application.Options{
		Name: projectName,
		Services: []application.Service{
			application.NewService(&binding),
		},
		Icon: iconlogo,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.wails.handy-translate",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				log.Printf("Second instance launched with args: %v", data.Args)
			},
			AdditionalData: map[string]string{
				"launchtime": time.Now().String(),
			},
		},
	})

	// ──────────────────────────────────────────
	// 2. 创建窗口
	// ──────────────────────────────────────────
	toolbar.NewWindow(wailsApp)
	translateWin.NewWindow(wailsApp)
	screenshotWin.NewWindow(wailsApp)

	// ──────────────────────────────────────────
	// 3. 初始化配置
	// ──────────────────────────────────────────
	config.Init(projectName)

	// ──────────────────────────────────────────
	// 4. 组装依赖（依赖注入）
	// ──────────────────────────────────────────

	// 事件总线（观察者模式）
	eventBus := event.NewBus(wailsApp)

	// 翻译提供者注册表（注册表模式 + 策略模式）
	providerRegistry := translate.NewRegistry()
	translate.RegisterAll(providerRegistry)

	// 历史记录服务
	historySvc := history.NewHistoryService()

	// 翻译业务门面（门面模式）
	wordCache := service.NewWordCache("data/word_cache")
	translator := service.NewTranslator(providerRegistry, &config.Data, eventBus, historySvc, wordCache)

	// 窗口管理器
	windowMgr := window.NewManager(wailsApp)
	windowMgr.Toolbar = toolbar.Window
	windowMgr.Translate = translateWin.Window
	windowMgr.Screenshot = screenshotWin.Window

	// OCR 服务
	ocrSvc := service.NewOCRService(".\\RapidOCR-json.exe")

	// 应用核心（依赖注入容器）
	app := internalApp.NewApplication(
		wailsApp,
		translator,
		windowMgr,
		eventBus,
		historySvc,
		ocrSvc,
	)

	// 将 Application 注入到 Binding（适配器模式）
	binding = *internalApp.NewBinding(app)

	// ──────────────────────────────────────────
	// 5. 注册事件（观察者模式）
	// ──────────────────────────────────────────
	app.RegisterEvents()

	// 从配置读取工具栏模式
	if config.Data.ToolbarMode != "" {
		app.SetToolbarMode(config.Data.ToolbarMode)
		slog.Info("从配置读取工具栏模式", slog.String("mode", config.Data.ToolbarMode))
	} else {
		app.SetToolbarMode("translate")
	}

	// 从配置读取工具栏固定状态
	if config.Data.ToolbarPinned {
		windowMgr.SetPinned(true)
		slog.Info("从配置读取工具栏固定状态", slog.Bool("pinned", true))
	}

	// ──────────────────────────────────────────
	// 6. 系统托盘
	// ──────────────────────────────────────────
	systemTray := wailsApp.SystemTray.New()
	myMenu := wailsApp.Menu.New()

	myMenu.Add("截图").OnClick(func(ctx *application.Context) {
		app.HandleScreenshotEvent()
	})

	myMenu.Add("退出").OnClick(func(ctx *application.Context) {
		wailsApp.Quit()
	})

	systemTray.SetMenu(myMenu)
	systemTray.SetIcon(iconlogo)

	systemTray.OnClick(func() {
		toolbar.Window.Show()
	})

	// ──────────────────────────────────────────
	// 7. 启动 Hook 监听 + 运行应用
	// ──────────────────────────────────────────
	go app.ProcessHook()

	err := wailsApp.Run()
	if err != nil {
		panic(err)
	}
}
