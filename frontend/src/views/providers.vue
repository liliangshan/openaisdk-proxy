<template>
  <div class="providers-page">
    <!-- 操作栏 -->
    <el-card shadow="never" class="toolbar">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        添加厂商
      </el-button>
      <el-button @click="loadProviders">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </el-card>
    
    <!-- 厂商列表 -->
    <el-card shadow="never">
      <el-table
        :data="providers"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="厂商名称" width="150" />
        <el-table-column prop="display_name" label="显示名称" width="150" />
        <el-table-column prop="base_url" label="接口地址" min-width="250" show-overflow-tooltip />
        <el-table-column prop="api_prefix" label="API前缀" width="180" />
        <el-table-column prop="api_key" label="API密钥" width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-input
              :type="showApiKeys[row.id] ? 'text' : 'password'"
              :model-value="maskApiKey(row.api_key)"
              readonly
              size="small"
              style="width: 150px"
            >
              <template #append>
                <el-button size="small" @click="toggleShowKey(row.id)">
                  <el-icon><View v-if="!showApiKeys[row.id]" /><Hide v-else /></el-icon>
                </el-button>
              </template>
            </el-input>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showEditDialog(row)">
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-if="!loading && providers.length === 0" description="暂无厂商数据" />
    </el-card>
    
    <!-- 添加/编辑厂商对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑厂商' : '添加厂商'"
      width="500px"
      center
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="厂商名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入厂商名称" :disabled="isEdit" />
        </el-form-item>
        
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="form.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        
        <el-form-item label="接口地址" prop="base_url">
          <el-input v-model="form.base_url" placeholder="例如: https://api.openai.com/v1" />
        </el-form-item>
        
        <el-form-item label="API前缀" prop="api_prefix">
          <el-input v-model="form.api_prefix" placeholder="例如: chat/completions" />
        </el-form-item>
        
        <el-form-item label="API密钥" prop="api_key">
          <el-input
            v-model="form.api_key"
            type="password"
            placeholder="请输入API密钥"
            show-password
          />
        </el-form-item>

      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ submitLoading ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus, Refresh, View, Hide } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { providerAPI } from '@/api'
import type { Provider, CreateProviderRequest } from '@/types'
import { formatDate } from '@/utils/date'

// 数据
const providers = ref<Provider[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const showApiKeys = ref<Record<number, boolean>>({})

// 表单数据
const form = reactive<CreateProviderRequest>({
  name: '',
  display_name: '',
  base_url: '',
  api_prefix: '',
  api_key: ''
})

// 表单引用
const formRef = ref<FormInstance>()

// 验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入厂商名称', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '厂商名称只能包含字母、数字和下划线，且以字母开头', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' }
  ],
  base_url: [
    { required: true, message: '请输入接口地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL地址', trigger: 'blur' }
  ],
  api_prefix: [
    { required: true, message: '请输入API前缀', trigger: 'blur' }
  ],
  api_key: [
    { required: true, message: '请输入API密钥', trigger: 'blur' }
  ]
}

// 加载厂商数据
const loadProviders = async () => {
  loading.value = true
  try {
    providers.value = await providerAPI.list()
  } catch (error) {
    ElMessage.error('加载厂商数据失败')
  } finally {
    loading.value = false
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    name: '',
    display_name: '',
    base_url: '',
    api_prefix: '',
    api_key: ''
  })
  dialogVisible.value = true
}

// 显示编辑对话框
const showEditDialog = (provider: Provider) => {
  isEdit.value = true
  editingId.value = provider.id
  Object.assign(form, {
    name: provider.name,
    display_name: provider.display_name,
    base_url: provider.base_url,
    api_prefix: provider.api_prefix,
    api_key: provider.api_key
  })
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitLoading.value = true
    try {
      if (isEdit.value && editingId.value) {
        await providerAPI.update(editingId.value, form)
        ElMessage.success('更新成功')
      } else {
        await providerAPI.create(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadProviders()
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    } finally {
      submitLoading.value = false
    }
  })
}

// 删除厂商
const handleDelete = async (provider: Provider) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除厂商 "${provider.display_name}" 吗？删除后相关模型将无法使用。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await providerAPI.delete(provider.id)
    ElMessage.success('删除成功')
    await loadProviders()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 切换显示API密钥
const toggleShowKey = (id: number) => {
  showApiKeys.value[id] = !showApiKeys.value[id]
}

// 隐藏API密钥
const maskApiKey = (key: string) => {
  if (key.length <= 8) return '••••••••'
  return key.substring(0, 4) + '••••••' + key.substring(key.length - 4)
}

// 初始化
onMounted(() => {
  loadProviders()
})
</script>

<style scoped>
.providers-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.toolbar {
  display: flex;
  gap: 10px;
}
</style>
