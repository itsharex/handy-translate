package windows

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// 因为windows下的robotgo鼠标获取文本内容有些瑕疵，故这里用windows原生api增强
const (
	WH_MOUSE_LL    = 14
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MBUTTONDOWN = 0x0207
	WM_MBUTTONUP   = 0x0208
	MK_LBUTTON     = 0x0001
	MK_RBUTTON     = 0x0002
	MK_MBUTTON     = 0x0010
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	setWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	callNextHookEx      = user32.NewProc("CallNextHookEx")
	unhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	getMessageW         = user32.NewProc("GetMessageW")
	keybdEventProc      = user32.NewProc("keybd_event") // 键盘事件函数

	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	KEYEVENTF_KEYUP = 0x0002
	VK_CONTROL      = 0x11
	VK_C            = 0x43
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101

	VK_CTRL  = 0xa2
	VK_SHIFT = 0xa0
	VK_F     = 0x46
)

type KBDLLHOOKSTRUCT struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type POINT struct {
	X, Y int32
}

var (
	hMouseHook    uintptr
	hKeyboardHook uintptr
)

var PressLock sync.RWMutex

// HookChan channle
var HookChan = make(chan string, 10)

var state = 0 // 0=等待Ctrl，1=等待Shift，2=等待F

// LowLevelMouseProc 代用windows api 才能做到选中文字，鼠标事件触发前执行模拟ctrl + c 操作
func LowLevelMouseProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	runtime.LockOSThread()
	r1, _, _ := callNextHookEx.Call(uintptr(nCode), wParam, lParam)
	if nCode >= 0 {
		// msllHookStruct := (*MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		switch wParam {
		case WM_MOUSEMOVE:
			// fmt.Println("鼠标移动：", msllHookStruct.Pt.X, msllHookStruct.Pt.Y)
		case WM_LBUTTONDOWN:
			// fmt.Println("左键按下")
		case WM_LBUTTONUP:
			// fmt.Println("左键释放")
		case WM_RBUTTONDOWN:
			// fmt.Println("右键按下")
		case WM_RBUTTONUP:
			// fmt.Println("右键释放")
		case WM_MBUTTONDOWN:
			PressCtrlC()
			time.Sleep(time.Millisecond * 100)
			HookChan <- "mouse"
		case WM_MBUTTONUP:
		}
	}
	return r1
}

func WindowsHook() {
	// 启动键盘钩子
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hMod, _, _ := getModuleHandleW.Call(0)

		var err error
		hKeyboardHook, _, err = setWindowsHookExW.Call(
			uintptr(WH_KEYBOARD_LL),
			syscall.NewCallback(onKeyboard),
			hMod,
			0,
		)
		if hKeyboardHook == 0 {
			fmt.Println("❌ 键盘钩子安装失败:", err)
			return
		}
		defer unhookWindowsHookEx.Call(hKeyboardHook)

		fmt.Println("✅ 钩子已安装，请依次按 Ctrl → Shift → F")

		var msg struct{}
		// 必须在同一个线程中处理消息循环
		for {
			ret, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if ret == 0 {
				break
			}
		}
	}()

	// 启动鼠标钩子（主线程中）
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hMouseHook, _, _ = setWindowsHookExW.Call(uintptr(WH_MOUSE_LL), syscall.NewCallback(LowLevelMouseProc), 0, 0)
	if hMouseHook != 0 {
		defer unhookWindowsHookEx.Call(hMouseHook)
	}

	var msg struct{}
	// 监听消息
	for {
		ret, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
	}
}

// 获取鼠标位置
func GetMousePosition() (x, y int32, err error) {
	var point struct{ X, Y int32 }
	// 调用 Windows API：GetCursorPos
	ret, _, err := syscall.NewLazyDLL("user32.dll").NewProc("GetCursorPos").Call(
		uintptr(unsafe.Pointer(&point)),
	)
	if ret == 0 {
		return 0, 0, err
	}
	return point.X, point.Y, nil
}

func onKeyboard(nCode int, wParam, lParam uintptr) uintptr {
	if nCode >= 0 {
		kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if wParam == WM_KEYDOWN {
			handleSequence(kbd.vkCode)
		}
	}
	ret, _, _ := callNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func handleSequence(key uint32) {
	switch state {
	case 0:
		if key == VK_CTRL {
			state = 1
		}
	case 1:
		if key == VK_SHIFT {
			state = 2
		} else if key != VK_CTRL {
			state = 0
		}
	case 2:
		if key == VK_F {
			fmt.Println("🎉 顺序匹配成功：Ctrl → Shift → F")
			HookChan <- "screenshot"
		}
		state = 0
	}
}
