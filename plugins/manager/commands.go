package main

import (
	"context"
	"fmt"
	"io"
)

// GetCommands 获取支持的命令列表
func (p *PluginManagerPlugin) GetCommands() []string {
	return []string{
		"list",
		"install",
		"uninstall",
		"enable",
		"disable",
		"upgrade",
		"info",
		"start",
		"stop",
		"restart",
		"status",
		"config",
	}
}

// Execute 执行命令
func (p *PluginManagerPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "list":
		return p.listPlugins(ctx, cmdArgs, output)
	case "install":
		return p.installPlugin(ctx, cmdArgs, output)
	case "uninstall":
		return p.uninstallPlugin(ctx, cmdArgs, output)
	case "enable":
		return p.enablePlugin(ctx, cmdArgs, output)
	case "disable":
		return p.disablePlugin(ctx, cmdArgs, output)
	case "upgrade":
		return p.upgradePlugin(ctx, cmdArgs, output)
	case "info":
		return p.pluginInfo(ctx, cmdArgs, output)
	case "start":
		return p.startService(ctx, cmdArgs, output)
	case "stop":
		return p.stopService(ctx, cmdArgs, output)
	case "restart":
		return p.restartService(ctx, cmdArgs, output)
	case "status":
		return p.serviceStatus(ctx, cmdArgs, output)
	case "config":
		return p.configService(ctx, cmdArgs, output)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
