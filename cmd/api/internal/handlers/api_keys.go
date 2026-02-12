package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/middleware"
	"github.com/model-system/api/internal/models"
)

// CreateAPIKey 创建API密钥
// POST /api/api-keys
func (h *Handler) CreateAPIKey(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	var req struct {
		KeyName string `json:"key_name"`
		Prompt  string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if req.KeyName == "" {
		req.KeyName = "Default Key"
	}

	apiKey, err := h.apiKeyService.GenerateAPIKey(userID, req.KeyName, req.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "创建成功",
		Data:    apiKey,
	})
}

// APIKeyWithPrompts API密钥带提示词列表
type APIKeyWithPrompts struct {
	ID        uint64                `json:"id"`
	UserID    uint64                `json:"user_id"`
	KeyName   string                `json:"key_name"`
	APIKey    string                `json:"api_key"`
	Prompt    string                `json:"prompt"`
	Prompts   []*models.APIKeyPrompt `json:"prompts"`    // 关联的提示词数组
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
}

// GetAPIKeys 获取当前用户的API密钥列表
// GET /api/api-keys
func (h *Handler) GetAPIKeys(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	apiKeys, err := h.apiKeyService.GetUserAPIKeys(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	// 转换为带提示词的结构
	result := make([]*APIKeyWithPrompts, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		// 获取每个密钥的提示词列表
		prompts, _ := h.apiKeyPromptService.GetPromptsByAPIKeyID(apiKey.ID)
		
		result = append(result, &APIKeyWithPrompts{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			KeyName:   apiKey.KeyName,
			APIKey:    apiKey.APIKey,
			Prompt:    apiKey.Prompt,
			Prompts:   prompts,
			CreatedAt: apiKey.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: apiKey.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    result,
	})
}

// DeleteAPIKey 删除API密钥
// DELETE /api/api-keys/:id
func (h *Handler) DeleteAPIKey(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	// 验证密钥所有权
	apiKey, err := h.apiKeyService.ValidateAPIKeyByID(id)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "API密钥不存在",
		})
	}

	if apiKey.UserID != userID {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权限删除此密钥",
		})
	}

	if err := h.apiKeyService.DeleteAPIKey(id); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "删除成功",
	})
}

// UpdateAPIKey 更新API密钥
// PUT /api/api-keys/:id
func (h *Handler) UpdateAPIKey(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	var req struct {
		KeyName string `json:"key_name"`
		Prompt  string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	apiKey, err := h.apiKeyService.UpdateAPIKey(id, userID, req.KeyName, req.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "更新成功",
		Data:    apiKey,
	})
}

// UpdatePrompt 更新API密钥提示词（兼容旧接口）
// PUT /api/api-keys/:id/prompt
func (h *Handler) UpdatePrompt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	var req struct {
		Prompt string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if err := h.apiKeyService.UpdatePrompt(id, req.Prompt); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "更新成功",
	})
}

// ================== API密钥提示词管理 ==================

// CreateAPIKeyPrompt 创建API密钥提示词
// POST /api/api-keys/:id/prompts
func (h *Handler) CreateAPIKeyPrompt(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	apiKeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的API密钥ID",
		})
	}

	// 验证密钥所有权
	apiKey, err := h.apiKeyService.ValidateAPIKeyByID(apiKeyID)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "API密钥不存在",
		})
	}

	if apiKey.UserID != userID {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权限操作此密钥",
		})
	}

	var req struct {
		ToolName string `json:"tool_name"`
		Prompt   string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	prompt, err := h.apiKeyPromptService.CreatePrompt(apiKeyID, req.ToolName, req.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "创建成功",
		Data:    prompt,
	})
}

// GetAPIKeyPrompts 获取API密钥的所有提示词
// GET /api/api-keys/:id/prompts
func (h *Handler) GetAPIKeyPrompts(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	apiKeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的API密钥ID",
		})
	}

	// 验证密钥所有权
	apiKey, err := h.apiKeyService.ValidateAPIKeyByID(apiKeyID)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "API密钥不存在",
		})
	}

	if apiKey.UserID != userID {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权限查看此密钥",
		})
	}

	prompts, err := h.apiKeyPromptService.GetPromptsByAPIKeyID(apiKeyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    prompts,
	})
}

// UpdateAPIKeyPrompt 更新API密钥提示词
// PUT /api/api-keys/:id/prompts/:prompt_id
func (h *Handler) UpdateAPIKeyPrompt(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	apiKeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的API密钥ID",
		})
	}

	promptID, err := strconv.ParseUint(c.Param("prompt_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的提示词ID",
		})
	}

	// 验证密钥所有权
	apiKey, err := h.apiKeyService.ValidateAPIKeyByID(apiKeyID)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "API密钥不存在",
		})
	}

	if apiKey.UserID != userID {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权限操作此密钥",
		})
	}

	var req struct {
		ToolName string `json:"tool_name"`
		Prompt   string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	prompt, err := h.apiKeyPromptService.UpdatePrompt(promptID, req.ToolName, req.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "更新成功",
		Data:    prompt,
	})
}

// DeleteAPIKeyPrompt 删除API密钥提示词
// DELETE /api/api-keys/:id/prompts/:prompt_id
func (h *Handler) DeleteAPIKeyPrompt(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	apiKeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的API密钥ID",
		})
	}

	promptID, err := strconv.ParseUint(c.Param("prompt_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的提示词ID",
		})
	}

	// 验证密钥所有权
	apiKey, err := h.apiKeyService.ValidateAPIKeyByID(apiKeyID)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "API密钥不存在",
		})
	}

	if apiKey.UserID != userID {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权限操作此密钥",
		})
	}

	if err := h.apiKeyPromptService.DeletePrompt(promptID); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "删除成功",
	})
}
