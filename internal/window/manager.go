// Package window 提供统一的窗口管理（窗口管理器模式）。
package window

import (
	"log/slog"
	"runtime"
	"sync"

	"handy-translate/config"
	"handy-translate/internal/event"
	"handy-translate/os_api/windows"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ──────────────────────────────────────────────
// 窗口名称常量
// ──────────────────────────────────────────────

const (
	ToolbarWindowName    = "ToolBar"
	TranslateWindowName  = "Translate"
	ScreenshotWindowName = "Screenshot"
)

// ──────────────────────────────────────────────
// 模式常量
// ──────────────────────────────────────────────

const (
	ExplainMode   = "explain"
	TranslateMode = "translate"
)

// Manager 窗口管理器，统一管理所有窗口的显示/隐藏/定位。
type Manager struct {
	app        *application.App
	eventBus   *event.Bus
	Toolbar    *application.WebviewWindow
	Translate  *application.WebviewWindow
	Screenshot *application.WebviewWindow

	// 工具栏状态（通过 mutex 保护）
	mu                sync.RWMutex
	isPinned          bool
	QueryResultHeight int
	QueryResultWidth  int
	toolbarShowing    bool
}

// NewManager 创建窗口管理器。
func NewManager(app *application.App, eventBus *event.Bus) *Manager {
	return &Manager{
		app:               app,
		eventBus:          eventBus,
		QueryResultHeight: 110,
		QueryResultWidth:  450,
	}
}

// Show 通过窗口名称显示窗口。
func (m *Manager) Show(windowName string) {
	win := m.GetWindow(windowName)
	if win == nil {
		slog.Error("Show: 窗口不存在", slog.String("windowName", windowName))
		return
	}
	win.Center()
	win.Show()
}

// Hide 通过窗口名称隐藏窗口。
func (m *Manager) Hide(windowName string) {
	win := m.GetWindow(windowName)
	if win == nil {
		slog.Error("Hide: 窗口不存在", slog.String("windowName", windowName))
		return
	}
	win.Hide()
}

// GetWindow 通过名称获取窗口。
func (m *Manager) GetWindow(name string) *application.WebviewWindow {
	switch name {
	case ToolbarWindowName:
		return m.Toolbar
	case TranslateWindowName:
		return m.Translate
	case ScreenshotWindowName:
		return m.Screenshot
	default:
		return nil
	}
}

// GetPinned 获取工具栏固定状态（并发安全）。
func (m *Manager) GetPinned() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isPinned
}

// IsToolbarShowing 获取工具栏是否正在显示（并发安全）。
func (m *Manager) IsToolbarShowing() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.toolbarShowing
}

// SetToolbarShowing 设置工具栏显示状态（并发安全）。
func (m *Manager) SetToolbarShowing(showing bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolbarShowing = showing
}

// ResetToolbarState 重置工具栏状态（并发安全）。
func (m *Manager) ResetToolbarState() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolbarShowing = false
}

// SetPinned 设置工具栏固定状态（同时操作窗口置顶）。
func (m *Manager) SetPinned(pinned bool) {
	m.mu.Lock()
	m.isPinned = pinned
	m.mu.Unlock()

	config.Data.ToolbarPinned = pinned
	_ = config.Save()

	if m.Toolbar != nil {
		m.Toolbar.SetAlwaysOnTop(pinned)
		// 通过 EventBus 使用常量发射事件
		m.eventBus.EmitToolbarPinnedUpdated(pinned)
		slog.Info("工具栏置顶状态已更新", slog.Bool("pinned", pinned))
	}
}

// ShowToolbarAtCursor 在鼠标位置附近显示工具栏。
func (m *Manager) ShowToolbarAtCursor(height int) {
	w := m.Toolbar
	if w == nil {
		return
	}

	m.mu.Lock()
	h := min(height, m.QueryResultHeight+500)
	if h == 0 {
		h = m.QueryResultHeight
	}
	m.QueryResultHeight = h
	showing := m.toolbarShowing
	isPinned := m.isPinned
	m.mu.Unlock()

	w.SetSize(w.Width(), h)
	w.SetAlwaysOnTop(isPinned)

	// 如果窗口已经显示，只调整大小
	if showing {
		return
	}

	xval := 0
	yval := 0

	if runtime.GOOS == "windows" {
		pos := windows.GetCursorPos()
		xval, yval = int(pos.X), int(pos.Y)
	} else {
		slog.Error("仅支持Windows平台", slog.String("platform", runtime.GOOS))
		return
	}

	sc, _ := w.GetScreen()
	c := int(float64(sc.Size.Height) * 0.1)

	if yval+h+c >= sc.Size.Height {
		gap := yval + h + c - sc.Size.Height
		w.SetPosition(xval+10, yval-gap)
	} else {
		w.SetPosition(xval+10, yval+10)
	}

	// 显示窗口
	if runtime.GOOS == "windows" {
		win := windows.FindWindow(ToolbarWindowName)
		if win != nil {
			win.ShowForWindows()
		} else {
			m.Toolbar.Show()
		}
	} else {
		m.Toolbar.Show()
	}

	m.mu.Lock()
	m.toolbarShowing = true
	m.mu.Unlock()
}
