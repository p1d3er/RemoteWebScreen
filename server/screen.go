package server

import (
	"bytes"
	"github.com/kbinani/screenshot"
	"image/jpeg"
)

var lastScreen []byte
var currentScreen int

func captureScreen(quality int) ([]byte, error) {
	bounds := screenshot.GetDisplayBounds(currentScreen) //
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		//log.Printf("screenshot Error capturing screen: %v", err)
		return nil, err
	}
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		//log.Printf("Encode Error capturing screen: %v", err)
		return nil, err
	}
	//检测图像变化
	if bytes.Equal(lastScreen, buf.Bytes()) {
		return nil, nil // 没有变化
	}
	lastScreen = make([]byte, len(buf.Bytes()))
	copy(lastScreen, buf.Bytes())
	return buf.Bytes(), nil
}
