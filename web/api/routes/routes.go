package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sorc/tcpserver/web/api/handlers"
	"github.com/sorc/tcpserver/web/api/middleware"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 添加中间件
	r.Use(middleware.CORS())

	// 公共路由
	r.POST("/api/auth/login", handlers.Login)

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(middleware.JWTAuth())
	{
		// 认证相关
		api.GET("/auth/validate", handlers.ValidateToken)

		// 插件管理
		api.GET("/plugins", handlers.ListPlugins)
		api.GET("/plugins/:id", handlers.GetPluginInfo)
		api.POST("/plugins/:id/start", handlers.StartPlugin)
		api.POST("/plugins/:id/stop", handlers.StopPlugin)
		api.POST("/command", handlers.ExecuteCommand)

		// 文件管理
		api.GET("/files", handlers.ListFiles)
		api.POST("/files/upload", handlers.UploadFile)
		api.GET("/files/download", handlers.DownloadFile)
		api.DELETE("/files", handlers.DeleteFile)
		api.POST("/files/mkdir", handlers.MakeDirectory)

		// 终端管理
		api.GET("/terminals", handlers.ListTerminals)
		api.POST("/terminals", handlers.CreateTerminal)
		api.DELETE("/terminals/:id", handlers.KillTerminal)
		api.POST("/terminals/write", handlers.WriteToTerminal)
		api.GET("/terminals/:id/read", handlers.ReadFromTerminal)

		// 代理服务
		api.GET("/proxy/status", handlers.GetProxyStatus)
		api.POST("/proxy/:type/start", handlers.StartProxy)
		api.POST("/proxy/:type/stop", handlers.StopProxy)

		// Shell命令
		api.POST("/shell/exec", handlers.ExecuteShellCommand)
	}

	return r
}
