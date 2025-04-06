package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// GetCommands 获取支持的命令列表
func (p *ProxyPlugin) GetCommands() []string {
	return []string{
		"status",
		"start",
		"stop",
	}
}

// Execute 执行命令
func (p *ProxyPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "status":
		return p.getStatus(ctx, output)
	case "start":
		return p.startProxy(ctx, cmdArgs, output)
	case "stop":
		return p.stopProxy(ctx, cmdArgs, output)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// Start 启动服务
func (p *ProxyPlugin) Start(ctx context.Context) error {
	// 启动所有代理服务
	if err := p.httpProxy.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP proxy: %w", err)
	}

	if err := p.socksProxy.Start(ctx); err != nil {
		p.httpProxy.Stop()
		return fmt.Errorf("failed to start SOCKS proxy: %w", err)
	}

	return p.BaseServicePlugin.Start(ctx)
}

// Stop 停止服务
func (p *ProxyPlugin) Stop() error {
	// 停止所有代理服务
	p.httpProxy.Stop()
	p.socksProxy.Stop()

	return p.BaseServicePlugin.Stop()
}

// getStatus 获取代理状态
func (p *ProxyPlugin) getStatus(ctx context.Context, output io.Writer) error {
	// 获取各代理服务状态
	status := []ProxyStatus{
		{
			Type:    "HTTP",
			Addr:    p.config.HTTPAddr,
			Running: p.httpProxy.IsRunning(),
		},
		{
			Type:    "SOCKS",
			Addr:    p.config.SocksAddr,
			Running: p.socksProxy.IsRunning(),
		},
	}

	// 序列化状态
	statusJson, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	fmt.Fprintf(output, "%s\n", statusJson)
	return nil
}

// startProxy 启动指定代理服务
func (p *ProxyPlugin) startProxy(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: start <proxy_type>")
	}

	proxyType := strings.ToUpper(args[0])

	var err error
	switch proxyType {
	case "HTTP":
		err = p.httpProxy.Start(ctx)
	case "SOCKS":
		err = p.socksProxy.Start(ctx)
	default:
		return fmt.Errorf("unknown proxy type: %s", proxyType)
	}

	if err != nil {
		return fmt.Errorf("failed to start %s proxy: %w", proxyType, err)
	}

	fmt.Fprintf(output, "{\"success\":true,\"type\":\"%s\"}\n", proxyType)
	return nil
}

// stopProxy 停止指定代理服务
func (p *ProxyPlugin) stopProxy(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stop <proxy_type>")
	}

	proxyType := strings.ToUpper(args[0])

	switch proxyType {
	case "HTTP":
		p.httpProxy.Stop()
	case "SOCKS":
		p.socksProxy.Stop()
	default:
		return fmt.Errorf("unknown proxy type: %s", proxyType)
	}

	fmt.Fprintf(output, "{\"success\":true,\"type\":\"%s\"}\n", proxyType)
	return nil
}
