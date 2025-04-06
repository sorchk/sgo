package main

import (
	"github.com/sorc/tcpserver/pkg/plugin"
)

// PluginManagerPlugin 插件管理插件
type PluginManagerPlugin struct {
	*plugin.BaseCommandPlugin
	pluginManager plugin.PluginManager
	pluginsDir    string
	configDir     string
}

// Config 插件配置
type Config struct {
	PluginsDir string `yaml:"plugins_dir"`
	ConfigDir  string `yaml:"config_dir"`
}
