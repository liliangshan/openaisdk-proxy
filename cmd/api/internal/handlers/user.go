package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// UserIndex 用户页面入口
// GET /user
// GET /user/index.html
func (h *Handler) UserIndex(c echo.Context) error {
	return c.File("user/index.html")
}

// UserStatic 用户静态文件
// GET /user/* (前端构建的资源文件，文件不存在时返回index.html支持SPA路由)
func (h *Handler) UserStatic(c echo.Context) error {
	path := c.Param("path")
	if path == "" || path == "/" {
		return c.File("user/index.html")
	}

	// 防止路径遍历攻击
	if strings.Contains(path, "..") || strings.Contains(path, ":") {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权访问该文件",
		})
	}

	// 构建完整路径
	fullPath := filepath.Join("user", path)

	// 确保路径仍然在 user 目录下
	if !strings.HasPrefix(fullPath, "user") {
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "无权访问该文件",
		})
	}

	// 检查文件是否存在，不存在则返回 index.html（SPA路由支持）
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.File("user/index.html")
	}

	return c.File(fullPath)
}
