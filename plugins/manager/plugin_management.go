package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// listPlugins 列出所有插件
func (p *PluginManagerPlugin) listPlugins(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	plugins := p.pluginManager.ListPlugins()

	fmt.Fprintln(output, "Installed Plugins:")
	fmt.Fprintln(output, "ID\tName\tVersion\tType\tState")
	fmt.Fprintln(output, "----------------------------------------------------")

	for _, plugin := range plugins {
		var typeStr string
		if plugin.Type() == 0 {
			typeStr = "Service"
		} else if plugin.Type() == 1 {
			typeStr = "Command"
		} else {
			typeStr = "Unknown"
		}

		var stateStr string
		if plugin.State() == 0 {
			stateStr = "Disabled"
		} else if plugin.State() == 1 {
			stateStr = "Enabled"
		} else if plugin.State() == 2 {
			stateStr = "Running"
		} else if plugin.State() == 3 {
			stateStr = "Paused"
		} else {
			stateStr = "Unknown"
		}

		fmt.Fprintf(output, "%s\t%s\t%s\t%s\t%s\n", plugin.ID(), plugin.Name(), plugin.Version(), typeStr, stateStr)
	}

	return nil
}

// installPlugin 安装插件
func (p *PluginManagerPlugin) installPlugin(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: install <plugin_path>")
	}

	pluginPath := args[0]

	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 复制插件文件到插件目录
	pluginFileName := filepath.Base(pluginPath)
	destPath := filepath.Join(p.pluginsDir, pluginFileName)

	// 复制插件文件
	if err := copyFile(pluginPath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin file: %w", err)
	}

	// 复制配置文件（如果存在）
	metadataPath := pluginPath + ".yml"
	if _, err := os.Stat(metadataPath); err == nil {
		destMetadataPath := destPath + ".yml"
		if err := copyFile(metadataPath, destMetadataPath); err != nil {
			return fmt.Errorf("failed to copy plugin metadata: %w", err)
		}
	}

	// 加载插件
	plugin, err := p.pluginManager.LoadPlugin(destPath)
	if err != nil {
		// 清理文件
		os.Remove(destPath)
		if _, err := os.Stat(destPath + ".yml"); err == nil {
			os.Remove(destPath + ".yml")
		}
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	fmt.Fprintf(output, "Plugin %s (%s) installed successfully\n", plugin.Name(), plugin.ID())
	return nil
}

// uninstallPlugin 卸载插件
func (p *PluginManagerPlugin) uninstallPlugin(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: uninstall <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 卸载插件
	if err := p.pluginManager.UnloadPlugin(pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// 删除插件文件
	pluginPath := filepath.Join(p.pluginsDir, pluginID+".so")
	if err := os.Remove(pluginPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plugin file: %w", err)
	}

	// 删除元数据文件
	metadataPath := pluginPath + ".yml"
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plugin metadata: %w", err)
	}

	// 删除配置文件
	configPath := filepath.Join(p.configDir, pluginID+".yml")
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plugin config: %w", err)
	}

	fmt.Fprintf(output, "Plugin %s (%s) uninstalled successfully\n", plugin.Name(), plugin.ID())
	return nil
}

// enablePlugin 启用插件
func (p *PluginManagerPlugin) enablePlugin(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: enable <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 启用插件
	if err := p.pluginManager.EnablePlugin(pluginID); err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	fmt.Fprintf(output, "Plugin %s (%s) enabled successfully\n", plugin.Name(), plugin.ID())
	return nil
}

// disablePlugin 禁用插件
func (p *PluginManagerPlugin) disablePlugin(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: disable <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 禁用插件
	if err := p.pluginManager.DisablePlugin(pluginID); err != nil {
		return fmt.Errorf("failed to disable plugin: %w", err)
	}

	fmt.Fprintf(output, "Plugin %s (%s) disabled successfully\n", plugin.Name(), plugin.ID())
	return nil
}

// upgradePlugin 升级插件
func (p *PluginManagerPlugin) upgradePlugin(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: upgrade <plugin_id> <plugin_path>")
	}

	pluginID := args[0]
	pluginPath := args[1]

	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 获取旧插件信息
	oldPlugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 复制新插件文件到插件目录
	pluginFileName := filepath.Base(pluginPath)
	destPath := filepath.Join(p.pluginsDir, pluginFileName)

	// 复制插件文件
	if err := copyFile(pluginPath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin file: %w", err)
	}

	// 复制配置文件（如果存在）
	metadataPath := pluginPath + ".yml"
	if _, err := os.Stat(metadataPath); err == nil {
		destMetadataPath := destPath + ".yml"
		if err := copyFile(metadataPath, destMetadataPath); err != nil {
			return fmt.Errorf("failed to copy plugin metadata: %w", err)
		}
	}

	// 升级插件
	if err := p.pluginManager.UpgradePlugin(pluginID, destPath); err != nil {
		// 清理文件
		os.Remove(destPath)
		if _, err := os.Stat(destPath + ".yml"); err == nil {
			os.Remove(destPath + ".yml")
		}
		return fmt.Errorf("failed to upgrade plugin: %w", err)
	}

	// 获取新插件信息
	newPlugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get upgraded plugin: %w", err)
	}

	fmt.Fprintf(output, "Plugin %s upgraded from %s to %s successfully\n", pluginID, oldPlugin.Version(), newPlugin.Version())
	return nil
}

// pluginInfo 获取插件信息
func (p *PluginManagerPlugin) pluginInfo(ctx context.Context, args []string, output io.Writer) error {
	if p.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: info <plugin_id>")
	}

	pluginID := args[0]

	// 获取插件
	plugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 读取插件元数据
	metadataPath := filepath.Join(p.pluginsDir, pluginID+".so.yml")
	var metadata struct {
		ID           string   `yaml:"id"`
		Name         string   `yaml:"name"`
		Version      string   `yaml:"version"`
		Type         int      `yaml:"type"`
		Description  string   `yaml:"description"`
		Author       string   `yaml:"author"`
		Dependencies []string `yaml:"dependencies,omitempty"`
	}

	if _, err := os.Stat(metadataPath); err == nil {
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to read plugin metadata: %w", err)
		}

		if err := yaml.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("failed to parse plugin metadata: %w", err)
		}
	}

	// 输出插件信息
	fmt.Fprintf(output, "Plugin Information:\n")
	fmt.Fprintf(output, "ID: %s\n", plugin.ID())
	fmt.Fprintf(output, "Name: %s\n", plugin.Name())
	fmt.Fprintf(output, "Version: %s\n", plugin.Version())

	var typeStr string
	if plugin.Type() == 0 {
		typeStr = "Service"
	} else if plugin.Type() == 1 {
		typeStr = "Command"
	} else {
		typeStr = "Unknown"
	}
	fmt.Fprintf(output, "Type: %s\n", typeStr)

	var stateStr string
	if plugin.State() == 0 {
		stateStr = "Disabled"
	} else if plugin.State() == 1 {
		stateStr = "Enabled"
	} else if plugin.State() == 2 {
		stateStr = "Running"
	} else if plugin.State() == 3 {
		stateStr = "Paused"
	} else {
		stateStr = "Unknown"
	}
	fmt.Fprintf(output, "State: %s\n", stateStr)

	if metadata.Description != "" {
		fmt.Fprintf(output, "Description: %s\n", metadata.Description)
	}
	if metadata.Author != "" {
		fmt.Fprintf(output, "Author: %s\n", metadata.Author)
	}
	if len(metadata.Dependencies) > 0 {
		fmt.Fprintf(output, "Dependencies: %s\n", strings.Join(metadata.Dependencies, ", "))
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
