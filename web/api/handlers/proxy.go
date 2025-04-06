package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/models"
)

// GetProxyStatus 获取代理状态
func GetProxyStatus(c *gin.Context) {
	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行proxy status命令获取代理状态
	output, err := tcpClient.ExecuteCommand("proxy", "status", []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get proxy status: %v", err),
		})
		return
	}

	// 解析JSON输出
	var proxyStatus []models.ProxyStatus
	if err := json.Unmarshal([]byte(output), &proxyStatus); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse proxy status: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    proxyStatus,
	})
}

// StartProxy 启动代理
func StartProxy(c *gin.Context) {
	proxyType := c.Param("type")
	if proxyType == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Proxy type is required",
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

	// 执行proxy start命令启动代理
	output, err := tcpClient.ExecuteCommand("proxy", "start", []string{proxyType})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to start proxy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("%s proxy started successfully", proxyType),
		Data: gin.H{
			"output": output,
		},
	})
}

// StopProxy 停止代理
func StopProxy(c *gin.Context) {
	proxyType := c.Param("type")
	if proxyType == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Proxy type is required",
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

	// 执行proxy stop命令停止代理
	output, err := tcpClient.ExecuteCommand("proxy", "stop", []string{proxyType})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to stop proxy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("%s proxy stopped successfully", proxyType),
		Data: gin.H{
			"output": output,
		},
	})
}
