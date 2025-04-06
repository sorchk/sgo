package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/middleware"
	"github.com/sorc/tcpserver/web/api/models"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
	Name   string `json:"name"`
}

// Clients 客户端列表
var Clients = []ClientConfig{
	{
		ID:     "client1",
		Secret: "this_is_a_very_long_secret_key_that_is_more_than_16_characters",
		Name:   "Default Client",
	},
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// 验证客户端凭据
	var validClient bool
	for _, client := range Clients {
		if client.ID == req.ClientID && client.Secret == req.Secret {
			validClient = true
			break
		}
	}

	if !validClient {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid client credentials",
		})
		return
	}

	// 生成JWT令牌
	token, expires, err := middleware.GenerateToken(req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.AuthResponse{
			Token:   token,
			Expires: expires,
		},
	})
}

// ValidateToken 验证令牌
func ValidateToken(c *gin.Context) {
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"client_id": clientID,
		},
	})
}
