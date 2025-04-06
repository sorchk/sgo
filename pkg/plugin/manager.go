package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginAlreadyExists = errors.New("plugin already exists")
	ErrPluginTypeMismatch  = errors.New("plugin type mismatch")
	ErrPluginDisabled      = errors.New("plugin is disabled")
	ErrPluginEnabled       = errors.New("plugin is already enabled")
	ErrInvalidPluginFile   = errors.New("invalid plugin file")
)

// PluginManager 定义插件管理器接口
type PluginManager interface {
	// RegisterPlugin 注册内建插件
	RegisterPlugin(p Plugin) error
	// LoadPlugin 加载插件
	LoadPlugin(path string) (Plugin, error)
	// UnloadPlugin 卸载插件
	UnloadPlugin(id string) error
	// EnablePlugin 启用插件
	EnablePlugin(id string) error
	// DisablePlugin 禁用插件
	DisablePlugin(id string) error
	// UpgradePlugin 升级插件
	UpgradePlugin(id string, path string) error
	// GetPlugin 获取插件
	GetPlugin(id string) (Plugin, error)
	// ListPlugins 列出所有插件
	ListPlugins() []Plugin
	// GetServicePlugin 获取服务类插件
	GetServicePlugin(id string) (IServicePlugin, error)
	// GetCommandPlugin 获取命令类插件
	GetCommandPlugin(id string) (ICommandPlugin, error)
}

// DefaultPluginManager 默认插件管理器实现
type DefaultPluginManager struct {
	plugins    map[string]Plugin
	pluginsDir string
	configDir  string
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager(pluginsDir, configDir string) PluginManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DefaultPluginManager{
		plugins:    make(map[string]Plugin),
		pluginsDir: pluginsDir,
		configDir:  configDir,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// RegisterPlugin 注册内建插件
func (pm *DefaultPluginManager) RegisterPlugin(p Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := pm.plugins[p.ID()]; exists {
		return ErrPluginAlreadyExists
	}

	// 读取插件配置
	configPath := filepath.Join(pm.configDir, p.ID()+".yml")
	var configBytes []byte
	if _, err := os.Stat(configPath); err == nil {
		configBytes, err = os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read plugin config: %w", err)
		}
	}

	// 初始化插件
	if err := p.Init(pm.ctx, configBytes); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 存储插件
	pm.plugins[p.ID()] = p

	return nil
}

// LoadPlugin 加载插件
func (pm *DefaultPluginManager) LoadPlugin(path string) (Plugin, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查插件文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}

	// 读取插件元数据
	metadataPath := filepath.Join(filepath.Dir(path), filepath.Base(path)+".yml")
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	var metadata PluginMetadata
	if err := yaml.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	// 检查插件是否已存在
	if _, exists := pm.plugins[metadata.ID]; exists {
		return nil, ErrPluginAlreadyExists
	}

	// 加载插件
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// 获取插件工厂函数
	var factory interface{}
	var p Plugin

	switch metadata.Type {
	case ServicePlugin:
		// 尝试获取服务类插件工厂
		factory, err = plug.Lookup("CreateServicePlugin")
		if err != nil {
			// 尝试获取通用插件工厂
			factory, err = plug.Lookup("CreatePlugin")
			if err != nil {
				return nil, fmt.Errorf("plugin does not export required factory function: %w", err)
			}
		}

		// 调用工厂函数创建插件实例
		if createFunc, ok := factory.(func() IServicePlugin); ok {
			p = createFunc()
		} else if createFunc, ok := factory.(func() Plugin); ok {
			p = createFunc()
			// 验证插件类型
			if p.Type() != ServicePlugin {
				return nil, ErrPluginTypeMismatch
			}
		} else {
			return nil, ErrPluginTypeMismatch
		}

	case CommandPlugin:
		// 尝试获取命令类插件工厂
		factory, err = plug.Lookup("CreateCommandPlugin")
		if err != nil {
			// 尝试获取通用插件工厂
			factory, err = plug.Lookup("CreatePlugin")
			if err != nil {
				return nil, fmt.Errorf("plugin does not export required factory function: %w", err)
			}
		}

		// 调用工厂函数创建插件实例
		if createFunc, ok := factory.(func() ICommandPlugin); ok {
			p = createFunc()
		} else if createFunc, ok := factory.(func() Plugin); ok {
			p = createFunc()
			// 验证插件类型
			if p.Type() != CommandPlugin {
				return nil, ErrPluginTypeMismatch
			}
		} else {
			return nil, ErrPluginTypeMismatch
		}

	default:
		return nil, fmt.Errorf("unknown plugin type: %d", metadata.Type)
	}

	// 读取插件配置
	configPath := filepath.Join(pm.configDir, metadata.ID+".yml")
	var configBytes []byte
	if _, err := os.Stat(configPath); err == nil {
		configBytes, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read plugin config: %w", err)
		}
	}

	// 初始化插件
	// 创建上下文，并将插件管理器传递给插件
	ctx := context.WithValue(pm.ctx, "plugin_manager", pm)
	if err := p.Init(ctx, configBytes); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 存储插件
	pm.plugins[p.ID()] = p

	return p, nil
}

// UnloadPlugin 卸载插件
func (pm *DefaultPluginManager) UnloadPlugin(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[id]
	if !exists {
		return ErrPluginNotFound
	}

	// 清理插件资源
	if err := p.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin: %w", err)
	}

	// 从管理器中移除插件
	delete(pm.plugins, id)

	return nil
}

// EnablePlugin 启用插件
func (pm *DefaultPluginManager) EnablePlugin(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[id]
	if !exists {
		return ErrPluginNotFound
	}

	if p.State() == Enabled || p.State() == Running {
		return ErrPluginEnabled
	}

	return p.SetState(Enabled)
}

// DisablePlugin 禁用插件
func (pm *DefaultPluginManager) DisablePlugin(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[id]
	if !exists {
		return ErrPluginNotFound
	}

	if p.State() == Disabled {
		return ErrPluginDisabled
	}

	// 如果是服务类插件且正在运行，先停止服务
	if p.Type() == ServicePlugin {
		if sp, ok := p.(IServicePlugin); ok && sp.State() == Running {
			if err := sp.Stop(); err != nil {
				return fmt.Errorf("failed to stop service plugin: %w", err)
			}
		}
	}

	return p.SetState(Disabled)
}

// UpgradePlugin 升级插件
func (pm *DefaultPluginManager) UpgradePlugin(id string, path string) error {
	// 先卸载旧插件
	if err := pm.UnloadPlugin(id); err != nil {
		return fmt.Errorf("failed to unload old plugin: %w", err)
	}

	// 加载新插件
	_, err := pm.LoadPlugin(path)
	if err != nil {
		return fmt.Errorf("failed to load new plugin: %w", err)
	}

	return nil
}

// GetPlugin 获取插件
func (pm *DefaultPluginManager) GetPlugin(id string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[id]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return p, nil
}

// ListPlugins 列出所有插件
func (pm *DefaultPluginManager) ListPlugins() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// GetServicePlugin 获取服务类插件
func (pm *DefaultPluginManager) GetServicePlugin(id string) (IServicePlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[id]
	if !exists {
		return nil, ErrPluginNotFound
	}

	if p.Type() != ServicePlugin {
		return nil, ErrPluginTypeMismatch
	}

	sp, ok := p.(IServicePlugin)
	if !ok {
		return nil, ErrPluginTypeMismatch
	}

	return sp, nil
}

// GetCommandPlugin 获取命令类插件
func (pm *DefaultPluginManager) GetCommandPlugin(id string) (ICommandPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[id]
	if !exists {
		return nil, ErrPluginNotFound
	}

	if p.Type() != CommandPlugin {
		return nil, ErrPluginTypeMismatch
	}

	cp, ok := p.(ICommandPlugin)
	if !ok {
		return nil, ErrPluginTypeMismatch
	}

	return cp, nil
}
