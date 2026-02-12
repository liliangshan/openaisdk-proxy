import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { authAPI } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)
  const loading = ref(false)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)

  // 方法
  async function login(username: string, password: string): Promise<boolean> {
    try {
      loading.value = true
      const response = await authAPI.login({ username, password })
      
      // 检查响应
      if (response.code === 0 && response.data) {
        token.value = response.data.token
        user.value = response.data.user
        
        // 持久化存储
        localStorage.setItem('token', token.value)
        return true
      } else {
        return false
      }
    } catch (error) {
      console.error('登录失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(username: string, password: string, email: string): Promise<boolean> {
    try {
      loading.value = true
      const response = await authAPI.register({ username, password, email })
      
      if (response.code === 0 && response.data) {
        token.value = response.data.token
        user.value = response.data.user
        
        // 持久化存储
        localStorage.setItem('token', token.value)
        return true
      } else {
        return false
      }
    } catch (error) {
      console.error('注册失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchUser(): Promise<void> {
    if (!token.value) return
    
    try {
      loading.value = true
      user.value = await authAPI.getProfile()
    } catch (error) {
      console.error('获取用户信息失败:', error)
      logout()
    } finally {
      loading.value = false
    }
  }

  function logout(): void {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  // 初始化时获取用户信息
  if (token.value) {
    fetchUser()
  }

  return {
    token,
    user,
    loading,
    isLoggedIn,
    login,
    register,
    fetchUser,
    logout
  }
})
