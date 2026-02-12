<template>
  <div class="models-page">
    <!-- 操作栏 -->
    <el-card shadow="never" class="toolbar">
      <div class="toolbar-content">
        <div class="toolbar-left"></div>
        <div class="toolbar-right">
          <el-select
            v-model="selectedProviderId"
            placeholder="选择厂商"
            clearable
            style="width: 200px"
            @change="loadModels"
          >
            <el-option label="全部厂商" :value="0" />
            <el-option
              v-for="provider in providers"
              :key="provider.id"
              :label="provider.display_name"
              :value="provider.id"
            />
          </el-select>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            添加模型
          </el-button>
        </div>
      </div>
    </el-card>
    
    <!-- 模型列表 -->
    <el-card shadow="never">
      <el-table
        :data="models"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="model_id" label="源ID" width="180" />
        <el-table-column label="前缀+别名" width="220">
          <template #default="{ row }">
            <el-input
              :model-value="getModelFullID(row)"
              readonly
              size="small"
              style="width: 160px"
            >
              <template #append>
                <el-button size="small" @click="copyModelID(row)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </template>
            </el-input>
          </template>
        </el-table-column>
        <el-table-column prop="provider_name" label="厂商" width="150" />
        <el-table-column prop="username" label="所属用户" width="120" />
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.is_active"
              @change="toggleActive(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="context_length" label="上下文" width="100">
          <template #default="{ row }">
            {{ row.context_length || 128 }}k
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
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
      
      <el-empty v-if="!loading && models.length === 0" description="暂无模型数据" />
    </el-card>
    
    <!-- 添加/编辑模型对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑模型' : '添加模型'"
      width="500px"
      center
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="厂商" prop="provider_id">
          <el-select
            v-model="form.provider_id"
            placeholder="请选择厂商"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="provider in providers"
              :key="provider.id"
              :label="provider.display_name"
              :value="provider.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="模型ID" prop="model_id">
          <el-input v-model="form.model_id" placeholder="例如: gpt-4" />
        </el-form-item>
        
        <el-form-item label="模型别名" prop="display_name">
          <el-input v-model="form.display_name" placeholder="请输入模型别名" />
        </el-form-item>

        <el-form-item label="上下文长度">
          <el-input-number
            v-model="form.context_length"
            :min="0"
            :max="102400"
            :step="8"
            placeholder="上下文长度（单位：k）"
          />
          <span class="form-tip">单位：k，默认 128k</span>
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="form.is_active" />
          <span class="form-tip">{{ form.is_active ? '启用' : '禁用' }}</span>
        </el-form-item>

        <el-divider content-position="left">Token 压缩配置</el-divider>

        <el-form-item label="启用压缩">
          <el-switch v-model="form.compress_enabled" />
          <span class="form-tip">是否启用消息压缩</span>
        </el-form-item>

        <el-form-item label="截断长度">
          <el-input-number
            v-model="form.compress_truncate_len"
            :min="10"
            :max="10000"
            :step="10"
          />
          <span class="form-tip">截断过长消息的字符数，默认 500</span>
        </el-form-item>

        <el-form-item label="User数量">
          <el-input-number
            v-model="form.compress_user_count"
            :min="1"
            :max="999999"
            :step="1"
          />
          <span class="form-tip">压缩倒数 n 个 user 之前的消息，默认 3</span>
        </el-form-item>

        <el-form-item label="角色类型">
          <el-checkbox-group v-model="form.compress_role_types">
            <el-checkbox label="user" value="user" />
            <el-checkbox label="assistant" value="assistant" />
            <el-checkbox label="tool" value="tool" />
          </el-checkbox-group>
          <span class="form-tip">多选，留空则默认所有非system类型</span>
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
import { Plus, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { modelAPI, providerAPI } from '@/api'
import type { ModelWithDetails, CreateModelRequest, Provider } from '@/types'
import { formatDate } from '@/utils/date'

// 数据
const models = ref<ModelWithDetails[]>([])
const providers = ref<Provider[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const selectedProviderId = ref<number>(0)

// 表单数据
const form = reactive<CreateModelRequest>({
  provider_id: null,
  model_id: '',
  display_name: '',
  context_length: 128,
  is_active: true,
  compress_enabled: true,
  compress_truncate_len: 500,
  compress_user_count: 3,
  compress_role_types: ''
})

// 表单引用
const formRef = ref<FormInstance>()

// 验证规则
const rules: FormRules = {
  provider_id: [
    { required: true, message: '请选择厂商', trigger: 'change', type: 'number' }
  ],
  model_id: [
    { required: true, message: '请输入模型ID', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入模型别名', trigger: 'blur' }
  ]
}

// 加载模型数据
const loadModels = async () => {
  loading.value = true
  try {
    if (selectedProviderId.value > 0) {
      models.value = await modelAPI.list(selectedProviderId.value)
    } else {
      models.value = await modelAPI.list()
    }
  } catch (error) {
    ElMessage.error('加载模型数据失败')
  } finally {
    loading.value = false
  }
}

// 加载厂商数据
const loadProviders = async () => {
  try {
    providers.value = await providerAPI.list()
  } catch (error) {
    console.error('加载厂商数据失败:', error)
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    provider_id: null,
    model_id: '',
    display_name: '',
    context_length: 128,
    is_active: true,
    compress_enabled: true,
    compress_truncate_len: 500,
    compress_user_count: 3,
    compress_role_types: ''
  })
  dialogVisible.value = true
}

// 显示编辑对话框
const showEditDialog = (model: ModelWithDetails) => {
  isEdit.value = true
  editingId.value = model.id
  // compress_role_types 从逗号分隔字符串转为数组
  const roleTypesArray = model.compress_role_types
    ? model.compress_role_types.split(',').filter(Boolean)
    : []
  Object.assign(form, {
    provider_id: model.provider_id,
    model_id: model.model_id,
    display_name: model.display_name,
    context_length: model.context_length || 128,
    is_active: model.is_active,
    compress_enabled: model.compress_enabled ?? true,
    compress_truncate_len: model.compress_truncate_len ?? 500,
    compress_user_count: model.compress_user_count ?? 3,
    compress_role_types: roleTypesArray
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
      // 处理 compress_role_types 多选转逗号分隔
      const submitData = {
        ...form,
        compress_role_types: Array.isArray(form.compress_role_types)
          ? (form.compress_role_types as string[]).join(',')
          : (form.compress_role_types || '')
      }

      if (isEdit.value && editingId.value) {
        await modelAPI.update(editingId.value, submitData)
        ElMessage.success('更新成功')
      } else {
        await modelAPI.create(submitData)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadModels()
    } catch (error) {
      console.error('更新模型失败:', error)
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    } finally {
      submitLoading.value = false
    }
  })
}

// 切换启用状态
const toggleActive = async (model: ModelWithDetails) => {
  try {
    await modelAPI.update(model.id, {
      provider_id: model.provider_id,
      model_id: model.model_id,
      display_name: model.display_name,
      is_active: !model.is_active,
      context_length: model.context_length || 128,
      compress_enabled: model.compress_enabled ?? true,
      compress_truncate_len: model.compress_truncate_len ?? 500,
      compress_user_count: model.compress_user_count ?? 3,
      compress_role_types: model.compress_role_types ?? ''
    })
    model.is_active = !model.is_active
    ElMessage.success(model.is_active ? '已启用' : '已禁用')
  } catch (error) {
    ElMessage.error('状态切换失败')
  }
}

// 删除模型
const handleDelete = async (model: ModelWithDetails) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模型 "${model.display_name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await modelAPI.delete(model.id)
    ElMessage.success('删除成功')
    await loadModels()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 获取完整的模型ID（厂商前缀-模型ID）
const getModelFullID = (model: ModelWithDetails): string => {
  return `${model.provider_api_prefix}-${model.display_name}`
}

// 复制模型ID
const copyModelID = async (model: ModelWithDetails) => {
  try {
    const fullID = getModelFullID(model)
    await navigator.clipboard.writeText(fullID)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 初始化
onMounted(async () => {
  await Promise.all([
    loadModels(),
    loadProviders()
  ])
})
</script>

<style scoped>
.models-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.form-tip {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
</style>
