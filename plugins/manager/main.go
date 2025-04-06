package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sorc/tcpserver/pkg/plugin"
	"gopkg.in/yaml.v3"
)

// CreateCommandPlugin 创建命令类插件实例
func CreateCommandPlugin() plugin.ICommandPlugin {
	return &PluginManagerPlugin{
		BaseCommandPlugin: plugin.NewBaseCommandPlugin("manager", "Plugin Manager", "1.0.0", plugin.OneTimeCommand),
	}
}

// CreatePlugin 创建插件实例
func CreatePlugin() plugin.Plugin {
	return CreateCommandPlugin()
}

// Init 初始化插件
func (p *PluginManagerPlugin) Init(ctx context.Context, configBytes []byte) error {
	if err := p.BaseCommandPlugin.Init(ctx, configBytes); err != nil {
		return err
	}

	// 解析配置
	var config Config
	if len(configBytes) > 0 {
		if err := yaml.Unmarshal(configBytes, &config); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// 设置默认值
	if config.PluginsDir == "" {
		config.PluginsDir = "plugins"
	}
	if config.ConfigDir == "" {
		config.ConfigDir = "config"
	}

	p.pluginsDir = config.PluginsDir
	p.configDir = config.ConfigDir

	// 插件管理器将在服务启动时自动设置

	// 创建配置目录
	if err := os.MkdirAll(p.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

func main() {}
