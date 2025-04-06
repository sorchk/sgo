package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sorc/tcpserver/pkg/plugin"
)

// startService 启动服务
func (p *PluginManagerPlugin) startService(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: start <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plug, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件类型
	if plug.Type() != plugin.ServicePlugin {
		return fmt.Errorf("plugin %s is not a service plugin", pluginID)
	}

	// 获取服务插件
	servicePlugin, err := p.pluginManager.GetServicePlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get service plugin: %w", err)
	}

	// 启动服务
	if err := servicePlugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Fprintf(output, "Service %s started successfully\n", pluginID)
	return nil
}

// stopService 停止服务
func (p *PluginManagerPlugin) stopService(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: stop <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plug, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件类型
	if plug.Type() != plugin.ServicePlugin {
		return fmt.Errorf("plugin %s is not a service plugin", pluginID)
	}

	// 获取服务插件
	servicePlugin, err := p.pluginManager.GetServicePlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get service plugin: %w", err)
	}

	// 停止服务
	if err := servicePlugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	fmt.Fprintf(output, "Service %s stopped successfully\n", pluginID)
	return nil
}

// restartService 重启服务
func (p *PluginManagerPlugin) restartService(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: restart <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plug, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件类型
	if plug.Type() != plugin.ServicePlugin {
		return fmt.Errorf("plugin %s is not a service plugin", pluginID)
	}

	// 获取服务插件
	servicePlugin, err := p.pluginManager.GetServicePlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get service plugin: %w", err)
	}

	// 重启服务
	if err := servicePlugin.Restart(ctx); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	fmt.Fprintf(output, "Service %s restarted successfully\n", pluginID)
	return nil
}

// serviceStatus 获取服务状态
func (p *PluginManagerPlugin) serviceStatus(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	// 如果没有指定插件ID，列出所有服务插件的状态
	if len(args) == 0 {
		plugins := p.pluginManager.ListPlugins()

		fmt.Fprintln(output, "Service Plugins Status:")
		fmt.Fprintln(output, "ID\tName\tVersion\tState")
		fmt.Fprintln(output, "----------------------------------------------------")

		for _, plug := range plugins {
			if plug.Type() == plugin.ServicePlugin {
				var stateStr string
				if plug.State() == 0 {
					stateStr = "Disabled"
				} else if plug.State() == 1 {
					stateStr = "Enabled"
				} else if plug.State() == 2 {
					stateStr = "Running"
				} else if plug.State() == 3 {
					stateStr = "Paused"
				} else {
					stateStr = "Unknown"
				}

				fmt.Fprintf(output, "%s\t%s\t%s\t%s\n", plug.ID(), plug.Name(), plug.Version(), stateStr)
			}
		}

		return nil
	}

	// 获取指定插件的状态
	pluginID := args[0]

	// 获取插件
	plug, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件类型
	if plug.Type() != plugin.ServicePlugin {
		return fmt.Errorf("plugin %s is not a service plugin", pluginID)
	}

	// 获取服务插件
	servicePlugin, err := p.pluginManager.GetServicePlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get service plugin: %w", err)
	}

	// 获取状态
	var stateStr string
	switch servicePlugin.State() {
	case plugin.Disabled:
		stateStr = "Disabled"
	case plugin.Enabled:
		stateStr = "Enabled"
	case plugin.Running:
		stateStr = "Running"
	case plugin.Paused:
		stateStr = "Paused"
	default:
		stateStr = "Unknown"
	}

	// 输出状态信息
	fmt.Fprintf(output, "Service Plugin: %s (%s)\n", servicePlugin.Name(), servicePlugin.ID())
	fmt.Fprintf(output, "Version: %s\n", servicePlugin.Version())
	fmt.Fprintf(output, "State: %s\n", stateStr)

	return nil
}

// configService 配置服务
func (p *PluginManagerPlugin) configService(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: config <plugin_id> [config_file]")
	}

	pluginID := args[0]

	// 获取插件
	plug, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件类型
	if plug.Type() != plugin.ServicePlugin {
		return fmt.Errorf("plugin %s is not a service plugin", pluginID)
	}

	// 如果没有指定配置文件，显示当前配置
	if len(args) == 1 {
		// 读取当前配置
		configPath := filepath.Join(p.configDir, pluginID+".yml")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(output, "No configuration file found for plugin %s\n", pluginID)
				return nil
			}
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// 显示配置
		fmt.Fprintf(output, "Current configuration for plugin %s:\n", pluginID)
		fmt.Fprintln(output, string(configData))
		return nil
	}

	// 读取新配置
	configFile := args[1]
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 保存新配置
	destPath := filepath.Join(p.configDir, pluginID+".yml")
	if err := os.WriteFile(destPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Fprintf(output, "Configuration for plugin %s updated successfully\n", pluginID)
	fmt.Fprintln(output, "Restart the plugin to apply the new configuration")
	return nil
}
