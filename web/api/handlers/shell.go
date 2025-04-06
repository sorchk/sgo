package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/models"
)

// ExecuteShellCommand 执行Shell命令
func ExecuteShellCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.Command == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Command is required",
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

	// 执行shell exec命令执行Shell命令
	output, err := tcpClient.ExecuteCommand("shell", "exec", []string{req.Command})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Command execution failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"command": req.Command,
			"output":  output,
		},
	})
}
