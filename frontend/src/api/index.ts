import { request } from './axios'
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  APIKey,
  APIKeyPrompt,
  Provider,
  CreateProviderRequest,
  Model,
  ModelWithDetails,
  CreateModelRequest,
  User
} from '@/types'

// 认证相关 API
export const authAPI = {
  // 登录
  async login(data: LoginRequest): Promise<AuthResponse> {
    return request.post<any>('/auth/login', data)
  },
  
  // 注册
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return request.post<any>('/auth/register', data)
  },
  
  // 获取当前用户信息
  getProfile(): Promise<User> {
    return request.get('/auth/profile')
  }
}

// API密钥相关 API
export const apiKeyAPI = {
  // 获取当前用户的API密钥列表
  async list(): Promise<APIKey[]> {
    const response = await request.get<any>('/api-keys')
    // response 格式: {code: 0, message: "...", data: [...]}
    if (response && response.data && Array.isArray(response.data)) {
      return response.data
    }
    return []
  },

  // 创建API密钥
  async create(keyName: string, prompt?: string | null): Promise<APIKey> {
    const data: Record<string, any> = { key_name: keyName }
    if (prompt) {
      data.prompt = prompt
    }
    const response = await request.post<any>('/api-keys', data)
    // response 格式: {code: 0, message: "...", data: APIKey对象}
    if (response && response.data) {
      return response.data
    }
    throw new Error('创建失败')
  },

  // 更新API密钥
  async update(id: number, data: { key_name?: string; prompt?: string | null }): Promise<APIKey> {
    const response = await request.put<any>(`/api-keys/${id}`, data)
    if (response && response.data) {
      return response.data
    }
    throw new Error('更新失败')
  },

  // 删除API密钥
  async delete(id: number): Promise<void> {
    await request.delete(`/api-keys/${id}`)
  },

  // 更新提示词
  async updatePrompt(id: number, prompt?: string | null): Promise<void> {
    await request.put(`/api-keys/${id}/prompt`, { prompt: prompt || '' })
  },

  // ========== 提示词管理 ==========

  // 获取API密钥的所有提示词
  async listPrompts(apiKeyId: number): Promise<APIKeyPrompt[]> {
    const response = await request.get<any>(`/api-keys/${apiKeyId}/prompts`)
    if (response && response.data && Array.isArray(response.data)) {
      return response.data
    }
    return []
  },

  // 创建提示词
  async createPrompt(apiKeyId: number, toolName: string, prompt: string): Promise<APIKeyPrompt> {
    const response = await request.post<any>(`/api-keys/${apiKeyId}/prompts`, {
      tool_name: toolName,
      prompt: prompt
    })
    if (response && response.data) {
      return response.data
    }
    throw new Error('创建失败')
  },

  // 更新提示词
  async updatePromptItem(apiKeyId: number, promptId: number, toolName: string, prompt: string): Promise<APIKeyPrompt> {
    const response = await request.put<any>(`/api-keys/${apiKeyId}/prompts/${promptId}`, {
      tool_name: toolName,
      prompt: prompt
    })
    if (response && response.data) {
      return response.data
    }
    throw new Error('更新失败')
  },

  // 删除提示词
  async deletePrompt(apiKeyId: number, promptId: number): Promise<void> {
    await request.delete(`/api-keys/${apiKeyId}/prompts/${promptId}`)
  }
}

// 厂商相关 API
export const providerAPI = {
  // 获取厂商列表
  async list(): Promise<Provider[]> {
    const response = await request.get<any>('/providers')
    // response 格式: {code: 0, message: "...", data: [...]}
    if (response && response.data && Array.isArray(response.data)) {
      return response.data
    }
    return []
  },

  // 获取单个厂商
  async get(id: number): Promise<Provider> {
    return request.get<any>(`/providers/${id}`)
  },

  // 创建厂商
  async create(data: CreateProviderRequest): Promise<Provider> {
    const response = await request.post<any>('/providers', data)
    if (response && response.data) {
      return response.data
    }
    throw new Error('创建失败')
  },

  // 更新厂商
  async update(id: number, data: Partial<CreateProviderRequest>): Promise<Provider> {
    const response = await request.put<any>(`/providers/${id}`, data)
    if (response && response.data) {
      return response.data
    }
    throw new Error('更新失败')
  },

  // 删除厂商
  async delete(id: number): Promise<void> {
    await request.delete(`/providers/${id}`)
  }
}

// 模型相关 API
export const modelAPI = {
  // 获取当前用户的模型列表
  async list(providerId?: number): Promise<ModelWithDetails[]> {
    let url = '/models'
    if (providerId && providerId > 0) {
      url += `?provider_id=${providerId}`
    }
    const response = await request.get<any>(url)
    // response 格式: {code: 0, message: "...", data: [...]}
    if (response && response.data && Array.isArray(response.data)) {
      return response.data
    }
    return []
  },

  // 获取所有模型（管理员）
  async listAll(providerId?: number): Promise<ModelWithDetails[]> {
    let url = '/admin/models'
    if (providerId && providerId > 0) {
      url += `?provider_id=${providerId}`
    }
    const response = await request.get<any>(url)
    if (response && response.data && Array.isArray(response.data)) {
      return response.data
    }
    return []
  },

  // 获取单个模型
  async get(id: number): Promise<ModelWithDetails> {
    return request.get<any>(`/models/${id}`)
  },

  // 创建模型
  async create(data: CreateModelRequest): Promise<Model> {
    const response = await request.post<any>('/models', data)
    if (response && response.data) {
      return response.data
    }
    throw new Error('创建失败')
  },

  // 更新模型
  async update(id: number, data: Partial<CreateModelRequest> & { is_active?: boolean }): Promise<Model> {
    const response = await request.put<any>(`/models/${id}`, data)
    if (response && response.code === 0) {
      return response.data || {}
    }
    throw new Error(response?.message || '更新失败')
  },

  // 删除模型
  async delete(id: number): Promise<void> {
    await request.delete(`/models/${id}`)
  },

  // 刷新模型缓存
  async refreshCache(): Promise<void> {
    await request.post('/admin/models/refresh')
  }
}
