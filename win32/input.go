package win32

import "strings"

// 提供与robotgo兼容的接口

// GetScreenSize 获取主屏幕尺寸（逻辑坐标）
func GetScreenSize() (int, int) {
	width := GetSystemMetrics(SM_CXSCREEN)
	height := GetSystemMetrics(SM_CYSCREEN)
	return width, height
}

// GetVirtualScreenSize 获取虚拟屏幕尺寸（包含所有显示器）
func GetVirtualScreenSize() (x, y, width, height int) {
	x = GetSystemMetrics(SM_XVIRTUALSCREEN)
	y = GetSystemMetrics(SM_YVIRTUALSCREEN)
	width = GetSystemMetrics(SM_CXVIRTUALSCREEN)
	height = GetSystemMetrics(SM_CYVIRTUALSCREEN)
	return
}

// Click 鼠标点击
func Click(button string, double bool) {
	MouseClick(button, double)
}

// Toggle 鼠标按钮切换
func Toggle(button string, state ...string) {
	s := "down"
	if len(state) > 0 {
		s = state[0]
	}
	MouseToggle(button, s)
}

// Move 移动鼠标
func Move(x, y int) {
	SetCursorPos(x, y)
}

// Scroll 鼠标滚轮
func Scroll(x, y int) {
	MouseScroll(x, y)
}

// TypeStr 输入字符串
func TypeStr(text string) {
	TypeString(text)
}

// GetScreenRect 获取屏幕矩形（简化版，仅支持主屏幕）
type ScreenRect struct {
	X, Y, W, H int
}

func GetScreenRect(screen int) ScreenRect {
	w, h := GetScreenSize()
	return ScreenRect{X: 0, Y: 0, W: w, H: h}
}

// VK码映射表
var keyMap = map[string]uint16{
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45, "f": 0x46,
	"g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A, "k": 0x4B, "l": 0x4C,
	"m": 0x4D, "n": 0x4E, "o": 0x4F, "p": 0x50, "q": 0x51, "r": 0x52,
	"s": 0x53, "t": 0x54, "u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58,
	"y": 0x59, "z": 0x5A,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74, "f6": 0x75,
	"f7": 0x76, "f8": 0x77, "f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
	// 特殊键
	"enter":     0x0D,
	"tab":       0x09,
	"space":     0x20,
	"backspace": 0x08,
	"escape":    0x1B,
	"esc":       0x1B,
	"delete":    0x2E,
	"insert":    0x2D,
	"home":      0x24,
	"end":       0x23,
	"pageup":    0x21,
	"pagedown":  0x22,
	// 方向键
	"left":  0x25,
	"up":    0x26,
	"right": 0x27,
	"down":  0x28,
	// 修饰键
	"shift":   0x10,
	"ctrl":    0x11,
	"control": 0x11,
	"alt":     0x12,
	"cmd":     0x5B, // Windows key
	"command": 0x5B,
	"win":     0x5B,
	// 数字键盘
	"numpad0": 0x60,
	"numpad1": 0x61,
	"numpad2": 0x62,
	"numpad3": 0x63,
	"numpad4": 0x64,
	"numpad5": 0x65,
	"numpad6": 0x66,
	"numpad7": 0x67,
	"numpad8": 0x68,
	"numpad9": 0x69,
	// 其他常用键
	"capslock":   0x14,
	"numlock":    0x90,
	"scrolllock": 0x91,
	"pause":      0x13,
	"printscreen": 0x2C,
}

// KeyTapByName 根据键名按键
func KeyTapByName(key string) {
	// 将键名转换为小写，以处理 CapsLock 开启时的情况
	keyLower := strings.ToLower(key)
	if vk, ok := keyMap[keyLower]; ok {
		KeyTap(vk)
	}
}

// KeyToggleByName 根据键名切换按键状态
func KeyToggleByName(key string, state string) {
	// 将键名转换为小写，以处理 CapsLock 开启时的情况
	keyLower := strings.ToLower(key)
	if vk, ok := keyMap[keyLower]; ok {
		KeyToggle(vk, state)
	}
}
