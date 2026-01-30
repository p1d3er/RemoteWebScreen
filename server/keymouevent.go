package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"math"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"RemoteWebScreen/keyboard"
	"RemoteWebScreen/win32"
)

var CaptureScreenquality int = 56

// cmdSession 存储每个连接的命令会话信息
type cmdSession struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	conn    *websocket.Conn
	mu      sync.Mutex
}

var (
	sessionsMu sync.RWMutex
	sessions   = make(map[*websocket.Conn]*cmdSession)
)

func SimulateDesktopHDMessage(conn *websocket.Conn, msg []byte) {
	var message map[string]interface{}
	if err := json.Unmarshal(msg, &message); err != nil {
		//log.Printf("if err := json.Unmarshal(msg, &message); err != nil Error: %v", err)
		return
	}

	switch messageType := message["type"].(string); messageType {
	case "1": // 移动
		go handleMouseMove(message)
	case "2": // 左键点击
		//log.Println("[DEBUG] Left click")
		win32.Click("left", false)
	case "3": // 键盘
		//log.Println("[DEBUG] Keyboard event:", message)
		handleKeyDown(message) // 使用keybd_event方式（更可靠）
	case "4": // 右击
		//log.Println("[DEBUG] Right click")
		win32.Click("right", false)
	case "5": // 鼠标按下
		//log.Println("[DEBUG] Mouse down")
		win32.Toggle("left")
	case "6": // 鼠标起来
		//log.Println("[DEBUG] Mouse up")
		win32.Toggle("left", "up")
	case "updateSettings":
		quality := message["quality"].(float64)
		CaptureScreenquality = int(quality)
	case "7": // 鼠标滚轮
		handleMouseScroll(message)
	case "8": // 处理键盘事件
		//log.Println("[DEBUG] Combo key event:", message)
		handleComboKeyEvent(message)
	case "9":
		SwitchToNextScreen()
	case "10":
		conn.Close()
	case "11": // 命令执行
		go handleCommandExecution(conn, message)
	case "12": // 获取日志
		go handleGetLog(conn, message)
	}
}
func handleComboKeyEvent(message map[string]interface{}) {
	key, _ := message["key"].(string)
	modifiers, _ := message["modifiers"].(map[string]interface{})

	// 对于每个修饰键，按下之前先检查它是否激活
	for mod, active := range modifiers {
		if active.(bool) {
			win32.KeyToggleByName(mod, "down")
		}
	}
	// 模拟按下主键
	win32.KeyTapByName(key)
	// 释放所有之前按下的修饰键
	for mod, active := range modifiers {
		if active.(bool) {
			win32.KeyToggleByName(mod, "up")
		}
	}
}

func handleMouseScroll(message map[string]interface{}) {
	direction, _ := message["direction"].(string)
	amount, _ := message["amount"].(float64)
	// 减少滚动量和调整方向
	scrollAmount := int(amount) / 36 // 除以一个因子来减速
	if direction == "up" {
		win32.Scroll(0, scrollAmount) // 上滚
	} else if direction == "down" {
		win32.Scroll(0, -scrollAmount) // 下滚
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

	// 边界检查：确保画布尺寸有效
	if canvasWidth <= 0 || canvasHeight <= 0 {
		return
	}

	// 获取当前屏幕的边界
	currentScreen := GetCurrentScreen()

	// 检查屏幕索引是否有效
	numDisplays := screenshot.NumActiveDisplays()
	if currentScreen < 0 || currentScreen >= numDisplays {
		return
	}

	// 获取截图库返回的屏幕边界（物理像素）
	bounds := screenshot.GetDisplayBounds(currentScreen)

	// 计算相对位置（0.0 到 1.0）
	relativeX := absX / canvasWidth
	relativeY := absY / canvasHeight

	// 获取虚拟屏幕边界
	vx, vy, vw, vh := win32.GetVirtualScreenSize()

	// 直接根据屏幕索引获取逻辑坐标范围
	logicalLeft, logicalTop, logicalWidth, logicalHeight, ok := win32.GetLogicalBoundsForScreen(currentScreen)

	var dpiScale float64 = 1.0
	if ok && logicalWidth > 0 && logicalHeight > 0 {
		// 通过比较物理尺寸和逻辑尺寸来计算DPI缩放比例
		scaleX := float64(bounds.Dx()) / float64(logicalWidth)
		_ = float64(bounds.Dy()) / float64(logicalHeight) // scaleY (未使用)
		dpiScale = scaleX // 使用X方向的缩放比例
		//log.Printf("[DEBUG] Physical bounds: %dx%d, Logical bounds: %dx%d", bounds.Dx(), bounds.Dy(), logicalWidth, logicalHeight)
		//log.Printf("[DEBUG] Calculated DPI scale: %.3f (scaleX=%.3f, scaleY=%.3f)", dpiScale, scaleX, scaleY)
	}

	// 详细调试日志
	//log.Printf("[DEBUG] ===== Mouse Move Debug =====")
	//log.Printf("[DEBUG] Current Screen: %d", currentScreen)
	//log.Printf("[DEBUG] Physical Bounds: (%d,%d) %dx%d", bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
	//log.Printf("[DEBUG] Logical Bounds: (%d,%d) %dx%d", logicalLeft, logicalTop, logicalWidth, logicalHeight)
	//log.Printf("[DEBUG] Virtual Screen: (%d,%d) %dx%d", vx, vy, vw, vh)
	//log.Printf("[DEBUG] DPI Scale: %.2f (%.0f%%)", dpiScale, dpiScale*100)
	//log.Printf("[DEBUG] Input: absX=%.1f, absY=%.1f, canvasW=%.1f, canvasH=%.1f", absX, absY, canvasWidth, canvasHeight)
	//log.Printf("[DEBUG] Relative: (%.3f, %.3f)", relativeX, relativeY)

	// screenshot库返回的是物理像素坐标
	// Windows SetCursorPos需要逻辑坐标
	// 需要将物理坐标除以DPI缩放比例得到逻辑坐标

	// 计算物理坐标（相对于当前屏幕左上角）
	physicalRelX := relativeX * float64(bounds.Dx())
	physicalRelY := relativeY * float64(bounds.Dy())

	// 转换为逻辑坐标
	logicalRelX := physicalRelX / dpiScale
	logicalRelY := physicalRelY / dpiScale

	// 使用逻辑坐标的偏移
	screenX := logicalLeft + int(math.Round(logicalRelX))
	screenY := logicalTop + int(math.Round(logicalRelY))

	//log.Printf("[DEBUG] Calculated: screenX=%d, screenY=%d", screenX, screenY)

	// 边界限制
	//originalX, originalY := screenX, screenY
	if screenX < vx {
		screenX = vx
	} else if screenX >= vx+vw {
		screenX = vx + vw - 1
	}

	if screenY < vy {
		screenY = vy
	} else if screenY >= vy+vh {
		screenY = vy + vh - 1
	}

	// if screenX != originalX || screenY != originalY {
	// 	//log.Printf("[DEBUG] Clamped: (%d,%d) -> (%d,%d)", originalX, originalY, screenX, screenY)
	// }

	//log.Printf("[DEBUG] Final: Moving to (%d, %d)", screenX, screenY)
	//log.Printf("[DEBUG] =============================")

	win32.Move(screenX, screenY)
}

func handleKeyDown(message map[string]interface{}) {
	keyCode, ok := message["keyCode"].(string)
	if !ok {
		return
	}
	// 转换JavaScript键名到我们的格式
	keyCode = normalizeKeyName(keyCode)

	// 单个字母或数字，使用TypeStr（支持大小写和特殊字符）
	if len(keyCode) == 1 {
		win32.TypeStr(keyCode)
	} else {
		// 特殊键，使用KeyTapByName
		win32.KeyTapByName(keyCode)
	}
}

// normalizeKeyName 将JavaScript的键名转换为我们的格式
func normalizeKeyName(key string) string {
	// 对于单个字符，保持原样（保留大小写）
	if len(key) == 1 {
		return key
	}

	// 对于多字符键名，转换为小写后处理
	keyLower := strings.ToLower(key)

	// 处理JavaScript的特殊键名
	switch keyLower {
	case "arrowleft":
		return "left"
	case "arrowright":
		return "right"
	case "arrowup":
		return "up"
	case "arrowdown":
		return "down"
	case "delete":
		return "delete"
	case "backspace":
		return "backspace"
	case "enter":
		return "enter"
	case "tab":
		return "tab"
	case "escape":
		return "escape"
	}

	return keyLower
}

// startCmdProcess 启动持久化的 cmd 进程
func startCmdProcess(session *cmdSession) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd")
	} else {
		cmd = exec.Command("sh")
	}

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	session.cmd = cmd
	session.stdin = stdin
	session.stdout = stdout
	session.stderr = stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		return err
	}

	// 启动输出读取协程
	go readCmdOutput(session, stdout, false)
	go readCmdOutput(session, stderr, true)

	return nil
}

// readCmdOutput 读取 cmd 输出并发送到 WebSocket
func readCmdOutput(session *cmdSession, reader io.ReadCloser, isError bool) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		outputStr := convertGBKToUTF8([]byte(line + "\n"))

		if isError {
			sendCommandErrorWithSession(session, outputStr)
		} else {
			sendCommandOutputWithSession(session, outputStr)
		}
	}
}

// handleCommandExecution 处理命令执行
func handleCommandExecution(conn *websocket.Conn, message map[string]interface{}) {
	command, ok := message["command"].(string)
	if !ok || command == "" {
		sendCommandError(conn, "无效的命令")
		return
	}

	// 获取或创建会话
	sessionsMu.Lock()
	session, exists := sessions[conn]
	if !exists {
		session = &cmdSession{
			conn: conn,
		}
		sessions[conn] = session

		// 启动持久化的 cmd 进程
		if err := startCmdProcess(session); err != nil {
			sessionsMu.Unlock()
			sendCommandError(conn, "无法启动命令进程: "+err.Error())
			return
		}
	}
	sessionsMu.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	// 发送命令到 stdin
	if session.stdin != nil {
		_, err := session.stdin.Write([]byte(command + "\r\n"))
		if err != nil {
			sendCommandError(conn, "命令发送失败: "+err.Error())
			return
		}
	}
}

// convertGBKToUTF8 将 GBK 编码转换为 UTF-8
func convertGBKToUTF8(gbkData []byte) string {
	if runtime.GOOS != "windows" {
		return string(gbkData)
	}

	// 尝试 GBK 转 UTF-8
	reader := transform.NewReader(bytes.NewReader(gbkData), simplifiedchinese.GBK.NewDecoder())
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		// 如果转换失败，返回原始字符串
		return string(gbkData)
	}
	return string(utf8Data)
}

// CleanupSession 清理连接的会话
func CleanupSession(conn *websocket.Conn) {
	sessionsMu.Lock()
	session, exists := sessions[conn]
	if exists {
		// 关闭所有管道
		if session.stdin != nil {
			session.stdin.Close()
		}
		if session.stdout != nil {
			session.stdout.Close()
		}
		if session.stderr != nil {
			session.stderr.Close()
		}
		// 终止 cmd 进程
		if session.cmd != nil && session.cmd.Process != nil {
			session.cmd.Process.Kill()
			// 等待进程完全退出，避免僵尸进程
			session.cmd.Wait()
		}
		delete(sessions, conn)
	}
	sessionsMu.Unlock()
}

// sendCommandOutput 发送命令输出到客户端
func sendCommandOutput(conn *websocket.Conn, output string) {
	response := map[string]interface{}{
		"type":   "cmdOutput",
		"output": output,
	}
	data, _ := json.Marshal(response)
	safeWriteMessage(conn, websocket.TextMessage, data)
}

// sendCommandOutputWithSession 发送命令输出到客户端（带互斥锁保护）
func sendCommandOutputWithSession(session *cmdSession, output string) {
	response := map[string]interface{}{
		"type":   "cmdOutput",
		"output": output,
	}
	data, _ := json.Marshal(response)
	safeWriteMessage(session.conn, websocket.TextMessage, data)
}

// sendCommandError 发送命令错误到客户端
func sendCommandError(conn *websocket.Conn, errorMsg string) {
	response := map[string]interface{}{
		"type":  "cmdError",
		"error": errorMsg,
	}
	data, _ := json.Marshal(response)
	safeWriteMessage(conn, websocket.TextMessage, data)
}

// sendCommandErrorWithSession 发送命令错误到客户端（带互斥锁保护）
func sendCommandErrorWithSession(session *cmdSession, errorMsg string) {
	response := map[string]interface{}{
		"type":  "cmdError",
		"error": errorMsg,
	}
	data, _ := json.Marshal(response)
	safeWriteMessage(session.conn, websocket.TextMessage, data)
}

// handleGetLog 处理获取日志请求
func handleGetLog(conn *websocket.Conn, message map[string]interface{}) {
	// 导入 keyboard 包以访问日志路径
	// 读取日志文件
	content, err := readLogFile()
	if err != nil {
		sendLogContent(conn, "读取日志失败: "+err.Error())
		return
	}

	sendLogContent(conn, content)
}

// sendLogContent 发送日志内容到客户端
func sendLogContent(conn *websocket.Conn, content string) {
	response := map[string]interface{}{
		"type":    "logContent",
		"content": content,
	}
	data, _ := json.Marshal(response)
	safeWriteMessage(conn, websocket.TextMessage, data)
}

// readLogFile 读取日志文件内容
func readLogFile() (string, error) {
	filePath := filepath.Join(keyboard.Screen_logPath, keyboard.Logfilename)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
