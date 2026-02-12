package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/model-system/api/internal/cache"
	"github.com/model-system/api/internal/models"
	"github.com/model-system/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrUserAlreadyExist = errors.New("用户已存在")
	ErrModelNotFound    = errors.New("模型不存在")
	ErrProviderNotFound = errors.New("厂商不存在")
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

// Register 注册用户
func (s *UserService) Register(username, password, email string) (*models.User, string, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", fmt.Errorf("检查用户名失败: %w", err)
	}
	if existingUser != nil {
		return nil, "", ErrUserAlreadyExist
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingEmail != nil {
		return nil, "", errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("加密密码失败: %w", err)
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("创建用户失败: %w", err)
	}

	return user, "", nil
}

// Login 登录
func (s *UserService) Login(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint64) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// APIKeyService API密钥服务
type APIKeyService struct {
	apiKeyRepo *repository.APIKeyRepository
}

// NewAPIKeyService 创建API密钥服务
func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: repository.NewAPIKeyRepository(),
	}
}

// GenerateAPIKey 生成API密钥
func (s *APIKeyService) GenerateAPIKey(userID uint64, keyName string, prompt string) (*models.APIKey, error) {
	var apiKeyValue string

	// 生成随机密钥，直到在缓存中不存在
	for {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			return nil, fmt.Errorf("生成密钥失败: %w", err)
		}
		apiKeyValue = "sk_" + hex.EncodeToString(bytes)

		// 检查缓存中是否已存在
		if _, exists := cache.GetCache().GetUserIDByAPIKey(apiKeyValue); !exists {
			break
		}
		// 如果存在，重新生成
	}

	apiKey := &models.APIKey{
		UserID:  userID,
		KeyName: keyName,
		APIKey:  apiKeyValue,
		Prompt:  prompt,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("保存API密钥失败: %w", err)
	}

	// 添加到缓存
	cache.GetCache().AddAPIKey(apiKeyValue, userID, prompt)

	return apiKey, nil
}

// GetUserAPIKeys 获取用户的所有API密钥
func (s *APIKeyService) GetUserAPIKeys(userID uint64) ([]*models.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(userID)
}

// ValidateAPIKey 验证API密钥
func (s *APIKeyService) ValidateAPIKey(apiKeyValue string) (*models.APIKey, error) {
	return s.apiKeyRepo.GetByAPIKey(apiKeyValue)
}

// ValidateAPIKeyByID 根据ID验证API密钥
func (s *APIKeyService) ValidateAPIKeyByID(id uint64) (*models.APIKey, error) {
	return s.apiKeyRepo.GetByID(id)
}

// DeleteAPIKey 删除API密钥
func (s *APIKeyService) DeleteAPIKey(id uint64) error {
	// 先获取要删除的密钥
	existingKey, err := s.apiKeyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查找API密钥失败: %w", err)
	}
	if existingKey == nil {
		return errors.New("API密钥不存在")
	}

	// 删除数据库记录
	if err := s.apiKeyRepo.Delete(id); err != nil {
		return fmt.Errorf("删除API密钥失败: %w", err)
	}

	// 从缓存中删除
	cache.GetCache().DeleteAPIKey(existingKey.APIKey)

	return nil
}

// UpdatePrompt 更新API密钥提示词
func (s *APIKeyService) UpdatePrompt(id uint64, prompt string) error {
	// 先获取密钥信息
	existingKey, err := s.apiKeyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查找API密钥失败: %w", err)
	}
	if existingKey == nil {
		return errors.New("API密钥不存在")
	}

	// 更新数据库
	if err := s.apiKeyRepo.UpdatePrompt(id, prompt); err != nil {
		return err
	}

	// 更新缓存
	cache.GetCache().DeleteAPIKey(existingKey.APIKey)
	cache.GetCache().AddAPIKey(existingKey.APIKey, existingKey.UserID, prompt)

	return nil
}

// UpdateAPIKey 更新API密钥基本信息
func (s *APIKeyService) UpdateAPIKey(id uint64, userID uint64, keyName string, prompt string) (*models.APIKey, error) {
	// 先获取密钥信息
	existingKey, err := s.apiKeyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找API密钥失败: %w", err)
	}
	if existingKey == nil {
		return nil, errors.New("API密钥不存在")
	}

	// 验证用户权限
	if existingKey.UserID != userID {
		return nil, errors.New("无权限修改此密钥")
	}

	// 更新数据库
	if err := s.apiKeyRepo.UpdateAPIKey(id, keyName, prompt); err != nil {
		return nil, err
	}

	// 更新缓存
	cache.GetCache().DeleteAPIKey(existingKey.APIKey)
	cache.GetCache().AddAPIKey(existingKey.APIKey, existingKey.UserID, prompt)

	// 返回更新后的密钥
	return s.apiKeyRepo.GetByID(id)
}

// ProviderService 厂商服务
type ProviderService struct {
	providerRepo *repository.ProviderRepository
}

// NewProviderService 创建厂商服务
func NewProviderService() *ProviderService {
	return &ProviderService{
		providerRepo: repository.NewProviderRepository(),
	}
}

// Create 创建厂商
func (s *ProviderService) Create(name, displayName, baseURL, apiPrefix, apiKey string) (*models.Provider, error) {
	// 检查厂商名是否已存在
	existing, err := s.providerRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("检查厂商名失败: %w", err)
	}
	if existing != nil {
		return nil, errors.New("厂商已存在")
	}

	provider := &models.Provider{
		Name:         name,
		DisplayName:  displayName,
		BaseURL:      baseURL,
		APIPrefix:    apiPrefix,
		APIKey:       apiKey,
	}

	if err := s.providerRepo.Create(provider); err != nil {
		return nil, fmt.Errorf("创建厂商失败: %w", err)
	}

	return provider, nil
}

// GetAll 获取所有厂商
func (s *ProviderService) GetAll() ([]*models.Provider, error) {
	return s.providerRepo.GetAll()
}

// GetByID 根据ID获取厂商
func (s *ProviderService) GetByID(id uint64) (*models.Provider, error) {
	provider, err := s.providerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询厂商失败: %w", err)
	}
	if provider == nil {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// Update 更新厂商
func (s *ProviderService) Update(provider *models.Provider) error {
	return s.providerRepo.Update(provider)
}

// Delete 删除厂商
func (s *ProviderService) Delete(id uint64) error {
	return s.providerRepo.Delete(id)
}

// ModelService 模型服务
type ModelService struct {
	modelRepo *repository.ModelRepository
	cache     *cache.MemoryCache
}

// NewModelService 创建模型服务
func NewModelService() *ModelService {
	return &ModelService{
		modelRepo: repository.NewModelRepository(),
		cache:     cache.GetCache(),
	}
}

// InitCache 初始化缓存
func (s *ModelService) InitCache() error {
	models, err := s.modelRepo.GetAllWithDetails(0)
	if err != nil {
		return fmt.Errorf("加载模型缓存失败: %w", err)
	}

	s.cache.LoadModels(models)
	return nil
}

// Create 创建模型
func (s *ModelService) Create(userID, providerID uint64, modelID, displayName string, contextLength int,
	compressEnabled bool, compressTruncateLen, compressUserCount int, compressRoleTypes string) (*models.Model, error) {
	// 检查是否已存在
	exists, err := s.modelRepo.Exists(userID, providerID, modelID)
	if err != nil {
		return nil, fmt.Errorf("检查模型是否存在失败: %w", err)
	}
	if exists {
		return nil, errors.New("模型已存在")
	}

	// 默认上下文长度
	if contextLength == 0 {
		contextLength = 128
	}

	// 设置压缩默认值
	if compressTruncateLen <= 0 {
		compressTruncateLen = 500
	}
	if compressUserCount <= 0 {
		compressUserCount = 3
	}

	model := &models.Model{
		UserID:              userID,
		ProviderID:          providerID,
		ModelID:             modelID,
		DisplayName:         displayName,
		IsActive:            true,
		ContextLength:       contextLength,
		CompressEnabled:     compressEnabled,
		CompressTruncateLen: compressTruncateLen,
		CompressUserCount:   compressUserCount,
		CompressRoleTypes:   compressRoleTypes,
	}

	if err := s.modelRepo.Create(model); err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	// 添加到缓存
	modelWithDetails, err := s.modelRepo.GetByIDWithDetails(model.ID)
	if err != nil {
		return nil, fmt.Errorf("获取模型详情失败: %w", err)
	}
	s.cache.AddModel(modelWithDetails)

	return model, nil
}

// GetAll 获取所有模型（从缓存）
// providerID 为 0 时返回所有厂商
func (s *ModelService) GetAll(providerID uint64) []*models.ModelWithDetails {
	cacheItems := s.cache.GetAllModels()
	result := make([]*models.ModelWithDetails, 0, len(cacheItems))
	for _, item := range cacheItems {
		if providerID == 0 || item.Model.ProviderID == providerID {
			result = append(result, s.toModelWithDetails(item))
		}
	}
	return result
}

// GetByID 根据ID获取模型（优先从缓存）
func (s *ModelService) GetByID(id uint64) (*models.ModelWithDetails, bool) {
	item, ok := s.cache.GetModel(id)
	if !ok {
		return nil, false
	}
	return s.toModelWithDetails(item), true
}

// GetByProviderAndModelID 根据厂商前缀和模型ID获取模型
func (s *ModelService) GetByProviderAndModelID(providerPrefix, modelID string) (*models.ModelWithDetails, bool) {
	item, ok := s.cache.GetModelByKey(providerPrefix, modelID)
	if !ok {
		return nil, false
	}
	return s.toModelWithDetails(item), true
}

// GetByUserID 根据用户ID获取模型
// providerID 为 0 时返回所有厂商
func (s *ModelService) GetByUserID(userID, providerID uint64) []*models.ModelWithDetails {
	cacheItems := s.cache.GetModelsByUser(userID)
	result := make([]*models.ModelWithDetails, 0, len(cacheItems))
	for _, item := range cacheItems {
		if providerID == 0 || item.Model.ProviderID == providerID {
			result = append(result, s.toModelWithDetails(item))
		}
	}
	return result
}

// toModelWithDetails 将缓存项转换为模型详情
func (s *ModelService) toModelWithDetails(item *cache.ModelCacheItem) *models.ModelWithDetails {
	return &models.ModelWithDetails{
		Model:               item.Model,
		ProviderName:        item.ProviderName,
		ProviderDisplayName:  item.ProviderDisplayName,
		ProviderBaseURL:     item.ProviderBaseURL,
		ProviderAPIPrefix:   item.ProviderAPIPrefix,
		Username:            item.Username,
		ProviderKey:         item.ProviderKey,
	}
}

// Update 更新模型
func (s *ModelService) Update(id uint64, userID, providerID uint64, modelID, displayName string, isActive bool, contextLength int,
	compressEnabled bool, compressTruncateLen, compressUserCount int, compressRoleTypes string) (*models.Model, error) {
	// 检查模型是否存在
	existing, err := s.modelRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询模型失败: %w", err)
	}
	if existing == nil {
		return nil, ErrModelNotFound
	}

	// 检查新的组合是否已存在（排除自己）
	if userID != existing.UserID || providerID != existing.ProviderID || modelID != existing.ModelID {
		exists, err := s.modelRepo.Exists(userID, providerID, modelID)
		if err != nil {
			return nil, fmt.Errorf("检查模型是否存在失败: %w", err)
		}
		if exists {
			return nil, errors.New("模型已存在")
		}
	}

	// 默认上下文长度
	if contextLength == 0 {
		contextLength = 128
	}

	// 设置压缩默认值
	if compressTruncateLen <= 0 {
		compressTruncateLen = 500
	}
	if compressUserCount <= 0 {
		compressUserCount = 3
	}

	model := &models.Model{
		ID:                  id,
		UserID:              userID,
		ProviderID:          providerID,
		ModelID:             modelID,
		DisplayName:         displayName,
		IsActive:            isActive,
		ContextLength:       contextLength,
		CompressEnabled:     compressEnabled,
		CompressTruncateLen: compressTruncateLen,
		CompressUserCount:   compressUserCount,
		CompressRoleTypes:   compressRoleTypes,
	}

	if err := s.modelRepo.Update(model); err != nil {
		return nil, fmt.Errorf("更新模型失败: %w", err)
	}

	// 更新缓存
	modelWithDetails, err := s.modelRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("获取模型详情失败: %w", err)
	}
	s.cache.UpdateModel(modelWithDetails)

	return model, nil
}

// Delete 删除模型
func (s *ModelService) Delete(id uint64) error {
	// 检查模型是否存在
	existing, err := s.modelRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查询模型失败: %w", err)
	}
	if existing == nil {
		return ErrModelNotFound
	}

	if err := s.modelRepo.Delete(id); err != nil {
		return fmt.Errorf("删除模型失败: %w", err)
	}

	// 从缓存中删除
	s.cache.DeleteModel(id)

	return nil
}

// RefreshCache 刷新缓存
func (s *ModelService) RefreshCache() error {
	models, err := s.modelRepo.GetAllWithDetails(0)
	if err != nil {
		return fmt.Errorf("重新加载模型缓存失败: %w", err)
	}

	s.cache.Refresh(models)
	return nil
}

// GetCache 获取缓存实例
func (s *ModelService) GetCache() *cache.MemoryCache {
	return s.cache
}

// APIKeyPromptService API密钥提示词服务
type APIKeyPromptService struct {
	promptRepo *repository.APIKeyPromptRepository
	apiKeyRepo *repository.APIKeyRepository
	cache      *cache.MemoryCache
}

// NewAPIKeyPromptService 创建API密钥提示词服务
func NewAPIKeyPromptService() *APIKeyPromptService {
	return &APIKeyPromptService{
		promptRepo: repository.NewAPIKeyPromptRepository(),
		apiKeyRepo: repository.NewAPIKeyRepository(),
		cache:      cache.GetCache(),
	}
}

// CreatePrompt 创建API密钥提示词
func (s *APIKeyPromptService) CreatePrompt(apiKeyID uint64, toolName string, prompt string) (*models.APIKeyPrompt, error) {
	// 验证API密钥存在
	apiKey, err := s.apiKeyRepo.GetByID(apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("查找API密钥失败: %w", err)
	}
	if apiKey == nil {
		return nil, errors.New("API密钥不存在")
	}

	// 只有工具名非空时才检查是否已存在
	if toolName != "" {
		existing, err := s.promptRepo.GetByAPIKeyIDAndToolName(apiKeyID, toolName)
		if err != nil {
			return nil, fmt.Errorf("检查提示词失败: %w", err)
		}
		if existing != nil {
			return nil, errors.New("该工具的提示词已存在")
		}
	}

	p := &models.APIKeyPrompt{
		APIKeyID: apiKeyID,
		ToolName: toolName,
		Prompt:   prompt,
	}

	if err := s.promptRepo.Create(p); err != nil {
		return nil, err
	}

	// 更新缓存
	s.cache.AddAPIKeyPrompt(apiKey.APIKey, toolName, prompt)

	return p, nil
}

// GetPromptsByAPIKeyID 根据API密钥ID获取所有提示词
func (s *APIKeyPromptService) GetPromptsByAPIKeyID(apiKeyID uint64) ([]*models.APIKeyPrompt, error) {
	return s.promptRepo.GetByAPIKeyID(apiKeyID)
}

// UpdatePrompt 更新API密钥提示词
func (s *APIKeyPromptService) UpdatePrompt(id uint64, toolName string, prompt string) (*models.APIKeyPrompt, error) {
	existing, err := s.promptRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找提示词失败: %w", err)
	}
	if existing == nil {
		return nil, errors.New("提示词不存在")
	}

	// 获取API密钥信息用于更新缓存
	apiKey, err := s.apiKeyRepo.GetByID(existing.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("查找API密钥失败: %w", err)
	}

	oldToolName := existing.ToolName

	// 只有工具名非空且变更时，才检查新工具名是否已存在
	if toolName != "" && toolName != existing.ToolName {
		duplicate, err := s.promptRepo.GetByAPIKeyIDAndToolName(existing.APIKeyID, toolName)
		if err != nil {
			return nil, fmt.Errorf("检查提示词失败: %w", err)
		}
		if duplicate != nil {
			return nil, errors.New("该工具的提示词已存在")
		}
	}

	existing.ToolName = toolName
	existing.Prompt = prompt

	if err := s.promptRepo.Update(existing); err != nil {
		return nil, err
	}

	// 更新缓存：先删除旧的，再添加新的
	if apiKey != nil {
		s.cache.DeleteAPIKeyPrompt(apiKey.APIKey, oldToolName)
		s.cache.AddAPIKeyPrompt(apiKey.APIKey, toolName, prompt)
	}

	return existing, nil
}

// DeletePrompt 删除API密钥提示词
func (s *APIKeyPromptService) DeletePrompt(id uint64) error {
	// 先获取提示词信息用于更新缓存
	existing, err := s.promptRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查找提示词失败: %w", err)
	}
	if existing == nil {
		return errors.New("提示词不存在")
	}

	// 获取API密钥信息
	apiKey, err := s.apiKeyRepo.GetByID(existing.APIKeyID)
	if err != nil {
		return fmt.Errorf("查找API密钥失败: %w", err)
	}

	if err := s.promptRepo.Delete(id); err != nil {
		return err
	}

	// 更新缓存
	if apiKey != nil {
		s.cache.DeleteAPIKeyPrompt(apiKey.APIKey, existing.ToolName)
	}

	return nil
}
