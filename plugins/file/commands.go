package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Init 初始化插件
func (p *FileTransferPlugin) Init(ctx context.Context, configBytes []byte) error {
	if err := p.BaseCommandPlugin.Init(ctx, configBytes); err != nil {
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
	if config.BaseDir == "" {
		config.BaseDir = "files"
	}

	p.baseDir = config.BaseDir

	// 创建基础目录
	if err := os.MkdirAll(p.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	return nil
}

// GetCommands 获取支持的命令列表
func (p *FileTransferPlugin) GetCommands() []string {
	return []string{
		"upload",
		"download",
		"list",
		"delete",
		"mkdir",
	}
}

// Execute 执行命令
func (p *FileTransferPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "upload":
		return p.upload(ctx, cmdArgs, input, output)
	case "download":
		return p.download(ctx, cmdArgs, input, output)
	case "list":
		return p.list(ctx, cmdArgs, output)
	case "delete":
		return p.delete(ctx, cmdArgs, output)
	case "mkdir":
		return p.mkdir(ctx, cmdArgs, output)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// list 列出文件
func (p *FileTransferPlugin) list(ctx context.Context, args []string, output io.Writer) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// 构建完整路径
	fullPath := filepath.Join(p.baseDir, path)

	// 检查路径是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// 如果是文件，直接返回文件信息
	if !fileInfo.IsDir() {
		// 计算MD5
		md5Sum, err := p.calculateMD5(fullPath)
		if err != nil {
			return fmt.Errorf("failed to calculate MD5: %w", err)
		}

		fileInfoJson, err := json.Marshal([]FileInfo{
			{
				Path:    path,
				Size:    fileInfo.Size(),
				Mode:    fileInfo.Mode(),
				ModTime: fileInfo.ModTime(),
				IsDir:   false,
				MD5:     md5Sum,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to marshal file info: %w", err)
		}
		fmt.Fprintf(output, "%s\n", fileInfoJson)
		return nil
	}

	// 读取目录内容
	files, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// 构建文件信息列表
	fileInfos := make([]FileInfo, 0, len(files))
	for _, file := range files {
		filePath := filepath.Join(path, file.Name())

		// 获取文件信息
		info, err := file.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		fileInfo := FileInfo{
			Path:    filePath,
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		// 如果是文件，计算MD5
		if !info.IsDir() {
			md5Sum, err := p.calculateMD5(filepath.Join(p.baseDir, filePath))
			if err != nil {
				return fmt.Errorf("failed to calculate MD5 for %s: %w", filePath, err)
			}
			fileInfo.MD5 = md5Sum
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	// 序列化文件信息
	fileInfosJson, err := json.Marshal(fileInfos)
	if err != nil {
		return fmt.Errorf("failed to marshal file infos: %w", err)
	}

	fmt.Fprintf(output, "%s\n", fileInfosJson)
	return nil
}

// delete 删除文件
func (p *FileTransferPlugin) delete(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: delete <path>")
	}

	path := args[0]

	// 构建完整路径
	fullPath := filepath.Join(p.baseDir, path)

	// 检查路径是否存在
	_, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// 删除文件或目录
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Fprintf(output, "{\"success\":true,\"path\":\"%s\"}\n", path)
	return nil
}

// mkdir 创建目录
func (p *FileTransferPlugin) mkdir(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mkdir <path>")
	}

	path := args[0]

	// 构建完整路径
	fullPath := filepath.Join(p.baseDir, path)

	// 创建目录
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Fprintf(output, "{\"success\":true,\"path\":\"%s\"}\n", path)
	return nil
}
