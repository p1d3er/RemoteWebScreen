package win32

import (
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
	windowFromPoint          = user32.NewProc("WindowFromPoint")
	getAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	psapi                    = syscall.NewLazyDLL("psapi.dll")
	getModuleBaseNameA       = psapi.NewProc("GetModuleBaseNameA")
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
