// 用户相关类型
export interface User {
  id: number
  username: string
  email: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
}

export interface AuthResponse {
  code: number
  message: string
  data: {
    user: User
    token: string
  }
}

// API密钥类型
export interface APIKey {
  id: number
  user_id: number
  key_name: string
  api_key: string
  prompt?: string | null
  prompts?: APIKeyPrompt[]
  created_at: string
  updated_at: string
}

// API密钥提示词类型
export interface APIKeyPrompt {
  id: number
  api_key_id: number
  tool_name: string
  prompt: string
  created_at: string
  updated_at: string
}

// 厂商类型
export interface Provider {
  id: number
  name: string
  display_name: string
  base_url: string
  api_prefix: string
  api_key: string
  created_at: string
  updated_at: string
}

export interface CreateProviderRequest {
  name: string
  display_name: string
  base_url: string
  api_prefix: string
  api_key: string
}

// 模型类型
export interface Model {
  id: number
  user_id: number
  provider_id: number
  model_id: string
  display_name: string
  is_active: boolean
  context_length?: number
  compress_enabled?: boolean
  compress_truncate_len?: number
  compress_user_count?: number
  compress_role_types?: string
  created_at: string
  updated_at: string
}

export interface ModelWithDetails extends Model {
  provider_name: string
  provider_display_name: string
  provider_base_url: string
  provider_api_prefix: string
  provider_api_key?: string
  username: string
}

export interface CreateModelRequest {
  provider_id?: number | null
  model_id: string
  display_name: string
  context_length?: number
  is_active?: boolean
  compress_enabled?: boolean
  compress_truncate_len?: number
  compress_user_count?: number
  compress_role_types?: string
}

// 通用响应类型
export interface Response<T = any> {
  code: number
  message: string
  data?: T
}

// 列表响应类型
export interface ListResponse<T> {
  code: number
  message: string
  data: T[]
}
