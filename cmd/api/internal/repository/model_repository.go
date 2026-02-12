package repository

import (
	"database/sql"
	"fmt"

	"github.com/model-system/api/internal/models"
)

// ModelRepository 模型仓库
type ModelRepository struct{}

// NewModelRepository 创建模型仓库
func NewModelRepository() *ModelRepository {
	return &ModelRepository{}
}

// Create 创建模型
func (r *ModelRepository) Create(model *models.Model) error {
	query := `
		INSERT INTO models (user_id, provider_id, model_id, display_name, is_active, context_length, compress_enabled, compress_truncate_len, compress_user_count, compress_role_types)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := models.DB.Exec(query,
		model.UserID, model.ProviderID, model.ModelID, model.DisplayName, model.IsActive, model.ContextLength,
		model.CompressEnabled, model.CompressTruncateLen, model.CompressUserCount, model.CompressRoleTypes)
	if err != nil {
		return fmt.Errorf("创建模型失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取模型ID失败: %w", err)
	}

	model.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取模型
func (r *ModelRepository) GetByID(id uint64) (*models.Model, error) {
	query := `
		SELECT id, user_id, provider_id, model_id, display_name, is_active, context_length, created_at, updated_at
		FROM models
		WHERE id = ?
	`

	model := &models.Model{}
	err := models.DB.QueryRow(query, id).Scan(
		&model.ID,
		&model.UserID,
		&model.ProviderID,
		&model.ModelID,
		&model.DisplayName,
		&model.IsActive,
		&model.ContextLength,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询模型失败: %w", err)
	}

	return model, nil
}

// GetByIDWithDetails 根据ID获取模型详情（包含厂商和用户信息）
func (r *ModelRepository) GetByIDWithDetails(id uint64) (*models.ModelWithDetails, error) {
	query := `
		SELECT
			m.id, m.user_id, m.provider_id, m.model_id, m.display_name, m.is_active, m.context_length,
			m.compress_enabled, m.compress_truncate_len, m.compress_user_count, m.compress_role_types,
			m.created_at, m.updated_at,
			p.name as provider_name, p.display_name as provider_display_name,
			p.base_url as provider_base_url, p.api_prefix as provider_api_prefix,
			p.api_key as provider_api_key,
			u.username
		FROM models m
		LEFT JOIN providers p ON m.provider_id = p.id
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = ?
	`

	model := &models.ModelWithDetails{}
	err := models.DB.QueryRow(query, id).Scan(
		&model.ID,
		&model.UserID,
		&model.ProviderID,
		&model.ModelID,
		&model.DisplayName,
		&model.IsActive,
		&model.ContextLength,
		&model.CompressEnabled,
		&model.CompressTruncateLen,
		&model.CompressUserCount,
		&model.CompressRoleTypes,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.ProviderName,
		&model.ProviderDisplayName,
		&model.ProviderBaseURL,
		&model.ProviderAPIPrefix,
		&model.ProviderKey,
		&model.Username,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询模型详情失败: %w", err)
	}

	return model, nil
}

// GetAllWithDetails 获取所有模型详情（包含厂商和用户信息）
// providerID 为 0 时查询所有厂商
func (r *ModelRepository) GetAllWithDetails(providerID uint64) ([]models.ModelWithDetails, error) {
	query := `
		SELECT
			m.id, m.user_id, m.provider_id, m.model_id, m.display_name, m.is_active, m.context_length,
			m.compress_enabled, m.compress_truncate_len, m.compress_user_count, m.compress_role_types,
			m.created_at, m.updated_at,
			p.name as provider_name, p.display_name as provider_display_name,
			p.base_url as provider_base_url, p.api_prefix as provider_api_prefix,
			p.api_key as provider_api_key,
			u.username
		FROM models m
		LEFT JOIN providers p ON m.provider_id = p.id
		LEFT JOIN users u ON m.user_id = u.id
	`
	args := []interface{}{}

	if providerID > 0 {
		query += ` WHERE m.provider_id = ?`
		args = append(args, providerID)
	}

	query += ` ORDER BY m.id DESC`

	rows, err := models.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询模型列表失败: %w", err)
	}
	defer rows.Close()

	var modelsList []models.ModelWithDetails
	for rows.Next() {
		model := models.ModelWithDetails{}
		if err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.ProviderID,
			&model.ModelID,
			&model.DisplayName,
			&model.IsActive,
			&model.ContextLength,
			&model.CompressEnabled,
			&model.CompressTruncateLen,
			&model.CompressUserCount,
			&model.CompressRoleTypes,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.ProviderName,
			&model.ProviderDisplayName,
			&model.ProviderBaseURL,
			&model.ProviderAPIPrefix,
			&model.ProviderKey,
			&model.Username,
		); err != nil {
			return nil, fmt.Errorf("扫描模型失败: %w", err)
		}
		modelsList = append(modelsList, model)
	}

	return modelsList, nil
}

// GetByUserIDWithDetails 根据用户ID获取模型详情
// providerID 为 0 时查询所有厂商
func (r *ModelRepository) GetByUserIDWithDetails(userID, providerID uint64) ([]models.ModelWithDetails, error) {
	query := `
		SELECT
			m.id, m.user_id, m.provider_id, m.model_id, m.display_name, m.is_active, m.context_length,
			m.compress_enabled, m.compress_truncate_len, m.compress_user_count, m.compress_role_types,
			m.created_at, m.updated_at,
			p.name as provider_name, p.display_name as provider_display_name,
			p.base_url as provider_base_url, p.api_prefix as provider_api_prefix,
			p.api_key as provider_api_key,
			u.username
		FROM models m
		LEFT JOIN providers p ON m.provider_id = p.id
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.user_id = ?
	`
	args := []interface{}{userID}

	if providerID > 0 {
		query += ` AND m.provider_id = ?`
		args = append(args, providerID)
	}

	query += ` ORDER BY m.id DESC`

	rows, err := models.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询模型列表失败: %w", err)
	}
	defer rows.Close()

	var modelsList []models.ModelWithDetails
	for rows.Next() {
		model := models.ModelWithDetails{}
		if err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.ProviderID,
			&model.ModelID,
			&model.DisplayName,
			&model.IsActive,
			&model.ContextLength,
			&model.CompressEnabled,
			&model.CompressTruncateLen,
			&model.CompressUserCount,
			&model.CompressRoleTypes,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.ProviderName,
			&model.ProviderDisplayName,
			&model.ProviderBaseURL,
			&model.ProviderAPIPrefix,
			&model.ProviderKey,
			&model.Username,
		); err != nil {
			return nil, fmt.Errorf("扫描模型失败: %w", err)
		}
		modelsList = append(modelsList, model)
	}

	return modelsList, nil
}

// Update 更新模型
func (r *ModelRepository) Update(model *models.Model) error {
	query := `
		UPDATE models
		SET user_id = ?, provider_id = ?, model_id = ?, display_name = ?, is_active = ?, context_length = ?,
			compress_enabled = ?, compress_truncate_len = ?, compress_user_count = ?, compress_role_types = ?
		WHERE id = ?
	`

	_, err := models.DB.Exec(query,
		model.UserID, model.ProviderID, model.ModelID, model.DisplayName, model.IsActive, model.ContextLength,
		model.CompressEnabled, model.CompressTruncateLen, model.CompressUserCount, model.CompressRoleTypes,
		model.ID)
	if err != nil {
		return fmt.Errorf("更新模型失败: %w", err)
	}

	return nil
}

// Delete 删除模型
func (r *ModelRepository) Delete(id uint64) error {
	query := `DELETE FROM models WHERE id = ?`

	_, err := models.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除模型失败: %w", err)
	}

	return nil
}

// Exists 检查模型是否已存在（用户+厂商+模型ID组合）
func (r *ModelRepository) Exists(userID, providerID uint64, modelID string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM models
		WHERE user_id = ? AND provider_id = ? AND model_id = ?
	`

	var count int
	err := models.DB.QueryRow(query, userID, providerID, modelID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查模型是否存在失败: %w", err)
	}

	return count > 0, nil
}
