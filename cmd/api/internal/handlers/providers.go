package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateProvider 创建厂商
// POST /api/providers
func (h *Handler) CreateProvider(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		BaseURL     string `json:"base_url"`
		APIPrefix   string `json:"api_prefix"`
		APIKey      string `json:"api_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if req.Name == "" || req.DisplayName == "" || req.BaseURL == "" || req.APIPrefix == "" || req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "所有字段都不能为空",
		})
	}

	provider, err := h.providerService.Create(req.Name, req.DisplayName, req.BaseURL, req.APIPrefix, req.APIKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "创建成功",
		Data:    provider,
	})
}

// GetProviders 获取所有厂商
// GET /api/providers
func (h *Handler) GetProviders(c echo.Context) error {
	providers, err := h.providerService.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    providers,
	})
}

// GetProvider 获取单个厂商
// GET /api/providers/:id
func (h *Handler) GetProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	provider, err := h.providerService.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    provider,
	})
}

// UpdateProvider 更新厂商
// PUT /api/providers/:id
func (h *Handler) UpdateProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		BaseURL     string `json:"base_url"`
		APIPrefix   string `json:"api_prefix"`
		APIKey      string `json:"api_key"`
		Prompt      string `json:"prompt"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	provider, err := h.providerService.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
	}

	provider.Name = req.Name
	provider.DisplayName = req.DisplayName
	provider.BaseURL = req.BaseURL
	provider.APIPrefix = req.APIPrefix
	provider.APIKey = req.APIKey

	if err := h.providerService.Update(provider); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	// 更新厂商后刷新所有模型的全局缓存
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新成功，但缓存刷新失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "更新成功",
		Data:    provider,
	})
}

// DeleteProvider 删除厂商
// DELETE /api/providers/:id
func (h *Handler) DeleteProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的ID",
		})
	}

	if err := h.providerService.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	// 删除厂商后刷新所有模型的全局缓存
	if err := h.modelService.RefreshCache(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "删除成功，但缓存刷新失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "删除成功",
	})
}
