package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sorc/tcpserver/web"
)

func main() {
	// 解析命令行参数
	configFile := flag.String("config", "webclient.json", "Path to config file")
	flag.Parse()

	// 读取配置文件
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 解析配置
	var config web.Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// 创建Web服务器
	server := web.NewServer(config)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动Web服务器
	go func() {
		fmt.Printf("Starting Web UI server at %s\n", config.HTTPAddr)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待信号
	<-sigChan
	fmt.Println("\nShutting down...")

	// 停止Web服务器
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	fmt.Println("Server stopped")
}
