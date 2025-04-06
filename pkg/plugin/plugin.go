package plugin

import (
	"context"
	"io"
)

// PluginType 定义插件类型
type PluginType int

const (
	// ServicePlugin 服务类插件
	ServicePlugin PluginType = iota
	// CommandPlugin 命令类插件
	CommandPlugin
)

// CommandType 定义命令类型
type CommandType int

const (
	// OneTimeCommand 一次性命令
	OneTimeCommand CommandType = iota
	// InteractiveCommand 交互式命令
	InteractiveCommand
)

// PluginState 定义插件状态
type PluginState int

const (
	// Disabled 禁用状态
	Disabled PluginState = iota
	// Enabled 启用状态
	Enabled
	// Running 运行中状态（仅适用于服务类插件）
	Running
	// Paused 暂停状态（仅适用于服务类插件）
	Paused
)

// Plugin 定义插件接口
type Plugin interface {
	// ID 返回插件唯一标识
	ID() string
	// Name 返回插件名称
	Name() string
	// Version 返回插件版本
	Version() string
	// Type 返回插件类型
	Type() PluginType
	// State 返回插件当前状态
	State() PluginState
	// SetState 设置插件状态
	SetState(state PluginState) error
	// Init 初始化插件
	Init(ctx context.Context, config []byte) error
	// Cleanup 清理插件资源
	Cleanup() error
}

// IServicePlugin 定义服务类插件接口
type IServicePlugin interface {
	Plugin
	// Start 启动服务
	Start(ctx context.Context) error
	// Stop 停止服务
	Stop() error
	// Restart 重启服务
	Restart(ctx context.Context) error
	// Pause 暂停服务
	Pause() error
	// Resume 恢复服务
	Resume() error
}

// ICommandPlugin 定义命令类插件接口
type ICommandPlugin interface {
	Plugin
	// CommandType 返回命令类型
	CommandType() CommandType
	// Execute 执行命令
	Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error
	// GetCommands 获取支持的命令列表
	GetCommands() []string
}

// PluginMetadata 定义插件元数据
type PluginMetadata struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	Version      string     `yaml:"version"`
	Type         PluginType `yaml:"type"`
	Description  string     `yaml:"description"`
	Author       string     `yaml:"author"`
	Dependencies []string   `yaml:"dependencies,omitempty"`
}

// PluginFactory 定义插件工厂函数类型
type PluginFactory func() Plugin

// ServicePluginFactory 定义服务类插件工厂函数类型
type ServicePluginFactory func() IServicePlugin

// CommandPluginFactory 定义命令类插件工厂函数类型
type CommandPluginFactory func() ICommandPlugin
