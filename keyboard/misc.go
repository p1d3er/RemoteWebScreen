package keyboard

import (
	"RemoteWebScreen/win32"
	"github.com/axgle/mahonia"
)

// 获取前置窗口句柄和标题
func getForegroundWindow() (win32.HWND, string, error) {
	hWnd, err := win32.GetForegroundWindow()
	if err != nil {
		return 0, "", err
	}
	windowText, err := getWindowText(hWnd)
	return hWnd, windowText, nil
}

// 获取窗口标题
func getWindowText(hWnd win32.HWND) (string, error) {
	windowText, err := win32.GetWindowTextA(hWnd)
	if err != nil {
		return "", err
	}
	if acp == 936 { // gbk
		dec := mahonia.NewDecoder("gbk")
		windowText = dec.ConvertString(windowText)
	}
	return windowText, nil
}

// 获取进程id和进程名字
func getProcessInfo(hWnd win32.HWND) (uint32, string, error) {
	pid, _, err := win32.GetWindowThreadProcessId(hWnd)
	if err != nil {
		return 0, "", err
	}
	handle, err := win32.OpenProcess(0x400|0x10, false, pid)
	if err != nil {
		return 0, "", err
	}
	defer win32.CloseHandel(handle)
	name, err := win32.GetModuleBaseNameA(handle)
	if err != nil {
		return 0, "", err
	}
	if acp == 936 { // gbk
		dec := mahonia.NewDecoder("gbk")
		name = dec.ConvertString(name)
	}
	return pid, name, nil
}

func getCapsLockSate() bool {
	state, _ := win32.GetKeyState(20)
	if state == 0 {
		return false
	}
	return true
}
