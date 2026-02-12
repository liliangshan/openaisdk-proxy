<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <h2>模型管理系统</h2>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据概览</span>
        </el-menu-item>
        
        <el-menu-item index="/providers">
          <el-icon><OfficeBuilding /></el-icon>
          <span>厂商管理</span>
        </el-menu-item>
        
        <el-menu-item index="/models">
          <el-icon><Box /></el-icon>
          <span>模型管理</span>
        </el-menu-item>
        
        <el-menu-item index="/api-keys">
          <el-icon><Key /></el-icon>
          <span>API密钥</span>
        </el-menu-item>
      </el-menu>
      
      <div class="user-info">
        <el-avatar :size="40" :src="userAvatar">
          {{ userInitial }}
        </el-avatar>
        <div class="user-detail">
          <span class="username">{{ authStore.user?.username }}</span>
          <el-link type="danger" underline="never" @click="handleLogout">
            退出登录
          </el-link>
        </div>
      </div>
    </el-aside>
    
    <!-- 主内容区 -->
    <el-container>
      <el-header class="header">
        <div class="breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-tag type="success">已登录</el-tag>
        </div>
      </el-header>
      
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataAnalysis, OfficeBuilding, Box, Key } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 当前激活的菜单
const activeMenu = computed(() => route.path)

// 当前页面标题
const currentPageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '数据概览',
    '/providers': '厂商管理',
    '/models': '模型管理',
    '/api-keys': 'API密钥管理'
  }
  return titles[route.path] || ''
})

// 用户头像
const userAvatar = computed(() => {
  return authStore.user?.username 
    ? `https://api.dicebear.com/7.x/avataaars/svg?seed=${authStore.user.username}`
    : ''
})

// 用户名首字母
const userInitial = computed(() => {
  return authStore.user?.username?.charAt(0).toUpperCase() || 'U'
})

// 退出登录
const handleLogout = () => {
  authStore.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
}

.logo h2 {
  color: white;
  font-size: 16px;
  margin: 0;
}

.el-menu {
  flex: 1;
  border-right: none;
}

.el-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-info {
  padding: 20px;
  border-top: 1px solid #1f2d3d;
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-detail {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.username {
  color: white;
  font-size: 14px;
}

.header {
  background: white;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.breadcrumb {
  display: flex;
  align-items: center;
}

.main {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>
