package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// download 下载文件
func (p *FileTransferPlugin) download(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	// 保存原始输出，以便我们可以向用户发送消息
	originalOutput := output

	// 创建一个变量来存储MD5值，用于文件完整性验证
	var expectedMD5 string
	// 支持两种方式：
	// 1. 旧的JSON格式：download <request_json>
	// 2. 新的参数格式：download <remote_path> <local_path> [--compress] [--offset <offset>]

	var req DownloadRequest

	// 检查是否是JSON格式
	if len(args) == 1 && strings.HasPrefix(args[0], "{") {
		// 解析JSON请求
		if err := json.Unmarshal([]byte(args[0]), &req); err != nil {
			return fmt.Errorf("failed to parse download request: %w", err)
		}
	} else {
		// 使用新的参数格式
		if len(args) < 2 {
			return fmt.Errorf("usage: download <remote_path> <local_path> [--compress] [--offset <offset>] [--recursive]")
		}

		remotePath := args[0]
		localPath := args[1]

		// 设置请求参数
		req.Path = remotePath

		// 处理可选参数
		for i := 2; i < len(args); i++ {
			switch args[i] {
			case "--compress":
				req.Compress = true
			case "--decompress":
				// 解压缩下载的文件
				req.Decompress = true
			case "--recursive":
				// 递归下载目录
				req.IsDir = true
			case "--offset":
				if i+1 < len(args) {
					offset, err := parseInt64(args[i+1])
					if err != nil {
						return fmt.Errorf("invalid offset: %w", err)
					}
					req.Offset = offset
					i++ // 跳过下一个参数
				} else {
					return fmt.Errorf("missing value for --offset")
				}
			default:
				return fmt.Errorf("unknown option: %s", args[i])
			}
		}

		// 创建本地文件
		var file *os.File
		var fileErr error

		// 检查文件是否存在
		if fileInfo, err := os.Stat(localPath); err == nil && req.Offset > 0 {
			// 如果文件存在且指定了偏移量，则以追加模式打开
			file, fileErr = os.OpenFile(localPath, os.O_WRONLY|os.O_APPEND, 0644)
			fmt.Fprintf(output, "Resuming download to %s (size: %d bytes)\n", localPath, fileInfo.Size())
		} else {
			// 否则创建新文件
			file, fileErr = os.Create(localPath)
			fmt.Fprintf(output, "Downloading to %s\n", localPath)
		}

		if fileErr != nil {
			return fmt.Errorf("failed to create local file: %w", fileErr)
		}

		// 在下载完成后关闭文件
		defer file.Close()

		// 设置输出为文件，原始输出已在函数开始保存
		output = file

		// 保存本地路径，以便在下载完成后进行验证
		_localPath := localPath

		// 修改下载完成后的处理
		defer func() {
			// 验证文件完整性
			if expectedMD5 != "" {
				// 计算下载文件的MD5
				downloadedMD5, err := p.calculateMD5(_localPath)
				if err != nil {
					fmt.Fprintf(originalOutput, "Warning: Failed to calculate MD5 for downloaded file: %v\n", err)
				} else if downloadedMD5 == expectedMD5 {
					fmt.Fprintf(originalOutput, "File integrity verified: MD5 checksum matches (%s)\n", expectedMD5)
				} else {
					fmt.Fprintf(originalOutput, "Warning: MD5 checksum mismatch: expected %s, got %s\n", expectedMD5, downloadedMD5)
				}
			}

			// 如果需要解压缩
			if req.Decompress {
				// 创建解压目录
				extractDir := _localPath + "_extracted"
				if err := os.MkdirAll(extractDir, 0755); err != nil {
					fmt.Fprintf(originalOutput, "Warning: Failed to create extract directory: %v\n", err)
				} else {
					// 解压文件
					fmt.Fprintf(originalOutput, "Extracting %s to %s...\n", _localPath, extractDir)
					if err := p.decompressFile(_localPath, extractDir); err != nil {
						fmt.Fprintf(originalOutput, "Warning: Failed to extract file: %v\n", err)
					} else {
						fmt.Fprintf(originalOutput, "Extraction completed: %s\n", extractDir)
					}
				}
			}

			fmt.Fprintf(originalOutput, "Download completed: %s\n", _localPath)
		}()
	}

	// 构建源路径
	srcPath := filepath.Join(p.baseDir, req.Path)

	// 检查文件是否存在
	fileInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 设置是否是目录
	req.IsDir = fileInfo.IsDir()

	// 如果是目录
	if fileInfo.IsDir() {
		// 如果需要压缩
		if req.Compress {
			// 创建临时文件
			tempFile, err := os.CreateTemp("", "download-*.zip")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			tempPath := tempFile.Name()
			tempFile.Close()

			// 压缩目录
			if err := p.compressDirectory(srcPath, tempPath); err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to compress directory: %w", err)
			}

			// 更新源路径
			srcPath = tempPath
			defer os.Remove(tempPath)

			// 获取新文件信息
			fileInfo, err = os.Stat(srcPath)
			if err != nil {
				return fmt.Errorf("failed to get compressed file info: %w", err)
			}
		} else if len(args) > 1 && args[1] != "" {
			// 如果是目录但不压缩，递归下载
			// 创建本地目录
			if err := os.MkdirAll(args[1], fileInfo.Mode()); err != nil {
				return fmt.Errorf("failed to create local directory: %w", err)
			}

			// 递归下载目录
			fmt.Fprintf(originalOutput, "Downloading directory %s to %s...\n", req.Path, args[1])
			return p.downloadDirectory(srcPath, args[1], originalOutput)
		}
	}

	// 打开文件
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 计算MD5
	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, file); err != nil {
		return fmt.Errorf("failed to calculate MD5: %w", err)
	}
	md5Sum := hex.EncodeToString(md5Hash.Sum(nil))

	// 设置期望的MD5值用于文件完整性验证
	expectedMD5 = md5Sum

	// 重置文件指针
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// 设置偏移量
	if req.Offset > 0 {
		if _, err := file.Seek(req.Offset, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to offset: %w", err)
		}
	}

	// 发送文件信息
	fileInfoJson, err := json.Marshal(FileInfo{
		Path:    req.Path,
		Size:    fileInfo.Size(),
		Mode:    fileInfo.Mode(),
		ModTime: fileInfo.ModTime(),
		IsDir:   fileInfo.IsDir(),
		MD5:     md5Sum,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal file info: %w", err)
	}

	// 如果使用的是新的参数格式，不将文件信息写入输出
	// 只有在使用旧的JSON格式时，才将文件信息写入输出
	if len(args) == 1 && strings.HasPrefix(args[0], "{") {
		fmt.Fprintf(output, "%s\n", fileInfoJson)
	} else if originalOutput != nil {
		// 在新格式下，将文件信息写入原始输出（用户终端）
		fmt.Fprintf(originalOutput, "File info: %s\n", fileInfoJson)
	}

	// 发送文件内容
	if _, err := io.Copy(output, file); err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}

	return nil
}

// downloadDirectory 递归下载目录
func (p *FileTransferPlugin) downloadDirectory(srcDir, destDir string, output io.Writer) error {
	// 递归遍历目录
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		if relPath == "." {
			return nil
		}

		// 构建目标路径
		destPath := filepath.Join(destDir, relPath)

		// 如果是目录，创建目录
		if info.IsDir() {
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			fmt.Fprintf(output, "Created directory: %s\n", relPath)
			return nil
		}

		// 如果是文件，复制文件
		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer srcFile.Close()

		// 计算MD5
		md5Hash := md5.New()
		if _, err := io.Copy(md5Hash, srcFile); err != nil {
			return fmt.Errorf("failed to calculate MD5: %w", err)
		}
		md5Sum := hex.EncodeToString(md5Hash.Sum(nil))

		// 重置文件指针
		if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to reset file pointer: %w", err)
		}

		// 创建目标文件
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer destFile.Close()

		// 复制文件内容
		md5Hash = md5.New()
		writer := io.MultiWriter(destFile, md5Hash)
		bytesWritten, err := io.Copy(writer, srcFile)
		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		// 验证MD5
		calculatedMD5 := hex.EncodeToString(md5Hash.Sum(nil))
		if calculatedMD5 != md5Sum {
			// 删除文件
			os.Remove(destPath)
			return fmt.Errorf("MD5 checksum mismatch: expected %s, got %s", md5Sum, calculatedMD5)
		}

		fmt.Fprintf(output, "Downloaded file: %s (%d bytes)\n", relPath, bytesWritten)
		return nil
	})
}
