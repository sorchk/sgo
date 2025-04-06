package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/models"
)

// TCPClient TCP客户端接口
type TCPClient interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	ExecuteCommand(plugin, command string, args []string) (string, error)
}

var tcpClient TCPClient

// SetTCPClient 设置TCP客户端
func SetTCPClient(client TCPClient) {
	tcpClient = client
}

// ListPlugins 获取插件列表
func ListPlugins(c *gin.Context) {
	if tcpClient == nil || !tcpClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "TCP client not connected",
		})
		return
	}

	// 执行manager list命令获取插件列表
	output, err := tcpClient.ExecuteCommand("manager", "list", []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list plugins: %v", err),
		})
		return
	}

	// 解析输出
	var plugins []models.PluginInfo
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    []models.PluginInfo{},
		})
		return
	}

	// 跳过标题行和分隔线
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			plugin := models.PluginInfo{
				ID:      fields[0],
				Name:    fields[1],
				Version: fields[2],
				Type:    fields[3],
				State:   fields[4],
			}
			plugins = append(plugins, plugin)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    plugins,
	})
}

// GetPluginInfo 获取插件信息
func GetPluginInfo(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Plugin ID is required",
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

	// 执行manager info命令获取插件信息
	output, err := tcpClient.ExecuteCommand("manager", "info", []string{pluginID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get plugin info: %v", err),
		})
		return
	}

	// 解析输出
	info := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "Plugin Information:" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			info[key] = value
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    info,
	})
}

// ExecuteCommand 执行命令
func ExecuteCommand(c *gin.Context) {
	var req models.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
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

	// 执行命令
	output, err := tcpClient.ExecuteCommand(req.Plugin, req.Command, req.Args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Command execution failed: %v", err),
			Data: models.CommandResponse{
				Success: false,
				Error:   err.Error(),
				Output:  output,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.CommandResponse{
			Success: true,
			Output:  output,
		},
	})
}

// StartPlugin 启动插件
func StartPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Plugin ID is required",
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

	// 执行manager start命令启动插件
	output, err := tcpClient.ExecuteCommand("manager", "start", []string{pluginID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to start plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Plugin %s started successfully", pluginID),
		Data: models.CommandResponse{
			Success: true,
			Output:  output,
		},
	})
}

// StopPlugin 停止插件
func StopPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Plugin ID is required",
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

	// 执行manager stop命令停止插件
	output, err := tcpClient.ExecuteCommand("manager", "stop", []string{pluginID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to stop plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Plugin %s stopped successfully", pluginID),
		Data: models.CommandResponse{
			Success: true,
			Output:  output,
		},
	})
}
