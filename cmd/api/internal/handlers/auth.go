package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/middleware"
)

// Register 注册用户
// /api POST/auth/register
func (h *Handler) Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "用户名、密码和邮箱不能为空",
		})
	}

	user, _, err := h.userService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
	}

	// 生成token
	token, err := middleware.GenerateToken(h.jwtSecret, h.jwtExpiration, user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "生成token失败",
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "注册成功",
		Data: map[string]interface{}{
			"user":  user,
			"token": token,
		},
	})
}

// Login 登录
// POST /api/auth/login
func (h *Handler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "用户名和密码不能为空",
		})
	}

	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: err.Error(),
		})
	}

	// user 实际上是 nil，因为 Login 返回的是错误信息字符串
	// 我们需要重新查询用户信息
	fullUser, err := h.userService.GetUserByID(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取用户信息失败",
		})
	}

	// 生成token
	token, err := middleware.GenerateToken(h.jwtSecret, h.jwtExpiration, fullUser.ID, fullUser.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "生成token失败",
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "登录成功",
		Data: map[string]interface{}{
			"user":  fullUser,
			"token": token,
		},
	})
}

// GetProfile 获取当前用户信息
// GET /api/auth/profile
func (h *Handler) GetProfile(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
		})
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "获取成功",
		Data:    user,
	})
}
