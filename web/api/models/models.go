package models

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AuthRequest 认证请求
type AuthRequest struct {
	ClientID string `json:"client_id"`
	Secret   string `json:"secret"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	State   string `json:"state"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
	MD5     string `json:"md5,omitempty"`
}

// TerminalInfo 终端信息
type TerminalInfo struct {
	ID        string   `json:"id"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	CreatedAt string   `json:"created_at"`
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	Type    string `json:"type"`
	Addr    string `json:"addr"`
	Running bool   `json:"running"`
}

// CommandRequest 命令请求
type CommandRequest struct {
	Plugin  string   `json:"plugin"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// CommandResponse 命令响应
type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}
