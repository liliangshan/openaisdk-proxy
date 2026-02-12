package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/config"
	"github.com/model-system/api/internal/service"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler HTTP处理器
type Handler struct {
	userService         *service.UserService
	apiKeyService       *service.APIKeyService
	apiKeyPromptService *service.APIKeyPromptService
	providerService     *service.ProviderService
	modelService        *service.ModelService
	cfg                 *config.Config
	jwtSecret           string
	jwtExpiration       time.Duration
}

// NewHandler 创建处理器
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		userService:         service.NewUserService(),
		apiKeyService:       service.NewAPIKeyService(),
		apiKeyPromptService: service.NewAPIKeyPromptService(),
		providerService:     service.NewProviderService(),
		modelService:        service.NewModelService(),
		cfg:                 cfg,
		jwtSecret:           cfg.JWT.Secret,
		jwtExpiration:       parseExpiration(cfg.JWT.Expiration),
	}
}

// parseExpiration 解析过期时间
func parseExpiration(exp string) time.Duration {
	duration, err := time.ParseDuration(exp)
	if err != nil {
		return 8760 * time.Hour // 默认一年
	}
	return duration
}

// Health 健康检查
// GET /health
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "OK",
	})
}

// DebugPath 调试用 - 打印请求路径
// GET /*path
func (h *Handler) DebugPath(c echo.Context) error {
	path := c.Param("*")
	println("Requested path:", path)
	return c.String(http.StatusOK, "Path: "+path)
}
