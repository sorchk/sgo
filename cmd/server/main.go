package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sorc/tcpserver/internal/auth"
	"github.com/sorc/tcpserver/internal/server"
	"github.com/sorc/tcpserver/pkg/plugin"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Server  server.ServerConfig `json:"server"`
	Clients []auth.Client       `json:"clients"`
}

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	// 读取配置文件
	configData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 解析配置
	var config ServerConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// 创建插件管理器
	pluginManager := plugin.NewPluginManager(config.Server.PluginsDir, config.Server.ConfigDir)

	// 创建服务器
	srv, err := server.NewServer(config.Server, pluginManager)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 注册客户端
	for _, client := range config.Clients {
		if err := srv.RegisterClient(&client); err != nil {
			log.Printf("Failed to register client %s: %v", client.ID, err)
		}
	}

	// 加载插件
	if err := loadPlugins(pluginManager, config.Server.PluginsDir); err != nil {
		log.Printf("Warning: Failed to load some plugins: %v", err)
	}

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 处理信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigCh
	log.Println("Shutting down server...")

	// 停止服务器
	if err := srv.Stop(); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	log.Println("Server stopped")
}

// 如果配置文件不存在，创建默认配置
// 加载插件
func loadPlugins(pm plugin.PluginManager, pluginsDir string) error {
	// 检查插件目录是否存在
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugins directory: %w", err)
		}
	}

	// 查找所有.so文件
	soFiles, err := filepath.Glob(filepath.Join(pluginsDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to list plugin files: %w", err)
	}

	// 加载每个插件
	for _, soFile := range soFiles {
		log.Printf("Loading plugin: %s", soFile)
		p, err := pm.LoadPlugin(soFile)
		if err != nil {
			log.Printf("Failed to load plugin %s: %v", soFile, err)
			continue
		}

		// 启用插件
		if err := pm.EnablePlugin(p.ID()); err != nil {
			log.Printf("Failed to enable plugin %s: %v", p.ID(), err)
		} else {
			log.Printf("Plugin %s (%s) loaded and enabled", p.Name(), p.ID())
		}
	}

	return nil
}

func createDefaultConfig(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建默认配置
	config := ServerConfig{
		Server: server.ServerConfig{
			Addr:       ":8888",
			PluginsDir: "plugins",
			ConfigDir:  "config",
		},
		Clients: []auth.Client{
			{
				ID:     "client1",
				Secret: "1234567890123456",
				Name:   "Default Client",
				Permissions: []auth.Permission{
					auth.PermPluginManage,
					auth.PermServiceManage,
					auth.PermPluginUse,
				},
			},
		},
	}

	// 序列化配置
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, configData, 0644)
}
