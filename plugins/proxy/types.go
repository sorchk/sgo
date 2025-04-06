package main

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/sorc/tcpserver/pkg/plugin"
)

// ProxyPlugin 代理服务插件
type ProxyPlugin struct {
	*plugin.BaseServicePlugin
	httpProxy  *HTTPProxy
	socksProxy *SocksProxy
	config     Config
}

// HTTPProxy HTTP代理服务
type HTTPProxy struct {
	server   *http.Server
	addr     string
	listener net.Listener
	mu       sync.Mutex
}

// SocksProxy SOCKS代理服务
type SocksProxy struct {
	addr     string
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
}

// Config 插件配置
type Config struct {
	HTTPAddr  string `yaml:"http_addr"`
	SocksAddr string `yaml:"socks_addr"`
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	Type    string `json:"type"`
	Addr    string `json:"addr"`
	Running bool   `json:"running"`
}
