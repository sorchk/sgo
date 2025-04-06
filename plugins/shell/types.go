package main

import (
	"github.com/sorc/tcpserver/pkg/plugin"
)

// ShellPlugin Shell执行插件
type ShellPlugin struct {
	*plugin.BaseCommandPlugin
	allowedCommands []string
	workingDir      string
}

// Config 插件配置
type Config struct {
	AllowedCommands []string `yaml:"allowed_commands"`
	WorkingDir      string   `yaml:"working_dir"`
}
