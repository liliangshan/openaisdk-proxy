package routes

import (
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/model-system/api/internal/config"
	"github.com/model-system/api/internal/handlers"
	"github.com/model-system/api/internal/middleware"
)

// SetupRoutes 设置所有路由
func SetupRoutes(e *echo.Echo, cfg *config.Config, jwtExpiration time.Duration) {
	e.HideBanner = true

	// 添加全局中间件
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			// 记录请求日志
			return nil
		},
	}))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	e.Use(echomiddleware.RequestID())

	// 创建处理器
	h := handlers.NewHandler(cfg)

	// 路由分组
	api := e.Group("/api")

	// 健康检查
	e.GET("/health", h.Health)

	// ========== 公开路由（无需认证）==========
	auth := api.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	// ========== 需要认证的路由 ==========
	authProfile := api.Group("/auth")
	authProfile.Use(middleware.JWTMiddleware(cfg.JWT.Secret, jwtExpiration))
	authProfile.GET("/profile", h.GetProfile)

	// ========== API密钥管理 ==========
	apiKeys := api.Group("/api-keys")
	apiKeys.Use(middleware.JWTMiddleware(cfg.JWT.Secret, jwtExpiration))
	apiKeys.POST("", h.CreateAPIKey)
	apiKeys.GET("", h.GetAPIKeys)
	apiKeys.PUT("/:id", h.UpdateAPIKey)
	apiKeys.DELETE("/:id", h.DeleteAPIKey)
	apiKeys.PUT("/:id/prompt", h.UpdatePrompt)
	// API密钥提示词管理
	apiKeys.GET("/:id/prompts", h.GetAPIKeyPrompts)
	apiKeys.POST("/:id/prompts", h.CreateAPIKeyPrompt)
	apiKeys.PUT("/:id/prompts/:prompt_id", h.UpdateAPIKeyPrompt)
	apiKeys.DELETE("/:id/prompts/:prompt_id", h.DeleteAPIKeyPrompt)

	// ========== 厂商管理 ==========
	providers := api.Group("/providers")
	providers.Use(middleware.JWTMiddleware(cfg.JWT.Secret, jwtExpiration))
	providers.POST("", h.CreateProvider)
	providers.GET("", h.GetProviders)
	providers.GET("/:id", h.GetProvider)
	providers.PUT("/:id", h.UpdateProvider)
	providers.DELETE("/:id", h.DeleteProvider)

	// ========== 模型管理 ==========
	models := api.Group("/models")
	models.Use(middleware.JWTMiddleware(cfg.JWT.Secret, jwtExpiration))
	models.POST("", h.CreateModel)
	models.GET("", h.GetModels)
	models.GET("/:id", h.GetModel)
	models.PUT("/:id", h.UpdateModel)
	models.DELETE("/:id", h.DeleteModel)

	// ========== 管理员路由 ==========
	admin := api.Group("/admin")
	admin.Use(middleware.JWTMiddleware(cfg.JWT.Secret, jwtExpiration))
	admin.GET("/models", h.GetAllModels)
	admin.POST("/models/refresh", h.RefreshModelCache)

	// ========== Chat Completion 路由（无需认证，通过API密钥验证）==========
	chat := api.Group("/v1/chat")
	chat.POST("/completions", h.ChatCompletion)

	// ========== 用户页面路由 ==========
	user := e.Group("/user")
	user.GET("", h.UserIndex)
	user.GET("/", h.UserIndex)
	user.GET("/index.html", h.UserIndex)
	user.GET("/:path", h.UserStatic)

	// ========== 证书验证路由 ==========
	cert := e.Group("/cert")
	cert.POST("/verify", h.VerifyCert)
	cert.GET("/info", h.GetCertInfo)

	// ========== 任意路径路由（调试用）==========
	e.GET("/*", h.DebugPath)
}
