package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
)

// calculateMD5 计算文件的MD5哈希值
func (p *FileTransferPlugin) calculateMD5(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建MD5哈希器
	md5Hash := md5.New()

	// 计算MD5
	if _, err := io.Copy(md5Hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	// 返回十六进制字符串
	return hex.EncodeToString(md5Hash.Sum(nil)), nil
}

// parseInt64 将字符串转换为int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
