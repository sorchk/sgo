package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/models"
)

// ListTerminals 列出终端
func ListTerminals(c *gin.Context) {
	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行terminal list命令获取终端列表
	output, err := tcpClient.ExecuteCommand("terminal", "list", []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list terminals: %v", err),
		})
		return
	}

	// 解析JSON输出
	var terminals []models.TerminalInfo
	if err := json.Unmarshal([]byte(output), &terminals); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse terminal list: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    terminals,
	})
}

// CreateTerminal 创建终端
func CreateTerminal(c *gin.Context) {
	var req struct {
		ID      string   `json:"id"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Terminal ID is required",
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

	// 构建创建终端的请求JSON
	createReq := map[string]interface{}{
		"id":      req.ID,
		"command": req.Command,
		"args":    req.Args,
	}
	createReqJSON, err := json.Marshal(createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		})
		return
	}

	// 执行terminal create命令创建终端
	output, err := tcpClient.ExecuteCommand("terminal", "create", []string{string(createReqJSON)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create terminal: %v", err),
		})
		return
	}

	// 解析JSON输出
	var terminalInfo models.TerminalInfo
	if err := json.Unmarshal([]byte(output), &terminalInfo); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse terminal info: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Terminal %s created successfully", req.ID),
		Data:    terminalInfo,
	})
}

// KillTerminal 终止终端
func KillTerminal(c *gin.Context) {
	terminalID := c.Param("id")
	if terminalID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Terminal ID is required",
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

	// 执行terminal kill命令终止终端
	output, err := tcpClient.ExecuteCommand("terminal", "kill", []string{terminalID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to kill terminal: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Terminal %s killed successfully", terminalID),
		Data: gin.H{
			"output": output,
		},
	})
}

// WriteToTerminal 向终端写入数据
func WriteToTerminal(c *gin.Context) {
	var req struct {
		ID   string `json:"id"`
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Terminal ID is required",
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

	// 构建写入终端的请求JSON
	writeReq := map[string]interface{}{
		"id":   req.ID,
		"data": req.Data,
	}
	writeReqJSON, err := json.Marshal(writeReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		})
		return
	}

	// 执行terminal write命令向终端写入数据
	output, err := tcpClient.ExecuteCommand("terminal", "write", []string{string(writeReqJSON)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write to terminal: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"output": output,
		},
	})
}

// ReadFromTerminal 从终端读取数据
func ReadFromTerminal(c *gin.Context) {
	terminalID := c.Param("id")
	if terminalID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Terminal ID is required",
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

	// 执行terminal read命令从终端读取数据
	output, err := tcpClient.ExecuteCommand("terminal", "read", []string{terminalID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read from terminal: %v", err),
		})
		return
	}

	// 解析JSON输出
	var readResult struct {
		ID     string `json:"id"`
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
	}
	if err := json.Unmarshal([]byte(output), &readResult); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse terminal output: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    readResult,
	})
}
