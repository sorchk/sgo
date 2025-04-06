package main

import (
	"context"
	"fmt"

	"github.com/sorc/tcpserver/pkg/plugin"
	"gopkg.in/yaml.v3"
)

// CreateServicePlugin 创建服务类插件实例
func CreateServicePlugin() plugin.IServicePlugin {
	return &ProxyPlugin{
		BaseServicePlugin: plugin.NewBaseServicePlugin("proxy", "Proxy Service", "1.0.0"),
	}
}

// CreatePlugin 创建插件实例
func CreatePlugin() plugin.Plugin {
	return CreateServicePlugin()
}

// Init 初始化插件
func (p *ProxyPlugin) Init(ctx context.Context, configBytes []byte) error {
	if err := p.BaseServicePlugin.Init(ctx, configBytes); err != nil {
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
	if config.HTTPAddr == "" {
		config.HTTPAddr = ":8080"
	}
	if config.SocksAddr == "" {
		config.SocksAddr = ":1080"
	}

	p.config = config

	// 创建代理服务
	p.httpProxy = &HTTPProxy{
		addr: config.HTTPAddr,
	}
	p.socksProxy = &SocksProxy{
		addr: config.SocksAddr,
	}

	return nil
}

func main() {}
