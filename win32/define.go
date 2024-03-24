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
