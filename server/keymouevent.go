package server

import (
	"encoding/json"
	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
	"math"
)

var CaptureScreenquality int = 56
var mainScreenScale float64
var assistScreenScale float64

func init() {
	bounds := screenshot.GetDisplayBounds(currentScreen)
	screenW, _ := robotgo.GetScreenSize()
	mainScreenScale = math.Round(float64(bounds.Dx())/float64(screenW)*100) / 100
}
func SimulateDesktopHDMessage(conn *websocket.Conn, msg []byte) {
	var message map[string]interface{}
	if err := json.Unmarshal(msg, &message); err != nil {
		//log.Printf("if err := json.Unmarshal(msg, &message); err != nil Error: %v", err)
		return
	}

	switch messageType := message["type"].(string); messageType {
	case "1": // 移动
		go handleMouseMove(message)
	case "2": // 移动
		robotgo.Click("left", false)
	case "3": // 键盘
		//log.Println("3", message)
		handleKeyDown(message)
	case "4": // 右击
		robotgo.MouseClick("right", false)
	case "5": // 鼠标按下
		robotgo.Toggle("left")
	case "6": // 鼠标起来
		robotgo.Toggle("left", "up")
	case "updateSettings":
		quality := message["quality"].(float64)
		CaptureScreenquality = int(quality)
	case "7": // 鼠标滚轮
		handleMouseScroll(message)
	case "8": // 处理键盘事件
		//log.Println("8", message)
		handleComboKeyEvent(message)
	case "9":
		currentScreen++
		if currentScreen >= screenshot.NumActiveDisplays() {
			currentScreen = 0
		}
	case "10":
		conn.Close()
	}
}
func handleComboKeyEvent(message map[string]interface{}) {
	key, _ := message["key"].(string)
	modifiers, _ := message["modifiers"].(map[string]interface{})

	// 对于每个修饰键，按下之前先检查它是否激活
	for mod, active := range modifiers {
		if active.(bool) {
			robotgo.KeyToggle(mod, "down")
		}
	}
	// 模拟按下主键
	robotgo.KeyTap(key)
	// 释放所有之前按下的修饰键
	for mod, active := range modifiers {
		if active.(bool) {
			robotgo.KeyToggle(mod, "up")
		}
	}
}

func handleMouseScroll(message map[string]interface{}) {
	direction, _ := message["direction"].(string)
	amount, _ := message["amount"].(float64)
	// 减少滚动量和调整方向
	scrollAmount := int(amount) / 36 // 除以一个因子来减速
	if direction == "up" {
		robotgo.Scroll(0, scrollAmount) // 上滚
	} else if direction == "down" {
		robotgo.Scroll(0, -scrollAmount) // 下滚
	}
}

func checkScale(scale float64) int {
	scaleInt := int(math.Round(scale * 100))
	switch scaleInt {
	case 100, 125, 150, 175:
		return scaleInt
	default:
		return 0
	}
}
func handleMouseMove(message map[string]interface{}) {
	absX, absY := message["absX"].(float64), message["absY"].(float64)
	canvasWidth, canvasHeight := message["canvasWidth"].(float64), message["canvasHeight"].(float64)
	var scaleX, scaleY float64
	var screenX, screenY int
	// 获取当前屏幕的边界
	bounds := screenshot.GetDisplayBounds(currentScreen)
	screen := robotgo.GetScreenRect(currentScreen) //1
	// 示例函数：获取主屏幕和扩展屏幕的缩放比例

	if currentScreen == 0 {
		screenW, _ := robotgo.GetScreenSize()
		mainScreenScale = math.Round(float64(bounds.Dx())/float64(screenW)*100) / 100
		scaleX = float64(bounds.Dx()) / mainScreenScale / canvasWidth
		scaleY = float64(bounds.Dy()) / mainScreenScale / canvasHeight
		screenX = bounds.Min.X + int(math.Round(absX*scaleX))
		screenY = bounds.Min.Y + int(math.Round(absY*scaleY))
	} else {
		var ScreenScalex float64
		ScreenScaleA := float64(bounds.Dx()) * (mainScreenScale / float64(screen.W-bounds.Min.X))
		ScreenScaleB := float64(mainScreenScale) * float64(bounds.Min.X+bounds.Dx()) / float64(screen.W)
		if float64(checkScale(ScreenScaleA)/100) == 0 {
			assistScreenScale = float64(checkScale(ScreenScaleB)) / 100
		} else {
			assistScreenScale = float64(checkScale(ScreenScaleA)) / 100
		}
		scaleX = float64(bounds.Dx()) / canvasWidth
		scaleY = float64(bounds.Dy()) / canvasHeight
		////<以下不适配分辨率
		////175	175
		////175	150
		////175	125
		////150	125
		//
		//if mainScreenScale > assistScreenScale {
		//	ScreenScalex = 1
		//} else {
		//	ScreenScalex = mainScreenScale
		//>}
		switch {
		case mainScreenScale >= 1.75 && assistScreenScale != 1:
			ScreenScalex = assistScreenScale
		case mainScreenScale == 1.5 && assistScreenScale == 1.25:
			ScreenScalex = assistScreenScale
		case mainScreenScale > assistScreenScale:
			ScreenScalex = 1
		default:
			ScreenScalex = mainScreenScale
		}
		screenX = int(float64(bounds.Min.X)/ScreenScalex) + int(math.Round(absX*scaleX/assistScreenScale))
		screenY = bounds.Min.Y + int(math.Round(absY*scaleY/assistScreenScale))
	}
	robotgo.Move(screenX, screenY)
}

func handleKeyDown(message map[string]interface{}) {
	keyCode, ok := message["keyCode"].(string)
	if !ok {
		return
	}

	isUpperCase := (keyCode >= "A" && keyCode <= "Z" || keyCode >= "a" && keyCode <= "z") && len(keyCode) == 1
	if isUpperCase {
		robotgo.TypeStr(keyCode)
		isUpperCase = false
	} else {
		robotgo.KeyTap(keyCode)
	}
}
