package web

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/client"
	"github.com/sorc/tcpserver/web/api/handlers"
	"github.com/sorc/tcpserver/web/api/middleware"
	"github.com/sorc/tcpserver/web/api/routes"
)

// Config Web服务器配置
type Config struct {
	HTTPAddr  string `json:"http_addr"`
	TCPAddr   string `json:"tcp_addr"`
	ClientID  string `json:"client_id"`
	Secret    string `json:"secret"`
	JWTSecret string `json:"jwt_secret"`
}

// Server Web服务器
type Server struct {
	config    Config
	router    *gin.Engine
	tcpClient *client.TCPClient
}

// NewServer 创建Web服务器
func NewServer(config Config) *Server {
	// 设置JWT密钥
	if config.JWTSecret != "" {
		middleware.JWTSecret = []byte(config.JWTSecret)
	}

	// 创建TCP客户端
	tcpClient := client.NewTCPClient(config.TCPAddr, config.ClientID, config.Secret)

	// 设置TCP客户端
	handlers.SetTCPClient(tcpClient)

	// 创建路由
	router := routes.SetupRouter()

	// 设置静态文件服务
	uiDir := filepath.Join("web", "ui", "dist")
	if _, err := os.Stat(uiDir); !os.IsNotExist(err) {
		router.StaticFile("/", filepath.Join(uiDir, "index.html"))
		router.Static("/assets", filepath.Join(uiDir, "assets"))
		router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(uiDir, "index.html"))
		})
	}

	return &Server{
		config:    config,
		router:    router,
		tcpClient: tcpClient,
	}
}

// Start 启动Web服务器
func (s *Server) Start() error {
	// 连接TCP服务器
	log.Printf("Connecting to TCP server at %s...", s.config.TCPAddr)
	if err := s.tcpClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to TCP server: %w", err)
	}
	log.Printf("Connected to TCP server successfully")

	// 启动HTTP服务器
	log.Printf("Starting HTTP server at %s...", s.config.HTTPAddr)
	return s.router.Run(s.config.HTTPAddr)
}

// Stop 停止Web服务器
func (s *Server) Stop() error {
	// 断开TCP连接
	if s.tcpClient != nil {
		return s.tcpClient.Disconnect()
	}
	return nil
}
