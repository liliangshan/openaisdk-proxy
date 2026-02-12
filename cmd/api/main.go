package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/cache"
	"github.com/model-system/api/internal/config"
	"github.com/model-system/api/internal/handlers"
	"github.com/model-system/api/internal/models"
	"github.com/model-system/api/internal/routes"
	"github.com/model-system/api/internal/service"
	"github.com/tiktoken-go/tokenizer"
)

//go:embed user/*
var staticFiles embed.FS

func main() {
	// 加载配置文件
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 解析命令行参数
	config.ParseArgs(cfg)
	if cfg.Debug {
		log.Println("调试模式已启用")
	}

	// 初始化数据库
	log.Println("正在连接数据库...")
	if err := models.InitDB(&cfg.Database); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer models.CloseDB()
	log.Println("数据库连接成功")

	// 初始化数据表
	log.Println("正在初始化数据表...")
	if err := models.InitTables(); err != nil {
		log.Fatalf("初始化数据表失败: %v", err)
	}
	log.Println("数据表初始化成功")

	// 初始化缓存
	cache.InitCache()
	log.Println("缓存初始化成功")

	// 创建模型服务并加载缓存
	modelService := service.NewModelService()
	if err := modelService.InitCache(); err != nil {
		log.Printf("警告: 加载模型缓存失败: %v", err)
	} else {
		log.Printf("模型缓存加载成功，共 %d 个模型", modelService.GetCache().GetModelCount())
	}

	// 加载API密钥到缓存
	log.Println("正在加载API密钥到缓存...")
	apiKeysWithUsers, err := models.GetAllAPIKeysWithUsers()
	if err != nil {
		log.Printf("警告: 查询API密钥失败: %v", err)
	} else {
		cache.GetCache().LoadAPIKeys(apiKeysWithUsers)
		log.Printf("API密钥缓存加载成功，共 %d 个密钥", cache.GetCache().GetAPIKeyCount())
	}

	// 加载API密钥工具提示词到缓存
	log.Println("正在加载API密钥工具提示词到缓存...")
	apiKeyPrompts, err := models.GetAllAPIKeyPrompts()
	if err != nil {
		log.Printf("警告: 查询API密钥提示词失败: %v", err)
	} else {
		// 构建 apiKeyID -> apiKey 的映射
		apiKeyIDToKey := make(map[uint64]string)
		for _, item := range apiKeysWithUsers {
			apiKeyIDToKey[item.ID] = item.APIKey
		}
		cache.GetCache().LoadAPIKeyPrompts(apiKeyIDToKey, apiKeyPrompts)
		log.Printf("API密钥工具提示词缓存加载成功，共 %d 条", len(apiKeyPrompts))
	}

	// 初始化 tokenizer
	log.Println("正在初始化 tokenizer...")
	tk, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		log.Printf("警告: 初始化 tokenizer 失败: %v", err)
	} else {
		handlers.SetTokenizer(tk)
		log.Println("tokenizer 初始化成功")
	}

	// 创建Echo实例
	e := echo.New()

	// JWT过期时间
	jwtExpiration := parseDuration(cfg.JWT.Expiration)

	// 设置路由
	routes.SetupRoutes(e, cfg, jwtExpiration)

	// 前端静态文件服务
	SetupFrontendRoutes(e)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)

	// 优雅关闭
	go func() {
		// 检查是否启用SSL
		if cfg.SSL.Enabled {
			// 检查证书文件是否存在
			if _, err := os.Stat(cfg.SSL.CertFile); os.IsNotExist(err) {
				log.Printf("警告: SSL证书文件不存在 %s，将使用HTTP模式", cfg.SSL.CertFile)
				log.Printf("启动HTTP服务器: %s", addr)
				if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP服务器启动失败: %v", err)
				}
			} else if _, err := os.Stat(cfg.SSL.KeyFile); os.IsNotExist(err) {
				log.Printf("警告: SSL密钥文件不存在 %s，将使用HTTP模式", cfg.SSL.KeyFile)
				log.Printf("启动HTTP服务器: %s", addr)
				if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP服务器启动失败: %v", err)
				}
			} else {
				// 启动HTTPS服务器（使用Echo实例的StartTLS方法）
				log.Printf("启动HTTPS服务器: %s", addr)
				if err := e.StartTLS(addr, cfg.SSL.CertFile, cfg.SSL.KeyFile); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTPS服务器启动失败: %v", err)
				}
			}
		} else {
			log.Printf("启动HTTP服务器: %s", addr)
			if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP服务器启动失败: %v", err)
			}
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
	if err := e.Close(); err != nil {
		log.Printf("服务器关闭错误: %v", err)
	}

	log.Println("服务器已关闭")
}

// parseDuration 解析过期时间
func parseDuration(exp string) time.Duration {
	duration, err := time.ParseDuration(exp)
	if err != nil {
		return 8760 * time.Hour // 默认一年
	}
	return duration
}

// SetupFrontendRoutes 设置前端静态文件路由（使用 embed）
func SetupFrontendRoutes(e *echo.Echo) {
	// 获取嵌入的静态文件
	staticFS, err := fs.Sub(staticFiles, "user")
	if err != nil {
		e.Logger.Warn("无法加载前端静态文件: %v", err)
		return
	}

	// 静态资源路由 - 直接提供文件（必须放在前面，更具体的路由优先）
	e.GET("/user/assets/:filepath", func(c echo.Context) error {
		filepath := c.Param("filepath")
		file, err := staticFS.Open("assets/" + filepath)
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer file.Close()

		// 根据文件后缀设置 Content-Type
		if strings.HasSuffix(filepath, ".js") {
			c.Response().Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(filepath, ".css") {
			c.Response().Header().Set("Content-Type", "text/css")
		}

		return c.Stream(http.StatusOK, "", file)
	})

	e.GET("/user/vite.svg", func(c echo.Context) error {
		file, err := staticFS.Open("vite.svg")
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer file.Close()
		c.Response().Header().Set("Content-Type", "image/svg+xml")
		return c.Stream(http.StatusOK, "", file)
	})

	// 前端页面路由 - 返回 index.html（支持 SPA，必须放在最后）
	e.GET("/user", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/html")
		file, err := staticFS.Open("index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer file.Close()
		return c.Stream(http.StatusOK, "", file)
	})
	e.GET("/user/", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/html")
		file, err := staticFS.Open("index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer file.Close()
		return c.Stream(http.StatusOK, "", file)
	})
	e.GET("/user/:filepath", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/html")
		file, err := staticFS.Open("index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer file.Close()
		return c.Stream(http.StatusOK, "", file)
	})
}
