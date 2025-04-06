package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/models"
)

// ListFiles 列出文件
func ListFiles(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "."
	}

	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行file list命令获取文件列表
	output, err := tcpClient.ExecuteCommand("file", "list", []string{path})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list files: %v", err),
		})
		return
	}

	// 解析JSON输出
	var files []models.FileInfo
	if err := json.Unmarshal([]byte(output), &files); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse file list: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    files,
	})
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	// 获取表单参数
	remotePath := c.PostForm("remote_path")
	compress := c.PostForm("compress") == "true"
	overwrite := c.PostForm("overwrite") == "true"

	if remotePath == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Remote path is required",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get file: %v", err),
		})
		return
	}
	defer file.Close()

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 复制文件内容到临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save file: %v", err),
		})
		return
	}

	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 构建上传命令参数
	args := []string{tempFile.Name(), remotePath}
	if compress {
		args = append(args, "--compress")
	}
	if overwrite {
		args = append(args, "--overwrite")
	}

	// 执行file upload命令上传文件
	output, err := tcpClient.ExecuteCommand("file", "upload", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to upload file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("File uploaded successfully to %s", remotePath),
		Data: gin.H{
			"output": output,
		},
	})
}

// DownloadFile 下载文件
func DownloadFile(c *gin.Context) {
	remotePath := c.Query("path")
	if remotePath == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Remote path is required",
		})
		return
	}

	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "download-*"+filepath.Ext(remotePath))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 执行file download命令下载文件
	_, err = tcpClient.ExecuteCommand("file", "download", []string{remotePath, tempFile.Name()})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to download file: %v", err),
		})
		return
	}

	// 打开临时文件
	file, err := os.Open(tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open downloaded file: %v", err),
		})
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get file info: %v", err),
		})
		return
	}

	// 设置响应头
	filename := filepath.Base(remotePath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 发送文件内容
	c.File(tempFile.Name())
}

// DeleteFile 删除文件
func DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Path is required",
		})
		return
	}

	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行file delete命令删除文件
	output, err := tcpClient.ExecuteCommand("file", "delete", []string{path})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("File %s deleted successfully", path),
		Data: gin.H{
			"output": output,
		},
	})
}

// MakeDirectory 创建目录
func MakeDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.Path == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Path is required",
		})
		return
	}

	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行file mkdir命令创建目录
	output, err := tcpClient.ExecuteCommand("file", "mkdir", []string{req.Path})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Directory %s created successfully", req.Path),
		Data: gin.H{
			"output": output,
		},
	})
}
