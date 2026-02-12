package cache

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/model-system/api/internal/models"
)

// ModelCacheItem 模型缓存项
type ModelCacheItem struct {
	Model              models.Model
	ProviderName       string
	ProviderDisplayName string
	ProviderBaseURL     string
	ProviderAPIPrefix   string
	Username           string
	ProviderKey        string
}

// APIKeyCacheItem API密钥缓存项
type APIKeyCacheItem struct {
	UserID   uint64
	Prompt   string
	Prompts  map[string]string // 工具名 -> 提示词
}

// MemoryCache 内存缓存
type MemoryCache struct {
	mu              sync.RWMutex
	models          map[uint64]*ModelCacheItem                    // ID -> ModelCacheItem
	modelsByKey     map[string]*ModelCacheItem                     // provider_prefix-model_id -> ModelCacheItem
	modelsByUser    map[uint64]map[uint64]*ModelCacheItem          // user_id -> model_id -> ModelCacheItem
	apiKeys         map[string]*APIKeyCacheItem                    // api_key -> APIKeyCacheItem
	lastUpdate      time.Time
}

// Cache 单例缓存实例
var cache *MemoryCache

// InitCache 初始化缓存
func InitCache() {
	cache = &MemoryCache{
		models:      make(map[uint64]*ModelCacheItem),
		modelsByKey: make(map[string]*ModelCacheItem),
		modelsByUser: make(map[uint64]map[uint64]*ModelCacheItem),
		apiKeys:     make(map[string]*APIKeyCacheItem),
	}
}

// GetCache 获取缓存实例
func GetCache() *MemoryCache {
	return cache
}

// sortModelsByIDDesc 按 ID 降序排序模型切片
func sortModelsByIDDesc(items []*ModelCacheItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Model.ID > items[j].Model.ID
	})
}

// generateCacheKey 生成缓存键：厂商前缀-模型别名
func generateCacheKey(providerPrefix, displayName string) string {
	return fmt.Sprintf("%s-%s", providerPrefix, displayName)
}

// newModelCacheItem 创建新的 ModelCacheItem
func newModelCacheItem(detail models.ModelWithDetails) *ModelCacheItem {
	return &ModelCacheItem{
		Model:               detail.Model,
		ProviderName:        detail.ProviderName,
		ProviderDisplayName: detail.ProviderDisplayName,
		ProviderBaseURL:     detail.ProviderBaseURL,
		ProviderAPIPrefix:   detail.ProviderAPIPrefix,
		Username:            detail.Username,
		ProviderKey:         detail.ProviderKey,
	}
}

// LoadModels 加载所有模型到缓存
func (c *MemoryCache) LoadModels(modelDetails []models.ModelWithDetails) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.models = make(map[uint64]*ModelCacheItem, len(modelDetails))
	c.modelsByKey = make(map[string]*ModelCacheItem, len(modelDetails))
	c.modelsByUser = make(map[uint64]map[uint64]*ModelCacheItem)
	c.lastUpdate = time.Now()

	for i := range modelDetails {
		detail := modelDetails[i]
		item := newModelCacheItem(detail)

		// 按ID缓存
		c.models[detail.ID] = item

		// 按厂商前缀-模型别名缓存
		cacheKey := generateCacheKey(detail.ProviderAPIPrefix, detail.DisplayName)
		c.modelsByKey[cacheKey] = item

		// 按用户分组
		if _, ok := c.modelsByUser[detail.UserID]; !ok {
			c.modelsByUser[detail.UserID] = make(map[uint64]*ModelCacheItem)
		}
		c.modelsByUser[detail.UserID][detail.ID] = item
	}
}

// AddModel 添加模型到缓存
func (c *MemoryCache) AddModel(model *models.ModelWithDetails) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := newModelCacheItem(*model)

	// 按ID缓存
	c.models[model.ID] = item

	// 按厂商前缀-模型别名缓存
	cacheKey := generateCacheKey(model.ProviderAPIPrefix, model.DisplayName)
	c.modelsByKey[cacheKey] = item

	c.lastUpdate = time.Now()

	// 按用户分组
	if _, ok := c.modelsByUser[model.UserID]; !ok {
		c.modelsByUser[model.UserID] = make(map[uint64]*ModelCacheItem)
	}
	c.modelsByUser[model.UserID][model.ID] = item
}

// UpdateModel 更新缓存中的模型
func (c *MemoryCache) UpdateModel(model *models.ModelWithDetails) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果模型别名或厂商前缀改变，需要删除旧的缓存键
	if existing, ok := c.models[model.ID]; ok {
		// 从用户的模型映射中移除旧信息
		delete(c.modelsByUser[existing.Model.UserID], model.ID)

		// 删除旧的厂商前缀-模型别名缓存键
		oldCacheKey := generateCacheKey(existing.ProviderAPIPrefix, existing.Model.DisplayName)
		delete(c.modelsByKey, oldCacheKey)
	}

	item := newModelCacheItem(*model)

	// 按ID缓存
	c.models[model.ID] = item

	// 按厂商前缀-模型别名缓存
	cacheKey := generateCacheKey(model.ProviderAPIPrefix, model.DisplayName)
	c.modelsByKey[cacheKey] = item

	c.lastUpdate = time.Now()

	// 按用户分组
	if _, ok := c.modelsByUser[model.UserID]; !ok {
		c.modelsByUser[model.UserID] = make(map[uint64]*ModelCacheItem)
	}
	c.modelsByUser[model.UserID][model.ID] = item
}

// DeleteModel 从缓存中删除模型
func (c *MemoryCache) DeleteModel(modelID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.models[modelID]; ok {
		// 从用户的模型映射中移除
		delete(c.modelsByUser[existing.Model.UserID], modelID)

		// 删除厂商前缀-模型别名缓存键
		oldCacheKey := generateCacheKey(existing.ProviderAPIPrefix, existing.Model.DisplayName)
		delete(c.modelsByKey, oldCacheKey)
	}

	// 删除ID缓存
	delete(c.models, modelID)
	c.lastUpdate = time.Now()
}

// GetModel 获取单个模型（按ID）
func (c *MemoryCache) GetModel(modelID uint64) (*ModelCacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	model, ok := c.models[modelID]
	return model, ok
}

// GetModelByKey 根据厂商前缀和模型别名获取模型
func (c *MemoryCache) GetModelByKey(providerPrefix, displayName string) (*ModelCacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheKey := generateCacheKey(providerPrefix, displayName)
	model, ok := c.modelsByKey[cacheKey]
	return model, ok
}

// GetModelByCacheKey 根据完整缓存键（如 "mm-gpt-4"）获取模型
func (c *MemoryCache) GetModelByCacheKey(cacheKey string) (*ModelCacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	model, ok := c.modelsByKey[cacheKey]
	return model, ok
}

// GetModelsByUser 获取用户的所有模型
func (c *MemoryCache) GetModelsByUser(userID uint64) []*ModelCacheItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if models, ok := c.modelsByUser[userID]; ok {
		result := make([]*ModelCacheItem, 0, len(models))
		for _, m := range models {
			result = append(result, m)
		}
		// 按 ID 降序排序，保持与数据库查询一致
		sortModelsByIDDesc(result)
		return result
	}
	return nil
}

// GetAllModels 获取所有模型
func (c *MemoryCache) GetAllModels() []*ModelCacheItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*ModelCacheItem, 0, len(c.models))
	for _, m := range c.models {
		result = append(result, m)
	}
	// 按 ID 降序排序，保持与数据库查询一致
	sortModelsByIDDesc(result)
	return result
}

// GetLastUpdate 获取最后更新时间
func (c *MemoryCache) GetLastUpdate() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastUpdate
}

// GetModelCount 获取模型数量
func (c *MemoryCache) GetModelCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.models)
}

// Refresh 刷新缓存
func (c *MemoryCache) Refresh(modelDetails []models.ModelWithDetails) {
	c.LoadModels(modelDetails)
}

// LoadAPIKeys 加载所有API密钥到缓存
func (c *MemoryCache) LoadAPIKeys(apiKeysWithUsers []models.APIKeyWithUser) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.apiKeys = make(map[string]*APIKeyCacheItem, len(apiKeysWithUsers))
	for _, item := range apiKeysWithUsers {
		c.apiKeys[item.APIKey] = &APIKeyCacheItem{
			UserID:  item.UserID,
			Prompt:  item.Prompt,
			Prompts: make(map[string]string), // 预分配以避免后续动态扩展
		}
	}
}

// LoadAPIKeyPrompts 加载所有API密钥工具提示词到缓存
func (c *MemoryCache) LoadAPIKeyPrompts(apiKeyIDToKey map[uint64]string, prompts []models.APIKeyPrompt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, p := range prompts {
		if apiKey, ok := apiKeyIDToKey[p.APIKeyID]; ok {
			if item, exists := c.apiKeys[apiKey]; exists {
				if item.Prompts == nil {
					item.Prompts = make(map[string]string)
				}
				item.Prompts[p.ToolName] = p.Prompt
			}
		}
	}
}

// AddAPIKey 添加API密钥到缓存
func (c *MemoryCache) AddAPIKey(apiKey string, userID uint64, prompt string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.apiKeys[apiKey] = &APIKeyCacheItem{
		UserID:  userID,
		Prompt:  prompt,
		Prompts: make(map[string]string),
	}
}

// DeleteAPIKey 从缓存中删除API密钥
func (c *MemoryCache) DeleteAPIKey(apiKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.apiKeys, apiKey)
}

// GetUserIDByAPIKey 根据API密钥获取用户ID
func (c *MemoryCache) GetUserIDByAPIKey(apiKey string) (uint64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.apiKeys[apiKey]; ok {
		return item.UserID, true
	}
	return 0, false
}

// GetAPIKeyPrompt 根据API密钥获取提示词
func (c *MemoryCache) GetAPIKeyPrompt(apiKey string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.apiKeys[apiKey]; ok {
		return item.Prompt, true
	}
	return "", false
}

// GetAPIKeyCount 获取API密钥数量
func (c *MemoryCache) GetAPIKeyCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.apiKeys)
}

// AddAPIKeyPrompt 添加或更新 API 密钥工具提示词缓存
func (c *MemoryCache) AddAPIKeyPrompt(apiKey string, toolName string, prompt string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.apiKeys[apiKey]; ok {
		if item.Prompts == nil {
			item.Prompts = make(map[string]string)
		}
		item.Prompts[toolName] = prompt
	}
}

// UpdateAPIKeyPrompt 已弃用，请使用 AddAPIKeyPrompt
// 保留此方法向后兼容
func (c *MemoryCache) UpdateAPIKeyPrompt(apiKey string, toolName string, prompt string) {
	c.AddAPIKeyPrompt(apiKey, toolName, prompt)
}

// DeleteAPIKeyPrompt 删除API密钥工具提示词缓存
func (c *MemoryCache) DeleteAPIKeyPrompt(apiKey string, toolName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.apiKeys[apiKey]; ok {
		delete(item.Prompts, toolName)
	}
}

// GetAPIKeyPromptByTool 根据API密钥和工具名获取提示词
func (c *MemoryCache) GetAPIKeyPromptByTool(apiKey string, toolName string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.apiKeys[apiKey]; ok {
		if prompt, exists := item.Prompts[toolName]; exists {
			return prompt, true
		}
		// 如果没有找到工具提示词，返回默认提示词
		if item.Prompt != "" {
			return item.Prompt, true
		}
	}
	return "", false
}

// GetAPIKeyAllPrompts 获取API密钥的所有提示词（包括默认提示词）
func (c *MemoryCache) GetAPIKeyAllPrompts(apiKey string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string)
	if item, ok := c.apiKeys[apiKey]; ok {
		// 复制工具提示词
		for k, v := range item.Prompts {
			result[k] = v
		}
		// 如果有默认提示词且没有空键的提示词，也加入
		if item.Prompt != "" {
			if _, exists := result[""]; !exists {
				result[""] = item.Prompt
			}
		}
	}
	return result
}
