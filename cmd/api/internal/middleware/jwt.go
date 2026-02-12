package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken    = errors.New("无效的token")
	ErrExpiredToken    = errors.New("token已过期")
	ErrMissingToken    = errors.New("缺少token")
	ErrInvalidFormat   = errors.New("token格式无效")
)

// JWTMiddleware JWT中间件
func JWTMiddleware(secret string, expiration time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrMissingToken.Error())
			}

			// 检查Bearer前缀
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidFormat.Error())
			}

			tokenString := parts[1]

			// 解析token
			claims := &JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// 验证签名方法
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, ErrInvalidToken
				}
				return []byte(secret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					return echo.NewHTTPError(http.StatusUnauthorized, ErrExpiredToken.Error())
				}
				return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidToken.Error())
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidToken.Error())
			}

			// 将用户信息存入context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// GenerateToken 生成JWT token
func GenerateToken(secret string, expiration time.Duration, userID uint64, username string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "model-system",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GetUserID 从context获取用户ID
func GetUserID(c echo.Context) (uint64, bool) {
	userID, ok := c.Get("user_id").(uint64)
	return userID, ok
}

// GetUsername 从context获取用户名
func GetUsername(c echo.Context) (string, bool) {
	username, ok := c.Get("username").(string)
	return username, ok
}

// GetClaims 获取完整的claims
func GetClaims(c echo.Context) *JWTClaims {
	claims, ok := c.Get("claims").(*JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
