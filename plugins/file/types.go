package main

import (
	"os"
	"time"

	"github.com/sorc/tcpserver/pkg/plugin"
)

// FileTransferPlugin 文件传输插件
type FileTransferPlugin struct {
	*plugin.BaseCommandPlugin
	baseDir string
}

// Config 插件配置
type Config struct {
	BaseDir string `yaml:"base_dir"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
	MD5     string      `json:"md5,omitempty"`
}

// UploadRequest 上传请求
type UploadRequest struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	MD5        string `json:"md5,omitempty"`
	Compress   bool   `json:"compress,omitempty"`
	Resume     bool   `json:"resume,omitempty"`
	Decompress bool   `json:"decompress,omitempty"`
	Offset     int64  `json:"offset,omitempty"`
	IsDir      bool   `json:"is_dir,omitempty"`
	Compressed bool   `json:"compressed,omitempty"`
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	Path       string `json:"path"`
	Compress   bool   `json:"compress,omitempty"`
	Offset     int64  `json:"offset,omitempty"`
	IsDir      bool   `json:"is_dir,omitempty"`
	Decompress bool   `json:"decompress,omitempty"`
}
