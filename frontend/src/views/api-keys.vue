<template>
  <div class="api-keys-page">
    <!-- 操作栏 -->
    <el-card shadow="never" class="toolbar">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        生成API密钥
      </el-button>
      <el-button @click="loadAPIKeys">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </el-card>

    <!-- API地址提示 -->
    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="api-url-alert"
    >
      <template #title>
        <span>API地址: {{ apiBaseURL }}</span>
        <el-button
          type="primary"
          link
          size="small"
          @click="copyApiUrl"
          style="margin-left: 12px"
        >
          <el-icon><CopyDocument /></el-icon>
          复制
        </el-button>
      </template>
    </el-alert>

    <!-- 密钥列表 -->
    <el-card shadow="never">
      <el-table
        :data="apiKeys"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="key_name" label="密钥名称" width="150" />
        <el-table-column prop="api_key" label="API密钥" width="320">
          <template #default="{ row }">
            <el-input
              :type="showKeys[row.id] ? 'text' : 'password'"
              :model-value="row.api_key"
              readonly
              size="small"
              style="width: 260px"
            >
              <template #append>
                <el-button size="small" @click="toggleShowKey(row.id)">
                  <el-icon><View v-if="!showKeys[row.id]" /><Hide v-else /></el-icon>
                </el-button>
                <el-button size="small" @click="copyKey(row.api_key)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </template>
            </el-input>
          </template>
        </el-table-column>
        <el-table-column label="默认提示词" width="100">
          <template #default="{ row }">
            <el-button
              v-if="!row.prompt"
              type="primary"
              link
              size="small"
              @click="showEditKeyDialog(row)"
            >
              设置
            </el-button>
            <el-tag v-else type="success" size="small">已设置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="工具提示词" min-width="200">
          <template #default="{ row }">
            <div class="prompts-cell">
              <el-tag
                v-for="p in row.prompts"
                :key="p.id"
                size="small"
                type="info"
                class="prompt-tag"
              >
                {{ p.tool_name }}
              </el-tag>
              <el-button
                type="primary"
                link
                size="small"
                @click="showPromptsDialog(row)"
              >
                管理
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ row.created_at }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showEditKeyDialog(row)">
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-if="!loading && apiKeys.length === 0" description="暂无API密钥" />
    </el-card>
    
    <!-- 复制提示 -->
    <el-dialog
      v-model="copyDialogVisible"
      title="API密钥"
      width="400px"
      center
      :close-on-click-modal="false"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        请立即复制并妥善保管您的API密钥，离开此页面后将无法再次查看。
      </el-alert>
      <el-input
        :model-value="newKey"
        readonly
        size="large"
        class="api-key-input"
      />
      <template #footer>
        <el-button type="primary" @click="copyKeyAndClose">复制并关闭</el-button>
      </template>
    </el-dialog>
    
    <!-- 生成密钥对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="生成API密钥"
      width="450px"
      center
    >
      <el-form :model="form" label-width="100px">
        <el-form-item label="密钥名称">
          <el-input v-model="form.key_name" placeholder="请输入密钥名称" />
        </el-form-item>
        <el-form-item label="提示词">
          <el-input
            v-model="form.prompt"
            type="textarea"
            :rows="3"
            placeholder="可选：设置默认系统提示词，用于指导 AI 助手的行为和回答风格"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ submitLoading ? '生成中...' : '生成' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑密钥对话框 -->
    <el-dialog
      v-model="editKeyDialogVisible"
      title="编辑API密钥"
      width="500px"
      center
    >
      <el-form :model="editKeyForm" label-width="100px">
        <el-form-item label="密钥名称">
          <el-input v-model="editKeyForm.key_name" placeholder="请输入密钥名称" />
        </el-form-item>
        <el-form-item label="默认提示词">
          <el-input
            v-model="editKeyForm.prompt"
            type="textarea"
            :rows="4"
            placeholder="设置默认系统提示词，用于指导 AI 助手的行为和回答风格。留空则不设置提示词。"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editKeyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editKeySubmitLoading" @click="handleUpdateKey">
          {{ editKeySubmitLoading ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 工具提示词管理对话框 -->
    <el-dialog
      v-model="promptsDialogVisible"
      title="工具提示词管理"
      width="700px"
      center
    >
      <div class="prompts-header">
        <span class="key-name">密钥: {{ currentKey?.key_name }}</span>
        <el-button type="primary" size="small" @click="showAddPromptDialog">
          <el-icon><Plus /></el-icon>
          添加提示词
        </el-button>
      </div>
      
      <el-table :data="currentPrompts" stripe style="width: 100%" v-if="currentPrompts.length > 0">
        <el-table-column prop="tool_name" label="工具名" width="180" />
        <el-table-column prop="prompt" label="提示词" min-width="300">
          <template #default="{ row }">
            <el-text line-clamp="2">{{ row.prompt || '无' }}</el-text>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button type="primary" link @click="showEditPromptItemDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDeletePrompt(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-else description="暂无工具提示词" />

      <template #footer>
        <el-button @click="promptsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 添加/编辑工具提示词对话框 -->
    <el-dialog
      v-model="promptItemDialogVisible"
      :title="isEditPromptItem ? '编辑工具提示词' : '添加工具提示词'"
      width="500px"
      center
    >
      <el-form :model="promptItemForm" label-width="80px">
        <el-form-item label="工具名">
          <el-input v-model="promptItemForm.tool_name" placeholder="可选，如 chat、code 等" />
        </el-form-item>
        <el-form-item label="提示词">
          <el-input
            v-model="promptItemForm.prompt"
            type="textarea"
            :rows="5"
            placeholder="设置该工具的专用提示词"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptItemDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="promptItemSubmitLoading" @click="handleSubmitPromptItem">
          {{ promptItemSubmitLoading ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { Plus, Refresh, View, Hide, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiKeyAPI } from '@/api'
import type { APIKey, APIKeyPrompt } from '@/types'

// API基础URL
const apiBaseURL = computed(() => {
  const { protocol, hostname, port } = window.location
  const portStr = port ? `:${port}` : ''
  return `${protocol}//${hostname}${portStr}/api/v1`
})

// 复制API地址
const copyApiUrl = async () => {
  try {
    await navigator.clipboard.writeText(apiBaseURL.value)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 数据
const apiKeys = ref<APIKey[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const copyDialogVisible = ref(false)
const editKeyDialogVisible = ref(false)
const promptsDialogVisible = ref(false)
const promptItemDialogVisible = ref(false)
const submitLoading = ref(false)
const editKeySubmitLoading = ref(false)
const promptItemSubmitLoading = ref(false)
const showKeys = ref<Record<number, boolean>>({})
const newKey = ref('')
const currentKey = ref<APIKey | null>(null)
const currentPrompts = ref<APIKeyPrompt[]>([])
const isEditPromptItem = ref(false)
const editingPromptItem = ref<APIKeyPrompt | null>(null)

// 生成密钥表单
const form = reactive({
  key_name: '',
  prompt: ''
})

// 编辑密钥表单
const editKeyForm = reactive({
  key_name: '',
  prompt: ''
})

// 工具提示词表单
const promptItemForm = reactive({
  tool_name: '',
  prompt: ''
})

// 加载API密钥数据
const loadAPIKeys = async () => {
  loading.value = true
  try {
    apiKeys.value = await apiKeyAPI.list()
  } catch (error) {
    ElMessage.error('加载API密钥数据失败')
  } finally {
    loading.value = false
  }
}

// 显示生成对话框
const showCreateDialog = () => {
  form.key_name = ''
  form.prompt = ''
  dialogVisible.value = true
}

// 提交生成
const handleSubmit = async () => {
  if (!form.key_name.trim()) {
    form.key_name = '默认密钥'
  }

  submitLoading.value = true
  try {
    const key = await apiKeyAPI.create(form.key_name, form.prompt || null)
    newKey.value = key.api_key
    dialogVisible.value = false
    copyDialogVisible.value = true
    await loadAPIKeys()
  } catch (error) {
    ElMessage.error('生成API密钥失败')
  } finally {
    submitLoading.value = false
  }
}

// 显示编辑密钥对话框
const showEditKeyDialog = (apiKey: APIKey) => {
  currentKey.value = apiKey
  editKeyForm.key_name = apiKey.key_name
  editKeyForm.prompt = apiKey.prompt || ''
  editKeyDialogVisible.value = true
}

// 更新密钥
const handleUpdateKey = async () => {
  if (!currentKey.value) return

  editKeySubmitLoading.value = true
  try {
    await apiKeyAPI.update(currentKey.value.id, {
      key_name: editKeyForm.key_name,
      prompt: editKeyForm.prompt || null
    })
    ElMessage.success('更新成功')
    editKeyDialogVisible.value = false
    await loadAPIKeys()
  } catch (error) {
    ElMessage.error('更新失败')
  } finally {
    editKeySubmitLoading.value = false
  }
}

// 切换显示密钥
const toggleShowKey = (id: number) => {
  showKeys.value[id] = !showKeys.value[id]
}

// 复制密钥
const copyKey = async (key: string) => {
  try {
    await navigator.clipboard.writeText(key)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 复制并关闭
const copyKeyAndClose = async () => {
  try {
    await navigator.clipboard.writeText(newKey.value)
    ElMessage.success('已复制到剪贴板')
    copyDialogVisible.value = false
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 删除密钥
const handleDelete = async (apiKey: APIKey) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除密钥 "${apiKey.key_name}" 吗？删除后无法恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await apiKeyAPI.delete(apiKey.id)
    ElMessage.success('删除成功')
    await loadAPIKeys()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 显示工具提示词管理对话框
const showPromptsDialog = (apiKey: APIKey) => {
  currentKey.value = apiKey
  currentPrompts.value = apiKey.prompts || []
  promptsDialogVisible.value = true
}

// 显示添加工具提示词对话框
const showAddPromptDialog = () => {
  isEditPromptItem.value = false
  editingPromptItem.value = null
  promptItemForm.tool_name = ''
  promptItemForm.prompt = ''
  promptItemDialogVisible.value = true
}

// 显示编辑工具提示词对话框
const showEditPromptItemDialog = (prompt: APIKeyPrompt) => {
  isEditPromptItem.value = true
  editingPromptItem.value = prompt
  promptItemForm.tool_name = prompt.tool_name
  promptItemForm.prompt = prompt.prompt
  promptItemDialogVisible.value = true
}

// 提交工具提示词
const handleSubmitPromptItem = async () => {
  if (!currentKey.value) return

  promptItemSubmitLoading.value = true
  try {
    if (isEditPromptItem.value && editingPromptItem.value) {
      // 编辑
      await apiKeyAPI.updatePromptItem(
        currentKey.value.id,
        editingPromptItem.value.id,
        promptItemForm.tool_name,
        promptItemForm.prompt
      )
      ElMessage.success('更新成功')
    } else {
      // 新增
      await apiKeyAPI.createPrompt(
        currentKey.value.id,
        promptItemForm.tool_name,
        promptItemForm.prompt
      )
      ElMessage.success('添加成功')
    }
    
    promptItemDialogVisible.value = false
    
    // 刷新当前密钥的提示词列表
    currentPrompts.value = await apiKeyAPI.listPrompts(currentKey.value.id)
    
    // 同时刷新主列表
    await loadAPIKeys()
  } catch (error: any) {
    ElMessage.error(error?.message || '操作失败')
  } finally {
    promptItemSubmitLoading.value = false
  }
}

// 删除工具提示词
const handleDeletePrompt = async (prompt: APIKeyPrompt) => {
  if (!currentKey.value) return
  
  try {
    await ElMessageBox.confirm(
      `确定要删除工具 "${prompt.tool_name}" 的提示词吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await apiKeyAPI.deletePrompt(currentKey.value.id, prompt.id)
    ElMessage.success('删除成功')
    
    // 刷新当前密钥的提示词列表
    currentPrompts.value = await apiKeyAPI.listPrompts(currentKey.value.id)
    
    // 同时刷新主列表
    await loadAPIKeys()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 初始化
onMounted(() => {
  loadAPIKeys()
})
</script>

<style scoped>
.api-keys-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.toolbar {
  display: flex;
  gap: 10px;
}

.api-url-alert {
  margin-bottom: 4px;
}

.api-url-alert :deep(.el-alert__title) {
  display: flex;
  align-items: center;
}

.api-key-input {
  font-family: 'Courier New', Courier, monospace;
  background-color: #f5f7fa;
  color: #303133;
  letter-spacing: 1px;
}

.prompts-cell {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.prompt-tag {
  cursor: default;
}

.prompts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.prompts-header .key-name {
  font-weight: 500;
  color: #606266;
}
</style>
