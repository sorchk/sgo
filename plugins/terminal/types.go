package main

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/sorc/tcpserver/pkg/plugin"
)

// TerminalPlugin 终端管理插件
type TerminalPlugin struct {
	*plugin.BaseCommandPlugin
	terminals   map[string]*Terminal
	terminalsMu sync.RWMutex
	workingDir  string
}

// Terminal 终端实例
type Terminal struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	CreatedAt time.Time `json:"created_at"`
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	ctx       context.Context
	cancel    context.CancelFunc
}

// Config 插件配置
type Config struct {
	WorkingDir string `yaml:"working_dir"`
}

// CreateTerminalRequest 终端创建请求
type CreateTerminalRequest struct {
	ID      string   `json:"id"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// TerminalDataRequest 终端数据请求
type TerminalDataRequest struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// ResizeTerminalRequest 终端大小调整请求
type ResizeTerminalRequest struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
