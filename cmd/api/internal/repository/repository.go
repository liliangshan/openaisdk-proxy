package repository

import (
	"database/sql"
	"fmt"

	"github.com/model-system/api/internal/models"
)

// UserRepository 用户仓库
type UserRepository struct{}

// NewUserRepository 创建用户仓库
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create 创建用户
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES (?, ?, ?)
	`

	result, err := models.DB.Exec(query, user.Username, user.Password, user.Email)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取用户ID失败: %w", err)
	}

	user.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uint64) (*models.User, error) {
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := models.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	user := &models.User{}
	err := models.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := models.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = ?, password = ?, email = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query, user.Username, user.Password, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint64) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := models.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// APIKeyRepository API密钥仓库
type APIKeyRepository struct{}

// NewAPIKeyRepository 创建API密钥仓库
func NewAPIKeyRepository() *APIKeyRepository {
	return &APIKeyRepository{}
}

// Create 创建API密钥
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, key_name, api_key, prompt)
		VALUES (?, ?, ?, ?)
	`

	result, err := models.DB.Exec(query, apiKey.UserID, apiKey.KeyName, apiKey.APIKey, apiKey.Prompt)
	if err != nil {
		return fmt.Errorf("创建API密钥失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取API密钥ID失败: %w", err)
	}

	apiKey.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取API密钥
func (r *APIKeyRepository) GetByID(id uint64) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, prompt, created_at, updated_at
		FROM api_keys
		WHERE id = ?
	`

	apiKey := &models.APIKey{}
	var promptBytes []byte
	err := models.DB.QueryRow(query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyName,
		&apiKey.APIKey,
		&promptBytes,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询API密钥失败: %w", err)
	}
	apiKey.Prompt = string(promptBytes)

	return apiKey, nil
}

// UpdatePrompt 更新API密钥提示词
func (r *APIKeyRepository) UpdatePrompt(id uint64, prompt string) error {
	query := `
		UPDATE api_keys
		SET prompt = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query, prompt, id)
	if err != nil {
		return fmt.Errorf("更新提示词失败: %w", err)
	}

	return nil
}

// GetByAPIKey 根据API密钥值获取
func (r *APIKeyRepository) GetByAPIKey(apiKeyValue string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, created_at, updated_at
		FROM api_keys
		WHERE api_key = ?
	`

	apiKey := &models.APIKey{}
	err := models.DB.QueryRow(query, apiKeyValue).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyName,
		&apiKey.APIKey,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询API密钥失败: %w", err)
	}

	return apiKey, nil
}

// GetByUserID 获取用户的所有API密钥
func (r *APIKeyRepository) GetByUserID(userID uint64) ([]*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, prompt, created_at, updated_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := models.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询API密钥列表失败: %w", err)
	}
	defer rows.Close()

	var apiKeys []*models.APIKey
	for rows.Next() {
		apiKey := &models.APIKey{}
		var promptBytes []byte
		if err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.KeyName,
			&apiKey.APIKey,
			&promptBytes,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描API密钥失败: %w", err)
		}
		apiKey.Prompt = string(promptBytes)
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// Delete 删除API密钥
func (r *APIKeyRepository) Delete(id uint64) error {
	query := `DELETE FROM api_keys WHERE id = ?`

	_, err := models.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除API密钥失败: %w", err)
	}

	return nil
}

// UpdateAPIKey 更新API密钥基本信息
func (r *APIKeyRepository) UpdateAPIKey(id uint64, keyName string, prompt string) error {
	query := `
		UPDATE api_keys
		SET key_name = ?, prompt = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query, keyName, prompt, id)
	if err != nil {
		return fmt.Errorf("更新API密钥失败: %w", err)
	}

	return nil
}

// ProviderRepository 厂商仓库
type ProviderRepository struct{}

// NewProviderRepository 创建厂商仓库
func NewProviderRepository() *ProviderRepository {
	return &ProviderRepository{}
}

// Create 创建厂商
func (r *ProviderRepository) Create(provider *models.Provider) error {
	query := `
		INSERT INTO providers (name, display_name, base_url, api_prefix, api_key)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := models.DB.Exec(query, provider.Name, provider.DisplayName, provider.BaseURL, provider.APIPrefix, provider.APIKey)
	if err != nil {
		return fmt.Errorf("创建厂商失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取厂商ID失败: %w", err)
	}

	provider.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取厂商
func (r *ProviderRepository) GetByID(id uint64) (*models.Provider, error) {
	query := `
		SELECT id, name, display_name, base_url, api_prefix, api_key, created_at, updated_at
		FROM providers
		WHERE id = ?
	`

	provider := &models.Provider{}
	err := models.DB.QueryRow(query, id).Scan(
		&provider.ID,
		&provider.Name,
		&provider.DisplayName,
		&provider.BaseURL,
		&provider.APIPrefix,
		&provider.APIKey,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询厂商失败: %w", err)
	}

	return provider, nil
}

// GetByName 根据名称获取厂商
func (r *ProviderRepository) GetByName(name string) (*models.Provider, error) {
	query := `
		SELECT id, name, display_name, base_url, api_prefix, api_key, created_at, updated_at
		FROM providers
		WHERE name = ?
	`

	provider := &models.Provider{}
	err := models.DB.QueryRow(query, name).Scan(
		&provider.ID,
		&provider.Name,
		&provider.DisplayName,
		&provider.BaseURL,
		&provider.APIPrefix,
		&provider.APIKey,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询厂商失败: %w", err)
	}

	return provider, nil
}

// GetAll 获取所有厂商
func (r *ProviderRepository) GetAll() ([]*models.Provider, error) {
	query := `
		SELECT id, name, display_name, base_url, api_prefix, api_key, created_at, updated_at
		FROM providers
		ORDER BY name ASC
	`

	rows, err := models.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询厂商列表失败: %w", err)
	}
	defer rows.Close()

	var providers []*models.Provider
	for rows.Next() {
		provider := &models.Provider{}

		if err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.DisplayName,
			&provider.BaseURL,
			&provider.APIPrefix,
			&provider.APIKey,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描厂商失败: %w", err)
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

// Update 更新厂商
func (r *ProviderRepository) Update(provider *models.Provider) error {
	query := `
		UPDATE providers
		SET name = ?, display_name = ?, base_url = ?, api_prefix = ?, api_key = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query, provider.Name, provider.DisplayName, provider.BaseURL, provider.APIPrefix, provider.APIKey, provider.ID)
	if err != nil {
		return fmt.Errorf("更新厂商失败: %w", err)
	}

	return nil
}

// Delete 删除厂商
func (r *ProviderRepository) Delete(id uint64) error {
	query := `DELETE FROM providers WHERE id = ?`

	_, err := models.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除厂商失败: %w", err)
	}

	return nil
}

// APIKeyPromptRepository API密钥提示词仓库
type APIKeyPromptRepository struct{}

// NewAPIKeyPromptRepository 创建API密钥提示词仓库
func NewAPIKeyPromptRepository() *APIKeyPromptRepository {
	return &APIKeyPromptRepository{}
}

// Create 创建API密钥提示词
func (r *APIKeyPromptRepository) Create(prompt *models.APIKeyPrompt) error {
	query := `
		INSERT INTO api_key_prompts (api_key_id, tool_name, prompt)
		VALUES (?, ?, ?)
	`

	result, err := models.DB.Exec(query, prompt.APIKeyID, prompt.ToolName, prompt.Prompt)
	if err != nil {
		return fmt.Errorf("创建API密钥提示词失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取API密钥提示词ID失败: %w", err)
	}

	prompt.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取API密钥提示词
func (r *APIKeyPromptRepository) GetByID(id uint64) (*models.APIKeyPrompt, error) {
	query := `
		SELECT id, api_key_id, tool_name, prompt, created_at, updated_at
		FROM api_key_prompts
		WHERE id = ?
	`

	prompt := &models.APIKeyPrompt{}
	var promptBytes []byte
	err := models.DB.QueryRow(query, id).Scan(
		&prompt.ID,
		&prompt.APIKeyID,
		&prompt.ToolName,
		&promptBytes,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询API密钥提示词失败: %w", err)
	}
	prompt.Prompt = string(promptBytes)

	return prompt, nil
}

// GetByAPIKeyID 根据API密钥ID获取所有提示词
func (r *APIKeyPromptRepository) GetByAPIKeyID(apiKeyID uint64) ([]*models.APIKeyPrompt, error) {
	query := `
		SELECT id, api_key_id, tool_name, prompt, created_at, updated_at
		FROM api_key_prompts
		WHERE api_key_id = ?
		ORDER BY tool_name ASC
	`

	rows, err := models.DB.Query(query, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("查询API密钥提示词列表失败: %w", err)
	}
	defer rows.Close()

	var prompts []*models.APIKeyPrompt
	for rows.Next() {
		prompt := &models.APIKeyPrompt{}
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

	return prompts, nil
}

// GetByAPIKeyIDAndToolName 根据API密钥ID和工具名获取提示词
func (r *APIKeyPromptRepository) GetByAPIKeyIDAndToolName(apiKeyID uint64, toolName string) (*models.APIKeyPrompt, error) {
	query := `
		SELECT id, api_key_id, tool_name, prompt, created_at, updated_at
		FROM api_key_prompts
		WHERE api_key_id = ? AND tool_name = ?
	`

	prompt := &models.APIKeyPrompt{}
	var promptBytes []byte
	err := models.DB.QueryRow(query, apiKeyID, toolName).Scan(
		&prompt.ID,
		&prompt.APIKeyID,
		&prompt.ToolName,
		&promptBytes,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询API密钥提示词失败: %w", err)
	}
	prompt.Prompt = string(promptBytes)

	return prompt, nil
}

// Update 更新API密钥提示词
func (r *APIKeyPromptRepository) Update(prompt *models.APIKeyPrompt) error {
	query := `
		UPDATE api_key_prompts
		SET tool_name = ?, prompt = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query, prompt.ToolName, prompt.Prompt, prompt.ID)
	if err != nil {
		return fmt.Errorf("更新API密钥提示词失败: %w", err)
	}

	return nil
}

// Delete 删除API密钥提示词
func (r *APIKeyPromptRepository) Delete(id uint64) error {
	query := `DELETE FROM api_key_prompts WHERE id = ?`

	_, err := models.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除API密钥提示词失败: %w", err)
	}

	return nil
}

// DeleteByAPIKeyID 删除指定API密钥的所有提示词
func (r *APIKeyPromptRepository) DeleteByAPIKeyID(apiKeyID uint64) error {
	query := `DELETE FROM api_key_prompts WHERE api_key_id = ?`

	_, err := models.DB.Exec(query, apiKeyID)
	if err != nil {
		return fmt.Errorf("删除API密钥提示词失败: %w", err)
	}

	return nil
}
