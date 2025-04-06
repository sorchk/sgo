package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTSecret JWT密钥
var JWTSecret = []byte("your-secret-key")

// Claims JWT声明
type Claims struct {
	ClientID string `json:"client_id"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(clientID string) (string, int64, error) {
	// 设置过期时间
	expireTime := time.Now().Add(24 * time.Hour)
	expiresAt := expireTime.Unix()

	// 创建声明
	claims := Claims{
		ClientID: clientID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "tcp-network-service",
		},
	}

	// 创建令牌
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(JWTSecret)

	return token, expiresAt, err
}

// ParseToken 解析JWT令牌
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		// 解析令牌
		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将客户端ID存储在上下文中
		c.Set("clientID", claims.ClientID)
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
