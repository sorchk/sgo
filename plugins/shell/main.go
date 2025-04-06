package main

import (
	"context"
	"fmt"

	"github.com/sorc/tcpserver/pkg/plugin"
	"gopkg.in/yaml.v3"
)

// CreateCommandPlugin 创建命令类插件实例
func CreateCommandPlugin() plugin.ICommandPlugin {
	return &ShellPlugin{
		BaseCommandPlugin: plugin.NewBaseCommandPlugin("shell", "Shell Executor", "1.0.0", plugin.InteractiveCommand),
	}
}

// CreatePlugin 创建插件实例
func CreatePlugin() plugin.Plugin {
	return CreateCommandPlugin()
}

// Init 初始化插件
func (p *ShellPlugin) Init(ctx context.Context, configBytes []byte) error {
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
	if config.WorkingDir == "" {
		config.WorkingDir = "."
	}

	p.allowedCommands = config.AllowedCommands
	p.workingDir = config.WorkingDir

	return nil
}

func main() {}
