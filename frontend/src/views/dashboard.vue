<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon providers">
            <el-icon :size="32"><OfficeBuilding /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ providerCount }}</div>
            <div class="stat-label">厂商数量</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon models">
            <el-icon :size="32"><Box /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ modelCount }}</div>
            <div class="stat-label">模型数量</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon keys">
            <el-icon :size="32"><Key /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ apiKeyCount }}</div>
            <div class="stat-label">API密钥</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 最近模型列表 -->
    <el-card shadow="hover" class="recent-models">
      <template #header>
        <div class="card-header">
          <span>最近使用的模型</span>
          <el-button type="primary" link @click="$router.push('/models')">
            查看全部
          </el-button>
        </div>
      </template>
      
      <el-table :data="recentModels" stripe style="width: 100%">
        <el-table-column prop="display_name" label="模型名称" width="200" />
        <el-table-column prop="provider_name" label="厂商" width="150" />
        <el-table-column prop="model_id" label="模型ID" width="200" />
        <el-table-column prop="username" label="所属用户" width="120" />
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
              {{ row.is_active ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-if="recentModels.length === 0" description="暂无模型数据" />
    </el-card>
    
    <!-- 快捷操作 -->
    <el-row :gutter="20" class="quick-actions">
      <el-col :span="8">
        <el-card shadow="hover" class="action-card" @click="$router.push('/providers')">
          <el-icon :size="48" color="#409EFF"><Plus /></el-icon>
          <span>添加厂商</span>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="action-card" @click="$router.push('/models')">
          <el-icon :size="48" color="#67C23A"><Upload /></el-icon>
          <span>添加模型</span>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="action-card" @click="$router.push('/api-keys')">
          <el-icon :size="48" color="#E6A23C"><Key /></el-icon>
          <span>生成密钥</span>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { OfficeBuilding, Box, Key, Plus, Upload } from '@element-plus/icons-vue'
import { modelAPI, providerAPI, apiKeyAPI } from '@/api'
import type { ModelWithDetails } from '@/types'
import { formatDate } from '@/utils/date'

// 数据
const providerCount = ref(0)
const modelCount = ref(0)
const apiKeyCount = ref(0)
const recentModels = ref<ModelWithDetails[]>([])

// 加载数据
onMounted(async () => {
  await Promise.all([
    loadProviders(),
    loadModels(),
    loadAPIKeys()
  ])
})

// 加载厂商数据
const loadProviders = async () => {
  try {
    const providers = await providerAPI.list()
    providerCount.value = providers.length
  } catch (error) {
    console.error('加载厂商数据失败:', error)
  }
}

// 加载模型数据
const loadModels = async () => {
  try {
    const models = await modelAPI.list()
    const modelArray = Array.isArray(models) ? models : []
    modelCount.value = modelArray.length
    recentModels.value = modelArray.slice(0, 5)
  } catch (error) {
    console.error('加载模型数据失败:', error)
  }
}

// 加载API密钥数据
const loadAPIKeys = async () => {
  try {
    const keys = await apiKeyAPI.list()
    const keysArray = Array.isArray(keys) ? keys : []
    apiKeyCount.value = keysArray.length
  } catch (error) {
    console.error('加载API密钥数据失败:', error)
  }
}
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stat-cards {
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.providers {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.models {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.keys {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 14px;
  color: #999;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.recent-models {
  margin-bottom: 20px;
}

.quick-actions {
  margin-top: 20px;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 15px;
  height: 150px;
  cursor: pointer;
  transition: all 0.3s;
}

.action-card:hover {
  transform: translateY(-5px);
}

.action-card span {
  font-size: 16px;
  color: #666;
}
</style>
