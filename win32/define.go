package win32

const (
	/*	WH_MIN             = -1
		WH_MSGFILTER       = -1
		WH_JOURNALRECORD   = 0
		WH_JOURNALPLAYBACK = 1
		WH_KEYBOARD        = 2
		WH_GETMESSAGE      = 3
		WH_CALLWNDPROC     = 4
		WH_CBT             = 5
		WH_SYSMSGFILTER    = 6
		WH_MOUSE           = 7
		WH_HARDWARE        = 8
		WH_DEBUG           = 9
		WH_SHELL           = 10
		WH_FOREGROUNDIDLE  = 11
		WH_CALLWNDPROCRET  = 12*/
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	/*WH_MAX             = 15

	WM_MOUSEFIRST  = 0x0200
	WM_MOUSEMOVE   = 0x0200*/
	WM_LBUTTONDOWN = 0x0201
	/*	WM_LBUTTONUP     = 0x0202
		WM_LBUTTONDBLCLK = 0x0203
		WM_RBUTTONDOWN   = 0x0204
		WM_RBUTTONUP     = 0x0205
		WM_RBUTTONDBLCLK = 0x0206
		WM_MBUTTONDOWN   = 0x0207
		WM_MBUTTONUP     = 0x0208
		WM_MBUTTONDBLCLK = 0x0209
		WM_MOUSEWHEEL    = 0x020A
		WM_MOUSELAST     = 0x020A

		WM_KEYFIRST = 0x0100*/
	WM_KEYDOWN = 0x0100
	/*	WM_KEYUP       = 0x0101
		WM_CHAR        = 0x0102
		WM_DEADCHAR    = 0x0103
		WM_SYSKEYDOWN  = 0x0104
		WM_SYSKEYUP    = 0x0105
		WM_SYSCHAR     = 0x0106
		WM_SYSDEADCHAR = 0x0107
		WM_KEYLAST     = 0x0108*/
)
const (
	VK_CONTROL = 0x11
)

// SendInput constants
const (
	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	MOUSEEVENTF_MOVE       = 0x0001
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_WHEEL      = 0x0800
	MOUSEEVENTF_ABSOLUTE   = 0x8000

	KEYEVENTF_KEYUP       = 0x0002
	KEYEVENTF_UNICODE     = 0x0004
	KEYEVENTF_SCANCODE    = 0x0008
	KEYEVENTF_EXTENDEDKEY = 0x0001

	SM_CXSCREEN        = 0
	SM_CYSCREEN        = 1
	SM_XVIRTUALSCREEN  = 76
	SM_YVIRTUALSCREEN  = 77
	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79
)

// INPUT structure for SendInput
// 在Go中模拟C的union，使用字节数组
type INPUT struct {
	Type uint32
	_    uint32 // padding for 64-bit alignment
	Data [32]byte // 增加到32字节以容纳MOUSEINPUT
}

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type HARDWAREINPUT struct {
	UMsg    uint32
	WParamL uint16
	WParamH uint16
}

type (
	ATOM          uint16
	HANDLE        uintptr
	HGLOBAL       HANDLE
	HINSTANCE     HANDLE
	LCID          uint32
	LCTYPE        uint32
	LANGID        uint16
	HMODULE       uintptr
	HWINEVENTHOOK HANDLE
	HRSRC         uintptr

	HACCEL    HANDLE
	HCURSOR   HANDLE
	HDWP      HANDLE
	HICON     HANDLE
	HMENU     HANDLE
	HMONITOR  HANDLE
	HRAWINPUT HANDLE
	HKL       HANDLE
	DWORD     uint32
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	HHOOK     HANDLE
	HWND      HANDLE
)

var NULL = 0
var (
	Jianpan_hok HHOOK
	Shubao_hok  HHOOK
)

type HOOKPROC func(int, WPARAM, LPARAM) LRESULT

type KBDLLHOOKSTRUCT struct {
	VkCode      DWORD
	ScanCode    DWORD
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   POINT
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}
type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

const (
	SW_HIDE = 0
	SW_SHOW = 5
)

var (
	getConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
	getCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")
	showWindowAsync     = user32.NewProc("ShowWindowAsync")
)

const HtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Log Viewer</title>
</head>
<style>
    body {
        font-family: Arial, sans-serif;
        line-height: 1.6;
        background-color: #f7f7f7;
        padding: 20px;
    }
    .txt-content {
        background-color: #fff;
        border-radius: 5px;
        padding: 20px;
        box-shadow: 0 0 10px rgba(0,0,0,0.1);
    }
    pre {
        white-space: pre-wrap; 
        word-wrap: break-word; 
    }
    .img{
        width: 50%; 
        height: auto;
    }
</style>
</head>
<body>

<div class="txt-content">
	<pre>
{{ .LogContent }}
	</pre>
</body>
</html>
`
