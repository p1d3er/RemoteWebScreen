package server

import (
	"bytes"
	"crypto/sha256"
	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
	"image/jpeg"
	"sync"
)

// bufferPool 用于复用 bytes.Buffer，减少内存分配
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// screenState 管理屏幕捕获状态
type screenState struct {
	mu            sync.RWMutex
	currentScreen int
}

var state = &screenState{
	currentScreen: 0,
}

// ConnectionScreenState 为每个连接维护独立的屏幕状态
type ConnectionScreenState struct {
	mu       sync.Mutex
	lastHash [32]byte
}

var (
	connStatesMu sync.RWMutex
	connStates   = make(map[*websocket.Conn]*ConnectionScreenState)
)

// GetConnectionState 获取或创建连接的屏幕状态
func GetConnectionState(conn *websocket.Conn) *ConnectionScreenState {
	connStatesMu.RLock()
	state, exists := connStates[conn]
	connStatesMu.RUnlock()

	if exists {
		return state
	}

	// 创建新状态
	connStatesMu.Lock()
	defer connStatesMu.Unlock()

	// 双重检查，避免并发创建
	if state, exists := connStates[conn]; exists {
		return state
	}

	state = &ConnectionScreenState{}
	connStates[conn] = state
	return state
}

// CleanupConnectionState 清理连接的屏幕状态
func CleanupConnectionState(conn *websocket.Conn) {
	connStatesMu.Lock()
	defer connStatesMu.Unlock()
	delete(connStates, conn)
}

// GetCurrentScreen 获取当前屏幕索引
func GetCurrentScreen() int {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.currentScreen
}

// SetCurrentScreen 设置当前屏幕索引
func SetCurrentScreen(screen int) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.currentScreen = screen
}

// SwitchToNextScreen 切换到下一个屏幕
func SwitchToNextScreen() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.currentScreen++
	if state.currentScreen >= screenshot.NumActiveDisplays() {
		state.currentScreen = 0
	}
}

func captureScreen(quality int, conn *websocket.Conn) ([]byte, error) {
	// 获取当前屏幕索引
	currentScreen := GetCurrentScreen()

	bounds := screenshot.GetDisplayBounds(currentScreen)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		//log.Printf("screenshot Error capturing screen: %v", err)
		return nil, err
	}

	// 从对象池获取 buffer
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		//log.Printf("Encode Error capturing screen: %v", err)
		return nil, err
	}

	// 使用哈希比较代替全字节比较，提高性能
	imgBytes := buf.Bytes()
	currentHash := sha256.Sum256(imgBytes)

	// 获取该连接的独立状态
	connState := GetConnectionState(conn)
	connState.mu.Lock()
	defer connState.mu.Unlock()

	if currentHash == connState.lastHash {
		return nil, nil // 没有变化
	}

	connState.lastHash = currentHash

	// 复制数据以便返回（因为 buf 会被放回池中）
	result := make([]byte, len(imgBytes))
	copy(result, imgBytes)

	return result, nil
}
