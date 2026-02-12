package models

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/model-system/api/internal/config"
)

// DB 全局数据库连接
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(cfg.GetConnMaxDuration())

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return nil
}

// InitTables 初始化数据表
func InitTables() error {
	// 用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(64) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 密钥表
	apiKeysTable := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL,
		key_name VARCHAR(64) NOT NULL,
		api_key VARCHAR(255) NOT NULL UNIQUE,
		prompt TEXT NULL COMMENT 'API密钥默认提示词',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_api_key (api_key),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 模型厂商表（包含厂商密钥）
	providersTable := `
	CREATE TABLE IF NOT EXISTS providers (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(64) NOT NULL UNIQUE,
		display_name VARCHAR(128) NOT NULL,
		base_url VARCHAR(512) NOT NULL COMMENT 'OpenAI格式的接口地址',
		api_prefix VARCHAR(64) NOT NULL COMMENT 'API请求前缀',
		api_key VARCHAR(255) NOT NULL COMMENT '厂商API密钥',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 模型表
	modelsTable := `
	CREATE TABLE IF NOT EXISTS models (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL,
		provider_id BIGINT UNSIGNED NOT NULL,
		model_id VARCHAR(128) NOT NULL COMMENT '厂商内部的模型ID',
		display_name VARCHAR(128) NOT NULL,
		is_active TINYINT DEFAULT 1,
		context_length INT DEFAULT 128 COMMENT '上下文长度，单位k',
		compress_enabled TINYINT DEFAULT 1 COMMENT '是否启用token压缩',
		compress_truncate_len INT DEFAULT 500 COMMENT '截断过长消息的长度阈值',
		compress_user_count INT DEFAULT 3 COMMENT '压缩的user消息倒数数量',
		compress_role_types VARCHAR(128) DEFAULT '' COMMENT '角色类型，多个用逗号分开',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_provider_id (provider_id),
		INDEX idx_model_id (model_id),
		UNIQUE KEY uk_user_provider_model (user_id, provider_id, model_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// API密钥提示词表（关联工具）
	apiKeyPromptsTable := `
	CREATE TABLE IF NOT EXISTS api_key_prompts (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		api_key_id BIGINT UNSIGNED NOT NULL COMMENT '关联api_keys表',
		tool_name VARCHAR(128) NULL COMMENT '关联工具名（可选）',
		prompt TEXT NULL COMMENT '工具提示词',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_api_key_id (api_key_id),
		INDEX idx_tool_name (tool_name),
		FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	tables := []string{
		userTable,
		apiKeysTable,
		providersTable,
		modelsTable,
		apiKeyPromptsTable,
	}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return fmt.Errorf("创建数据表失败: %w", err)
		}
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// WithTx 执行事务
func WithTx(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// User 用户模型
type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey 密钥模型
type APIKey struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	KeyName   string    `json:"key_name"`
	APIKey    string    `json:"api_key"`
	Prompt    string    `json:"prompt"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Provider 模型厂商模型（包含API密钥）
type Provider struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	BaseURL     string    `json:"base_url"`
	APIPrefix   string    `json:"api_prefix"`
	APIKey      string    `json:"api_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Model 模型表（关联用户和厂商）
type Model struct {
	ID                  uint64    `json:"id"`
	UserID              uint64    `json:"user_id"`
	ProviderID          uint64    `json:"provider_id"`
	ModelID             string    `json:"model_id"`
	DisplayName         string    `json:"display_name"`
	IsActive            bool      `json:"is_active"`
	ContextLength       int       `json:"context_length"` // 上下文长度，单位k
	CompressEnabled     bool      `json:"compress_enabled"`
	CompressTruncateLen int       `json:"compress_truncate_len"`
	CompressUserCount   int       `json:"compress_user_count"`
	CompressRoleTypes   string    `json:"compress_role_types"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ModelWithDetails 模型详情（含厂商和用户信息）
type ModelWithDetails struct {
	Model
	ProviderName        string `json:"provider_name"`
	ProviderDisplayName string `json:"provider_display_name"`
	ProviderBaseURL     string `json:"provider_base_url"`
	ProviderAPIPrefix   string `json:"provider_api_prefix"`
	Username            string `json:"username"`
	ProviderKey         string `json:"provider_key,omitempty"`
}

// APIKeyWithUser 密钥与用户关联
type APIKeyWithUser struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	KeyName   string    `json:"key_name"`
	APIKey    string    `json:"api_key"`
	Prompt    string    `json:"prompt,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAllAPIKeysWithUsers 获取所有API密钥（包含用户ID）
func GetAllAPIKeysWithUsers() ([]APIKeyWithUser, error) {
	query := `
		SELECT id, user_id, key_name, api_key, prompt, created_at, updated_at
		FROM api_keys
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询API密钥失败: %w", err)
	}
	defer rows.Close()

	var apiKeys []APIKeyWithUser
	for rows.Next() {
		var apiKey APIKeyWithUser
		var apiKeyBytes []byte
		var promptBytes []byte
		if err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.KeyName,
			&apiKeyBytes,
			&promptBytes,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描API密钥失败: %w", err)
		}
		apiKey.APIKey = string(apiKeyBytes)
		apiKey.Prompt = string(promptBytes)
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历API密钥失败: %w", err)
	}

	return apiKeys, nil
}

// APIKeyPrompt API密钥提示词模型（关联工具）
type APIKeyPrompt struct {
	ID        uint64    `json:"id"`
	APIKeyID  uint64    `json:"api_key_id"`
	ToolName  string    `json:"tool_name"`
	Prompt    string    `json:"prompt"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAllAPIKeyPrompts 获取所有API密钥提示词（数组形式）
func GetAllAPIKeyPrompts() ([]APIKeyPrompt, error) {
	query := `
		SELECT id, api_key_id, tool_name, prompt, created_at, updated_at
		FROM api_key_prompts
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询API密钥提示词失败: %w", err)
	}
	defer rows.Close()

	var prompts []APIKeyPrompt
	for rows.Next() {
		var prompt APIKeyPrompt
		var promptBytes []byte
		if err := rows.Scan(
			&prompt.ID,
			&prompt.APIKeyID,
			&prompt.ToolName,
			&promptBytes,
			&prompt.CreatedAt,
			&prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描API密钥提示词失败: %w", err)
		}
		prompt.Prompt = string(promptBytes)
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历API密钥提示词失败: %w", err)
	}

	return prompts, nil
}
