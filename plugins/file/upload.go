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

// upload 上传文件
func (p *FileTransferPlugin) upload(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	// 支持两种方式：
	// 1. 旧的JSON格式：upload <request_json>
	// 2. 新的参数格式：upload <local_path> <remote_path> [--compress] [--overwrite]

	var req UploadRequest

	// 检查是否是JSON格式
	if len(args) == 1 && strings.HasPrefix(args[0], "{") {
		// 解析JSON请求
		if err := json.Unmarshal([]byte(args[0]), &req); err != nil {
			return fmt.Errorf("failed to parse upload request: %w", err)
		}
	} else {
		// 使用新的参数格式
		if len(args) < 2 {
			return fmt.Errorf("usage: upload <local_path> <remote_path> [--compress] [--overwrite]")
		}

		localPath := args[0]
		remotePath := args[1]

		// 检查文件是否存在
		fileInfo, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("local file not found: %w", err)
		}

		// 设置请求参数
		req.Path = remotePath
		req.Size = fileInfo.Size()

		// 处理可选参数
		for i := 2; i < len(args); i++ {
			switch args[i] {
			case "--compress":
				req.Compress = true
			case "--overwrite":
				// 默认就是覆盖，这里仅为了兼容参数
			case "--resume":
				req.Resume = true
			case "--decompress":
				req.Decompress = true
			default:
				return fmt.Errorf("unknown option: %s", args[i])
			}
		}

		// 检查是否是目录
		if fileInfo.IsDir() {
			// 如果需要压缩
			if req.Compress {
				// 创建临时文件
				tempFile, err := os.CreateTemp("", "upload-*.zip")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				tempPath := tempFile.Name()
				tempFile.Close()

				// 压缩目录
				fmt.Fprintf(output, "Compressing directory %s...\n", localPath)
				if err := p.compressDirectory(localPath, tempPath); err != nil {
					os.Remove(tempPath)
					return fmt.Errorf("failed to compress directory: %w", err)
				}

				// 更新源路径
				localPath = tempPath
				defer os.Remove(tempPath)

				// 获取新文件信息
				fileInfo, err = os.Stat(localPath)
				if err != nil {
					return fmt.Errorf("failed to get compressed file info: %w", err)
				}

				// 更新请求参数
				req.Size = fileInfo.Size()
				req.IsDir = false
				req.Compressed = true

				// 设置输入源
				file, err := os.Open(localPath)
				if err != nil {
					return fmt.Errorf("failed to open compressed file: %w", err)
				}
				defer file.Close()

				// 计算MD5
				md5Hash := md5.New()
				if _, err := io.Copy(md5Hash, file); err != nil {
					return fmt.Errorf("failed to calculate MD5: %w", err)
				}
				req.MD5 = hex.EncodeToString(md5Hash.Sum(nil))

				// 重置文件指针
				if _, err := file.Seek(0, io.SeekStart); err != nil {
					return fmt.Errorf("failed to reset file pointer: %w", err)
				}

				// 使用文件作为输入
				input = file
			} else {
				// 如果是目录但不压缩，创建远程目录
				remoteDir := filepath.Join(p.baseDir, req.Path)
				if err := os.MkdirAll(remoteDir, 0755); err != nil {
					return fmt.Errorf("failed to create remote directory: %w", err)
				}

				// 递归上传目录中的文件
				fmt.Fprintf(output, "Uploading directory %s to %s...\n", localPath, req.Path)
				return p.uploadDirectory(ctx, localPath, req.Path, req.Compress, output)
			}
		} else {
			// 如果是普通文件
			// 设置输入源
			file, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("failed to open local file: %w", err)
			}
			defer file.Close()

			// 计算MD5
			md5Hash := md5.New()
			if _, err := io.Copy(md5Hash, file); err != nil {
				return fmt.Errorf("failed to calculate MD5: %w", err)
			}
			req.MD5 = hex.EncodeToString(md5Hash.Sum(nil))

			// 重置文件指针
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return fmt.Errorf("failed to reset file pointer: %w", err)
			}

			// 使用文件作为输入
			input = file
		}
	}

	// 构建目标路径
	destPath := filepath.Join(p.baseDir, req.Path)

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 检查是否需要断点续传
	var file *os.File
	var err error
	var offset int64 = 0

	if req.Resume {
		// 检查文件是否存在
		if _, err := os.Stat(destPath); err == nil {
			// 获取文件大小
			fileInfo, err := os.Stat(destPath)
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			offset = fileInfo.Size()

			// 打开文件进行追加
			file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file for append: %w", err)
			}
		} else {
			// 文件不存在，创建新文件
			file, err = os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
		}
	} else {
		// 创建新文件
		file, err = os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}
	defer file.Close()

	// 设置偏移量
	if req.Offset > 0 {
		offset = req.Offset
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to offset: %w", err)
		}
	}

	// 发送偏移量
	fmt.Fprintf(output, "{\"offset\":%d}\n", offset)

	// 计算需要读取的字节数
	bytesToRead := req.Size - offset
	if bytesToRead <= 0 {
		return fmt.Errorf("file already complete")
	}

	// 读取并写入文件
	var writer io.Writer = file
	md5Hash := md5.New()
	if req.MD5 != "" {
		writer = io.MultiWriter(file, md5Hash)
	}

	bytesRead, err := io.CopyN(writer, input, bytesToRead)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 验证MD5
	if req.MD5 != "" {
		calculatedMD5 := hex.EncodeToString(md5Hash.Sum(nil))
		if calculatedMD5 != req.MD5 {
			// 删除文件
			os.Remove(destPath)
			return fmt.Errorf("MD5 checksum mismatch: expected %s, got %s", req.MD5, calculatedMD5)
		}

		// 文件一致性验证成功
		fmt.Fprintf(output, "File integrity verified: MD5 checksum matches (%s)\n", calculatedMD5)
	}

	// 如果需要解压缩
	if req.Decompress {
		// 关闭文件
		file.Close()

		// 解压文件
		if err := p.decompressFile(destPath, destDir); err != nil {
			return fmt.Errorf("failed to decompress file: %w", err)
		}

		// 删除压缩文件
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove compressed file: %w", err)
		}
	}

	fmt.Fprintf(output, "{\"success\":true,\"bytes_written\":%d}\n", bytesRead+offset)
	return nil
}

// uploadDirectory 递归上传目录
func (p *FileTransferPlugin) uploadDirectory(ctx context.Context, localDir, remoteDir string, compress bool, output io.Writer) error {
	// 压缩参数在这里没有使用，因为已经在上层函数中处理了
	// 递归遍历目录
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		if relPath == "." {
			return nil
		}

		// 构建远程路径
		remotePath := filepath.Join(remoteDir, relPath)

		// 如果是目录，创建远程目录
		if info.IsDir() {
			destDir := filepath.Join(p.baseDir, remotePath)
			if err := os.MkdirAll(destDir, info.Mode()); err != nil {
				return fmt.Errorf("failed to create remote directory: %w", err)
			}
			fmt.Fprintf(output, "Created directory: %s\n", remotePath)
			return nil
		}

		// 如果是文件，上传文件
		file, err := os.Open(path)
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

		// 重置文件指针
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to reset file pointer: %w", err)
		}

		// 构建目标路径
		destPath := filepath.Join(p.baseDir, remotePath)

		// 确保目标目录存在
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// 创建目标文件
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer destFile.Close()

		// 复制文件内容
		md5Hash = md5.New()
		writer := io.MultiWriter(destFile, md5Hash)
		bytesWritten, err := io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		// 验证MD5
		calculatedMD5 := hex.EncodeToString(md5Hash.Sum(nil))
		if calculatedMD5 != md5Sum {
			// 删除文件
			os.Remove(destPath)
			return fmt.Errorf("MD5 checksum mismatch: expected %s, got %s", md5Sum, calculatedMD5)
		}

		fmt.Fprintf(output, "Uploaded file: %s (%d bytes)\n", remotePath, bytesWritten)
		return nil
	})
}
