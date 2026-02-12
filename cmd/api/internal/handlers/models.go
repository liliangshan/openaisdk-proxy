package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/middleware"
)

// CreateModel 创建模型
// POST /api/models
func (h *Handler) CreateModel(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	var req struct {
		ProviderID          uint64 `json:"provider_id"`
		ModelID             string `json:"model_id"`
		DisplayName         string `json:"display_name"`
		ContextLength       int    `json:"context_length"`
		CompressEnabled     bool   `json:"compress_enabled"`
		CompressTruncateLen int    `json:"compress_truncate_len"`
		CompressUserCount   int    `json:"compress_user_count"`
		CompressRoleTypes   string `json:"compress_role_types"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if req.ProviderID == 0 || req.ModelID == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "厂商ID和模型ID不能为空",
		})
	}

	if req.DisplayName == "" {
		req.DisplayName = req.ModelID
	}

	// 设置默认值
	if req.CompressTruncateLen <= 0 {
		req.CompressTruncateLen = 500
	}
	if req.CompressUserCount <= 0 {
		req.CompressUserCount = 3
	}

	model, err := h.modelService.Create(userID, req.ProviderID, req.ModelID, req.DisplayName, req.ContextLength,
		req.CompressEnabled, req.CompressTruncateLen, req.CompressUserCount, req.CompressRoleTypes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
	}

	// 刷新全局缓存
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "模型创建成功，但缓存刷新失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "创建成功",
		Data:    model,
	})
}

// GetModels 获取所有模型
// GET /api/models
func (h *Handler) GetModels(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	// 获取厂商查询参数
	providerID := uint64(0)
	if providerIDStr := c.QueryParam("provider_id"); providerIDStr != "" {
		if parsed, err := strconv.ParseUint(providerIDStr, 10, 64); err == nil {
			providerID = parsed
		}
	}

	// 获取当前用户的模型
	models := h.modelService.GetByUserID(userID, providerID)

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    models,
	})
}

// GetAllModels 获取所有模型（管理员用）
// GET /api/admin/models
func (h *Handler) GetAllModels(c echo.Context) error {
	// 获取厂商查询参数
	providerID := uint64(0)
	if providerIDStr := c.QueryParam("provider_id"); providerIDStr != "" {
		if parsed, err := strconv.ParseUint(providerIDStr, 10, 64); err == nil {
			providerID = parsed
		}
	}

	models := h.modelService.GetAll(providerID)

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    models,
	})
}

// GetModel 获取单个模型
// GET /api/models/:id
func (h *Handler) GetModel(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	model, ok := h.modelService.GetByID(id)
	if !ok {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "模型不存在",
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    model,
	})
}

// UpdateModel 更新模型
// PUT /api/models/:id
func (h *Handler) UpdateModel(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	var req struct {
		ProviderID          uint64 `json:"provider_id"`
		ModelID             string `json:"model_id"`
		DisplayName         string `json:"display_name"`
		IsActive            bool   `json:"is_active"`
		ContextLength       int    `json:"context_length"`
		CompressEnabled     bool   `json:"compress_enabled"`
		CompressTruncateLen int    `json:"compress_truncate_len"`
		CompressUserCount   int    `json:"compress_user_count"`
		CompressRoleTypes   string `json:"compress_role_types"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	// 获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	// 设置默认值
	if req.CompressTruncateLen <= 0 {
		req.CompressTruncateLen = 500
	}
	if req.CompressUserCount <= 0 {
		req.CompressUserCount = 3
	}

	_, err = h.modelService.Update(id, userID, req.ProviderID, req.ModelID, req.DisplayName, req.IsActive, req.ContextLength,
		req.CompressEnabled, req.CompressTruncateLen, req.CompressUserCount, req.CompressRoleTypes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
	}

	// 刷新全局缓存
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "模型更新成功，但缓存刷新失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "更新成功",
	})
}

// DeleteModel 删除模型
// DELETE /api/models/:id
func (h *Handler) DeleteModel(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	if err := h.modelService.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	// 刷新全局缓存
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "模型删除成功，但缓存刷新失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "删除成功",
	})
}

// RefreshModelCache 刷新模型缓存
// POST /api/admin/models/refresh
func (h *Handler) RefreshModelCache(c echo.Context) error {
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "缓存刷新成功",
	})
}
