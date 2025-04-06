package plugin

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// BasePlugin 提供插件基础实现
type BasePlugin struct {
	id      string
	name    string
	version string
	pType   PluginType
	state   PluginState
	mu      sync.RWMutex
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(id, name, version string, pType PluginType) *BasePlugin {
	return &BasePlugin{
		id:      id,
		name:    name,
		version: version,
		pType:   pType,
		state:   Disabled,
	}
}

// ID 返回插件唯一标识
func (p *BasePlugin) ID() string {
	return p.id
}

// Name 返回插件名称
func (p *BasePlugin) Name() string {
	return p.name
}

// Version 返回插件版本
func (p *BasePlugin) Version() string {
	return p.version
}

// Type 返回插件类型
func (p *BasePlugin) Type() PluginType {
	return p.pType
}

// State 返回插件当前状态
func (p *BasePlugin) State() PluginState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// SetState 设置插件状态
func (p *BasePlugin) SetState(state PluginState) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = state
	return nil
}

// Init 初始化插件（基础实现）
func (p *BasePlugin) Init(ctx context.Context, config []byte) error {
	return nil
}

// Cleanup 清理插件资源（基础实现）
func (p *BasePlugin) Cleanup() error {
	return nil
}

// BaseServicePlugin 提供服务类插件基础实现
type BaseServicePlugin struct {
	*BasePlugin
}

// NewBaseServicePlugin 创建服务类基础插件
func NewBaseServicePlugin(id, name, version string) *BaseServicePlugin {
	return &BaseServicePlugin{
		BasePlugin: NewBasePlugin(id, name, version, ServicePlugin),
	}
}

// Start 启动服务（基础实现）
func (p *BaseServicePlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = Running
	return nil
}

// Stop 停止服务（基础实现）
func (p *BaseServicePlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = Enabled
	return nil
}

// Restart 重启服务（基础实现）
func (p *BaseServicePlugin) Restart(ctx context.Context) error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start(ctx)
}

// Pause 暂停服务（基础实现）
func (p *BaseServicePlugin) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = Paused
	return nil
}

// Resume 恢复服务（基础实现）
func (p *BaseServicePlugin) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = Running
	return nil
}

// BaseCommandPlugin 提供命令类插件基础实现
type BaseCommandPlugin struct {
	*BasePlugin
	cmdType CommandType
}

// NewBaseCommandPlugin 创建命令类基础插件
func NewBaseCommandPlugin(id, name, version string, cmdType CommandType) *BaseCommandPlugin {
	return &BaseCommandPlugin{
		BasePlugin: NewBasePlugin(id, name, version, CommandPlugin),
		cmdType:    cmdType,
	}
}

// CommandType 返回命令类型
func (p *BaseCommandPlugin) CommandType() CommandType {
	return p.cmdType
}

// Execute 执行命令（基础实现）
func (p *BaseCommandPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	return fmt.Errorf("command execution not implemented")
}

// GetCommands 获取支持的命令列表（基础实现）
func (p *BaseCommandPlugin) GetCommands() []string {
	return []string{}
}
