package win32

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	getACP   = kernel32.NewProc("GetACP")

	user32                   = syscall.NewLazyDLL("user32.dll")
	getForegroundWindow      = user32.NewProc("GetForegroundWindow")
	getWindowTextA           = user32.NewProc("GetWindowTextA")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	setWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	callNextHookEx           = user32.NewProc("CallNextHookEx")
	unhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
	getMessageW              = user32.NewProc("GetMessageW")
	toAsciiEx                = user32.NewProc("ToAsciiEx")
	getKeyState              = user32.NewProc("GetKeyState")
	getAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	sendInput                = user32.NewProc("SendInput")
	setCursorPos             = user32.NewProc("SetCursorPos")
	getSystemMetrics         = user32.NewProc("GetSystemMetrics")
	getDpiForSystem          = user32.NewProc("GetDpiForSystem")
	monitorFromPoint         = user32.NewProc("MonitorFromPoint")
	enumDisplayMonitors      = user32.NewProc("EnumDisplayMonitors")
	getMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	mouseEvent               = user32.NewProc("mouse_event")
	keybdEvent               = user32.NewProc("keybd_event")

	shcore           = syscall.NewLazyDLL("shcore.dll")
	getDpiForMonitor = shcore.NewProc("GetDpiForMonitor")

	psapi              = syscall.NewLazyDLL("psapi.dll")
	getModuleBaseNameA = psapi.NewProc("GetModuleBaseNameA")
)

func isErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() != "The operation completed successfully."
}
func IsKeyDown(vkCode int) bool {
	result, _, _ := getAsyncKeyState.Call(uintptr(vkCode))
	return (result & 0x8000) != 0
}

// IsCapsLockOn 检测 CapsLock 是否开启
func IsCapsLockOn() bool {
	result, _, _ := getKeyState.Call(uintptr(0x14)) // VK_CAPITAL = 0x14
	return (result & 0x0001) != 0
}

// 获取系统前台窗口句柄
func GetForegroundWindow() (HWND, error) {
	r0, _, err := getForegroundWindow.Call()
	if isErr(err) {
		return 0, err
	}
	return HWND(r0), nil
}

// 获取窗口标题
func GetWindowTextA(hWnd HWND) (string, error) {
	title := [1024]byte{}
	length, _, err := getWindowTextA.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&title[0])), 1024)
	if isErr(err) {
		return "", err
	}
	return string(title[:length]), nil
}

// 获取线程号和进程号
func GetWindowThreadProcessId(hWnd HWND) (lpdwProcessId uint32, threadId uint32, err error) {
	var tId uintptr
	tId, _, err = getWindowThreadProcessId.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&lpdwProcessId)))
	threadId = uint32(tId)
	if !isErr(err) {
		err = nil
	}
	return
}

// 获取系统编码
func GetACP() (uint, error) {
	acp, _, err := getACP.Call()
	if isErr(err) {
		return 0, err
	}
	return uint(acp), nil
}

// 打开|获取 进程句柄
func OpenProcess(da uint32, inheritHandle bool, pid uint32) (handel HANDLE, err error) {
	h, err := syscall.OpenProcess(da, inheritHandle, pid)
	handel = HANDLE(h)
	if !isErr(err) {
		err = nil
	}
	return
}
func GetModuleBaseNameA(handel HANDLE) (string, error) {
	buf := [1024]byte{}
	length, _, err := getModuleBaseNameA.Call(uintptr(handel), 0,
		uintptr(unsafe.Pointer(&buf)), 1024)
	if isErr(err) {
		return "", err
	}
	return string(buf[:length]), nil
}

func CloseHandel(handel HANDLE) error {
	return syscall.CloseHandle(syscall.Handle(handel))
}

// set hook
func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod HINSTANCE, dwThreadId DWORD) (HHOOK, error) {
	ret, _, err := setWindowsHookExW.Call(
		uintptr(idHook),
		uintptr(syscall.NewCallback(lpfn)),
		uintptr(hMod),
		uintptr(dwThreadId),
	)
	if isErr(err) {
		return 0, err
	}
	return HHOOK(ret), nil
}

// call next hook
//
// 对于某些类型的HOOK，系统将向该类的所有HOOK函数发送消息，这时， HOOK函数中的CallNextHookEx语句将被忽略
func CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) (LRESULT, error) {
	ret, _, err := callNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	if !isErr(err) {
		err = nil
	}
	return LRESULT(ret), err
}

// 卸载hook
func UnhookWindowsHookEx(hhk HHOOK) (bool, error) {
	ret, _, err := unhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	if isErr(err) {
		return false, err
	}
	return ret != 0, nil
}

// 获取消息
func GetMessage(msg *MSG, hWnd HWND, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := getMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hWnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}

/*// 将指定的虚拟键盘和键盘状态翻译为相应的字符
func ToAsciiEx(uVirtkey uint, uScanCode uint, lpKeySate *byte, uFlags uint, dwhkl HKL) (string, error) {
	char := [256]byte{}
	len, _, err := toAsciiEx.Call(
		uintptr(uVirtkey),
		uintptr(uScanCode),
		uintptr(unsafe.Pointer(lpKeySate)),
		uintptr(unsafe.Pointer(&char)),
		uintptr(uFlags),
		uintptr(dwhkl),
	)
	if isErr(err) {
		return "", err
	}
	return string(char[:len]), nil
}*/

func GetKeyState(vkCode uint32) (int8, error) {
	res, _, err := getKeyState.Call(uintptr(vkCode))
	if isErr(err) {
		return 0, err
	}
	return int8(res), err
}

func HideConsole() {
	ShowConsoleAsync(SW_HIDE)
}
func ShowConsoleAsync(commandShow uintptr) {
	console := GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := GetWindowThreadProcessId_test(console)
		if GetCurrentProcessId() == consoleProcID {
			ShowWindowAsync(console, commandShow)
		}
	}
}
func GetConsoleWindow() uintptr {
	ret, _, _ := getConsoleWindow.Call()
	return ret
}
func GetWindowThreadProcessId_test(hwnd uintptr) (uintptr, uint32) {
	var processId uint32
	ret, _, _ := getWindowThreadProcessId.Call(
		hwnd,
		uintptr(unsafe.Pointer(&processId)),
	)
	return ret, processId
}
func GetCurrentProcessId() uint32 {
	id, _, _ := getCurrentProcessId.Call()
	return uint32(id)
}
func ShowWindowAsync(window, commandShow uintptr) bool {
	ret, _, _ := showWindowAsync.Call(window, commandShow)
	return ret != 0
}

// GetSystemMetrics 获取系统指标（如屏幕尺寸）
func GetSystemMetrics(nIndex int) int {
	ret, _, _ := getSystemMetrics.Call(uintptr(nIndex))
	// GetSystemMetrics 返回的是有符号整数
	// 对于虚拟屏幕坐标（SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN），可能是负数
	// 需要将 uintptr 正确转换为有符号整数
	return int(int32(ret))
}

// SetCursorPos 设置鼠标位置
func SetCursorPos(x, y int) error {
	ret, _, err := setCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return err
	}
	return nil
}

// SendInput 发送输入事件
func SendInput(inputs []INPUT) uint32 {
	if len(inputs) == 0 {
		return 0
	}
	ret, _, _ := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	return uint32(ret)
}

// MouseClick 鼠标点击（使用mouse_event API，兼容性更好）
func MouseClick(button string, double bool) {
	var downFlag, upFlag uint32
	switch button {
	case "left":
		downFlag = MOUSEEVENTF_LEFTDOWN
		upFlag = MOUSEEVENTF_LEFTUP
	case "right":
		downFlag = MOUSEEVENTF_RIGHTDOWN
		upFlag = MOUSEEVENTF_RIGHTUP
	case "middle":
		downFlag = MOUSEEVENTF_MIDDLEDOWN
		upFlag = MOUSEEVENTF_MIDDLEUP
	default:
		return
	}

	// 使用mouse_event API（较老但兼容性更好，不易被安全软件拦截）
	mouseEvent.Call(
		uintptr(downFlag),
		0, 0, 0, 0,
	)
	mouseEvent.Call(
		uintptr(upFlag),
		0, 0, 0, 0,
	)

	if double {
		mouseEvent.Call(
			uintptr(downFlag),
			0, 0, 0, 0,
		)
		mouseEvent.Call(
			uintptr(upFlag),
			0, 0, 0, 0,
		)
	}
}

// MouseToggle 鼠标按钮切换（使用mouse_event API）
func MouseToggle(button, state string) {
	var flag uint32
	switch button {
	case "left":
		if state == "down" {
			flag = MOUSEEVENTF_LEFTDOWN
		} else {
			flag = MOUSEEVENTF_LEFTUP
		}
	case "right":
		if state == "down" {
			flag = MOUSEEVENTF_RIGHTDOWN
		} else {
			flag = MOUSEEVENTF_RIGHTUP
		}
	case "middle":
		if state == "down" {
			flag = MOUSEEVENTF_MIDDLEDOWN
		} else {
			flag = MOUSEEVENTF_MIDDLEUP
		}
	default:
		return
	}

	mouseEvent.Call(
		uintptr(flag),
		0, 0, 0, 0,
	)
}

// MouseScroll 鼠标滚轮（使用mouse_event API）
func MouseScroll(x, y int) {
	mouseEvent.Call(
		uintptr(MOUSEEVENTF_WHEEL),
		0, 0,
		uintptr(uint32(y*120)), // WHEEL_DELTA = 120
		0,
	)
}

// KeyTap 按键点击（使用keybd_event API）
func KeyTap(vkCode uint16) {
	// 按下
	keybdEvent.Call(
		uintptr(vkCode),
		0, 0, 0,
	)
	// 释放
	keybdEvent.Call(
		uintptr(vkCode),
		0,
		uintptr(KEYEVENTF_KEYUP),
		0,
	)
}

// KeyToggle 按键切换（使用keybd_event API）
func KeyToggle(vkCode uint16, state string) {
	var flags uint32
	if state == "up" {
		flags = KEYEVENTF_KEYUP
	}
	keybdEvent.Call(
		uintptr(vkCode),
		0,
		uintptr(flags),
		0,
	)
}

// TypeString 输入字符串（使用keybd_event，兼容360）
func TypeString(text string) {
	for _, r := range text {
		typeChar(r)
	}
}

// typeChar 输入单个字符
func typeChar(char rune) {
	// 检测 CapsLock 状态
	capsLockOn := IsCapsLockOn()

	// 检查是否需要shift
	needShift := false
	var vk uint16

	// 处理常见字符
	switch {
	case char >= 'A' && char <= 'Z':
		// 大写字母
		vk = uint16(char) // A-Z的VK码就是ASCII码
		// 如果 CapsLock 开启，不需要 shift；如果关闭，需要 shift
		needShift = !capsLockOn
	case char >= 'a' && char <= 'z':
		// 小写字母，转换为大写的VK码
		vk = uint16(char - 32) // 转换为大写的VK码
		// 如果 CapsLock 开启，需要 shift（反转）；如果关闭，不需要 shift
		needShift = capsLockOn
	case char >= '0' && char <= '9':
		// 数字键
		vk = uint16(char)
	case char == ' ':
		vk = 0x20 // VK_SPACE
	default:
		// 特殊字符
		vk, needShift = getVKForSpecialChar(char)
		if vk == 0 {
			// 无法处理的字符，使用Unicode方式（可能被360拦截）
			typeCharUnicode(char)
			return
		}
	}

	// 如果需要shift，先按下shift
	if needShift {
		keybdEvent.Call(uintptr(0x10), 0, 0, 0) // VK_SHIFT down
	}

	// 按下并释放键
	keybdEvent.Call(uintptr(vk), 0, 0, 0)                                // key down
	keybdEvent.Call(uintptr(vk), 0, uintptr(KEYEVENTF_KEYUP), 0)         // key up

	// 如果需要shift，释放shift
	if needShift {
		keybdEvent.Call(uintptr(0x10), 0, uintptr(KEYEVENTF_KEYUP), 0) // VK_SHIFT up
	}
}

// getVKForSpecialChar 获取特殊字符的虚拟键码和是否需要shift
func getVKForSpecialChar(char rune) (vk uint16, needShift bool) {
	// 需要shift的特殊字符
	shiftChars := map[rune]uint16{
		'!': 0x31, '@': 0x32, '#': 0x33, '$': 0x34, '%': 0x35, // Shift + 1-5
		'^': 0x36, '&': 0x37, '*': 0x38, '(': 0x39, ')': 0x30, // Shift + 6-0
		'_': 0xBD, // VK_OEM_MINUS
		'+': 0xBB, // VK_OEM_PLUS
		'{': 0xDB, // VK_OEM_4
		'}': 0xDD, // VK_OEM_6
		'|': 0xDC, // VK_OEM_5
		':': 0xBA, // VK_OEM_1
		'"': 0xDE, // VK_OEM_7
		'<': 0xBC, // VK_OEM_COMMA
		'>': 0xBE, // VK_OEM_PERIOD
		'?': 0xBF, // VK_OEM_2
		'~': 0xC0, // VK_OEM_3
	}

	// 不需要shift的特殊字符
	noShiftChars := map[rune]uint16{
		'-': 0xBD, // VK_OEM_MINUS
		'=': 0xBB, // VK_OEM_PLUS
		'[': 0xDB, // VK_OEM_4
		']': 0xDD, // VK_OEM_6
		'\\': 0xDC, // VK_OEM_5
		';': 0xBA, // VK_OEM_1
		'\'': 0xDE, // VK_OEM_7
		',': 0xBC, // VK_OEM_COMMA
		'.': 0xBE, // VK_OEM_PERIOD
		'/': 0xBF, // VK_OEM_2
		'`': 0xC0, // VK_OEM_3
	}

	if vkCode, ok := shiftChars[char]; ok {
		return vkCode, true
	}
	if vkCode, ok := noShiftChars[char]; ok {
		return vkCode, false
	}

	return 0, false
}

// typeCharUnicode 使用Unicode方式输入字符（可能被360拦截）
func typeCharUnicode(char rune) {
	// 按下
	downInput := INPUT{Type: INPUT_KEYBOARD}
	ki := KEYBDINPUT{WScan: uint16(char), DwFlags: KEYEVENTF_UNICODE}
	*(*KEYBDINPUT)(unsafe.Pointer(&downInput.Data[0])) = ki

	// 释放
	upInput := INPUT{Type: INPUT_KEYBOARD}
	ki2 := KEYBDINPUT{WScan: uint16(char), DwFlags: KEYEVENTF_UNICODE | KEYEVENTF_KEYUP}
	*(*KEYBDINPUT)(unsafe.Pointer(&upInput.Data[0])) = ki2

	inputs := []INPUT{downInput, upInput}
	SendInput(inputs)
}

// GetDpiForSystem 获取系统DPI（Windows 10 1607+）
func GetDpiForSystem() uint32 {
	ret, _, _ := getDpiForSystem.Call()
	if ret == 0 {
		return 96 // 默认DPI
	}
	return uint32(ret)
}

// GetSystemDpiScale 获取系统DPI缩放比例
func GetSystemDpiScale() float64 {
	dpi := GetDpiForSystem()
	return float64(dpi) / 96.0 // 96 DPI = 100%
}

// MonitorFromPoint 从点获取监视器句柄
func MonitorFromPoint(x, y int, dwFlags uint32) uintptr {
	pt := POINT{X: int32(x), Y: int32(y)}
	ret, _, _ := monitorFromPoint.Call(
		uintptr(unsafe.Pointer(&pt)),
		uintptr(dwFlags),
	)
	return ret
}

// GetDpiForMonitor 获取指定监视器的DPI
// monitorType: 0 = MDT_EFFECTIVE_DPI, 1 = MDT_ANGULAR_DPI, 2 = MDT_RAW_DPI
func GetDpiForMonitor(hMonitor uintptr) (dpiX, dpiY uint32) {
	var x, y uint32
	ret, _, _ := getDpiForMonitor.Call(
		hMonitor,
		uintptr(0), // MDT_EFFECTIVE_DPI
		uintptr(unsafe.Pointer(&x)),
		uintptr(unsafe.Pointer(&y)),
	)
	if ret != 0 {
		// 失败，返回默认DPI
		return 96, 96
	}
	return x, y
}

// GetDpiScaleForPoint 获取指定点所在监视器的DPI缩放比例
func GetDpiScaleForPoint(x, y int) (float64, uintptr, uint32) {
	const MONITOR_DEFAULTTONEAREST = 2
	hMonitor := MonitorFromPoint(x, y, MONITOR_DEFAULTTONEAREST)
	if hMonitor == 0 {
		return 1.0, 0, 96
	}
	dpiX, _ := GetDpiForMonitor(hMonitor)
	return float64(dpiX) / 96.0, hMonitor, dpiX
}

// RECT 结构体
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// MONITORINFO 结构体
type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

// GetMonitorInfo 获取监视器信息（逻辑坐标）
func GetMonitorInfo(hMonitor uintptr) (*MONITORINFO, error) {
	var mi MONITORINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	ret, _, err := getMonitorInfoW.Call(
		hMonitor,
		uintptr(unsafe.Pointer(&mi)),
	)
	if ret == 0 {
		return nil, err
	}
	return &mi, nil
}

// GetLogicalBoundsForPoint 获取指定点所在监视器的逻辑坐标范围
func GetLogicalBoundsForPoint(x, y int) (left, top, width, height int, ok bool) {
	const MONITOR_DEFAULTTONEAREST = 2
	hMonitor := MonitorFromPoint(x, y, MONITOR_DEFAULTTONEAREST)
	if hMonitor == 0 {
		return 0, 0, 0, 0, false
	}

	mi, err := GetMonitorInfo(hMonitor)
	if err != nil {
		return 0, 0, 0, 0, false
	}

	return int(mi.RcMonitor.Left), int(mi.RcMonitor.Top),
		int(mi.RcMonitor.Right - mi.RcMonitor.Left), int(mi.RcMonitor.Bottom - mi.RcMonitor.Top), true
}

// 全局变量用于缓存显示器信息和callback
var (
	monitorEnumCallback uintptr
	monitorCacheMutex   sync.Mutex
)

// 初始化枚举显示器的callback（只创建一次）
func init() {
	// 创建全局callback，避免每次调用都创建新的callback
	callback := func(hMonitor uintptr, hdcMonitor uintptr, lprcMonitor *RECT, dwData uintptr) uintptr {
		// dwData是指向monitors切片的指针
		monitorsPtr := (*[]monitorData)(unsafe.Pointer(dwData))
		mi, err := GetMonitorInfo(hMonitor)
		if err == nil {
			*monitorsPtr = append(*monitorsPtr, monitorData{
				index: len(*monitorsPtr),
				rect:  mi.RcMonitor,
			})
		}
		return 1 // 继续枚举
	}
	monitorEnumCallback = syscall.NewCallback(callback)
}

type monitorData struct {
	index int
	rect  RECT
}

// GetLogicalBoundsForScreen 获取指定屏幕索引的逻辑坐标范围
// 通过枚举所有显示器来获取
func GetLogicalBoundsForScreen(screenIndex int) (left, top, width, height int, ok bool) {
	monitorCacheMutex.Lock()
	defer monitorCacheMutex.Unlock()

	var monitors []monitorData

	// 调用EnumDisplayMonitors，使用全局callback
	enumDisplayMonitors.Call(
		0, // hdc
		0, // lprcClip
		monitorEnumCallback,
		uintptr(unsafe.Pointer(&monitors)), // 传递monitors切片指针
	)

	// 查找指定索引的显示器
	if screenIndex >= 0 && screenIndex < len(monitors) {
		m := monitors[screenIndex]
		return int(m.rect.Left), int(m.rect.Top),
			int(m.rect.Right - m.rect.Left), int(m.rect.Bottom - m.rect.Top), true
	}

	return 0, 0, 0, 0, false
}
