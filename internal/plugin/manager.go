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

	// 获取插件实例
	symPlugin, err := plug.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export 'Plugin' symbol: %w", err)
	}

	// 将插件转换为插件接口
	p, ok := symPlugin.(Plugin)
	if !ok {
		return nil, ErrPluginTypeMismatch
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
	if err := p.Init(pm.ctx, configBytes); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 存储插件
	pm.plugins[metadata.ID] = p

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
	if p.Type() == PluginTypeService {
		if sp, ok := p.(ServicePluginInterface); ok && sp.State() == Running {
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
func (pm *DefaultPluginManager) GetServicePlugin(id string) (ServicePluginInterface, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[id]
	if !exists {
		return nil, ErrPluginNotFound
	}

	if p.Type() != PluginTypeService {
		return nil, ErrPluginTypeMismatch
	}

	// 尝试将插件转换为服务类插件接口
	sp, ok := p.(ServicePluginInterface)
	if !ok {
		// 如果转换失败，尝试使用类型断言
		return p.(ServicePluginInterface), nil
	}

	return sp, nil
}

// GetCommandPlugin 获取命令类插件
func (pm *DefaultPluginManager) GetCommandPlugin(id string) (CommandPluginInterface, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[id]
	if !exists {
		return nil, ErrPluginNotFound
	}

	if p.Type() != PluginTypeCommand {
		return nil, ErrPluginTypeMismatch
	}

	// 尝试将插件转换为命令类插件接口
	cp, ok := p.(CommandPluginInterface)
	if !ok {
		// 如果转换失败，尝试使用类型断言
		return p.(CommandPluginInterface), nil
	}

	return cp, nil
}
