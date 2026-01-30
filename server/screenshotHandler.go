package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// gzipWriterPool 用于复用 gzip.Writer，减少内存分配
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// bufferPool 用于复用 bytes.Buffer
var sendBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ConnectionPool 管理 WebSocket 连接池
type ConnectionPool struct {
	mu    sync.RWMutex
	conns map[*websocket.Conn]bool
}

// connWriteMutex 保护每个连接的写操作
var (
	connWriteMutexes   = make(map[*websocket.Conn]*sync.Mutex)
	connWriteMutexesMu sync.RWMutex
)

// getConnWriteMutex 获取连接的写锁
func getConnWriteMutex(conn *websocket.Conn) *sync.Mutex {
	connWriteMutexesMu.RLock()
	mu, exists := connWriteMutexes[conn]
	connWriteMutexesMu.RUnlock()

	if exists {
		return mu
	}

	connWriteMutexesMu.Lock()
	defer connWriteMutexesMu.Unlock()

	// 双重检查
	if mu, exists := connWriteMutexes[conn]; exists {
		return mu
	}

	mu = &sync.Mutex{}
	connWriteMutexes[conn] = mu
	return mu
}

// removeConnWriteMutex 移除连接的写锁
func removeConnWriteMutex(conn *websocket.Conn) {
	connWriteMutexesMu.Lock()
	defer connWriteMutexesMu.Unlock()
	delete(connWriteMutexes, conn)
}

// safeWriteMessage 安全地写入 WebSocket 消息
func safeWriteMessage(conn *websocket.Conn, messageType int, data []byte) error {
	mu := getConnWriteMutex(conn)
	mu.Lock()
	defer mu.Unlock()
	return conn.WriteMessage(messageType, data)
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		conns: make(map[*websocket.Conn]bool),
	}
}

// Add 添加连接到池中
func (cp *ConnectionPool) Add(conn *websocket.Conn) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.conns[conn] = true
}

// Remove 从池中移除连接
func (cp *ConnectionPool) Remove(conn *websocket.Conn) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.conns, conn)
}

// Cleanup 清理无效连接
func (cp *ConnectionPool) Cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for conn := range cp.conns {
		if err := safeWriteMessage(conn, websocket.PingMessage, nil); err != nil {
			conn.Close()
			delete(cp.conns, conn)
		}
	}
}

var connectionPool = NewConnectionPool()

func ScreenshotHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Printf("Error screenshotHandler: %v", err)
		return
	}

	// 使用 context 控制所有 goroutine 的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 添加到连接池
	connectionPool.Add(conn)
	defer func() {
		connectionPool.Remove(conn)
		CleanupSession(conn)        // 清理命令会话
		CleanupConnectionState(conn) // 清理屏幕状态
		removeConnWriteMutex(conn)   // 清理写锁
		conn.Close()
	}()

	var wg sync.WaitGroup

	// 心跳 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := safeWriteMessage(conn, websocket.PingMessage, []byte{}); err != nil {
					//log.Printf("Ping failed: %v", err)
					cancel() // 通知其他 goroutine 退出
					return
				}
			}
		}
	}()

	// 屏幕推流 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(33 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				imgBytes, err := captureScreen(CaptureScreenquality, conn)
				if err != nil {
					//log.Printf("imgBytes, err := captureScreen(captureScreenquality, captureScreenscale) Error: %v", err)
					cancel()
					return
				}
				if imgBytes == nil {
					continue
				}
				err = sendImage(conn, imgBytes)
				if err != nil {
					//log.Printf("err = sendImage(conn, imgBytes) Error: %v", err)
					cancel()
					return
				}
			}
		}
	}()

	// 消息接收 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					//log.Printf("_, msg, err := conn.ReadMessage Error: %v", err)
					cancel()
					return
				}
				SimulateDesktopHDMessage(conn, msg)
			}
		}
	}()

	// 等待所有 goroutine 完成
	wg.Wait()
}

// CleanupConnections 关闭无用连接
func CleanupConnections() {
	connectionPool.Cleanup()
}

func sendImage(conn *websocket.Conn, imgBytes []byte) error {
	// 从对象池获取 buffer
	buf := sendBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer sendBufferPool.Put(buf)

	// 从对象池获取 gzip.Writer
	gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
	gzipWriter.Reset(buf)
	defer gzipWriterPool.Put(gzipWriter)

	// 写入数据
	if _, err := gzipWriter.Write(imgBytes); err != nil {
		//log.Printf("Write Error sending image: %v", err)
		return err
	}

	// 关闭 gzip writer 以刷新缓冲区
	if err := gzipWriter.Close(); err != nil {
		//log.Printf("Close Error sending image: %v", err)
		return err
	}

	// 发送压缩后的数据（使用安全写入）
	err := safeWriteMessage(conn, websocket.BinaryMessage, buf.Bytes())
	return err
}
