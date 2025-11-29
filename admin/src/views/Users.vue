<template>
  <div>
    <n-space vertical :size="16">
      <!-- 页面标题和操作 -->
      <div class="flex justify-between items-center">
        <div>
          <h1 class="text-2xl font-bold mb-2">用户列表</h1>
          <p class="text-gray-500">管理系统所有用户</p>
        </div>
        <n-button type="primary" @click="showModal = true">
          添加用户
        </n-button>
      </div>

      <!-- 搜索和筛选 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-space>
          <n-input
            v-model:value="searchText"
            placeholder="搜索用户名或邮箱"
            clearable
            style="width: 300px"
          >
            <template #prefix>
              <n-icon>🔍</n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="statusFilter"
            :options="statusOptions"
            placeholder="状态筛选"
            clearable
            style="width: 150px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>
      </n-card>

      <!-- 用户表格 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-data-table
          :columns="columns"
          :data="tableData"
          :pagination="pagination"
          :bordered="false"
          :loading="loading"
        />
      </n-card>
    </n-space>

    <!-- 添加/编辑用户弹窗 -->
    <n-modal v-model:show="showModal" preset="card" title="添加用户" style="width: 600px">
      <n-form ref="formRef" :model="formValue" :rules="rules">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="formValue.username" placeholder="请输入用户名" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="formValue.email" placeholder="请输入邮箱" />
        </n-form-item>
        <n-form-item label="角色" path="role">
          <n-select v-model:value="formValue.role" :options="roleOptions" />
        </n-form-item>
        <n-form-item label="状态" path="status">
          <n-switch v-model:value="formValue.status">
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import {
  NSpace,
  NCard,
  NButton,
  NInput,
  NSelect,
  NIcon,
  NDataTable,
  NTag,
  NModal,
  NForm,
  NFormItem,
  NSwitch,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'

const message = useMessage()

// 搜索和筛选
const searchText = ref('')
const statusFilter = ref<string | null>(null)

const statusOptions = [
  { label: '全部', value: null },
  { label: '正常', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

const roleOptions = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
]

// 表格数据
const loading = ref(false)
const columns: DataTableColumns = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
  },
  {
    title: '用户名',
    key: 'username',
  },
  {
    title: '邮箱',
    key: 'email',
  },
  {
    title: '角色',
    key: 'role',
    render: (row: any) => {
      return h(NTag, { type: 'info' }, { default: () => row.role === 'admin' ? '管理员' : '普通用户' })
    },
  },
  {
    title: '状态',
    key: 'status',
    render: (row: any) => {
      return h(
        NTag,
        { type: row.status === 'active' ? 'success' : 'error' },
        { default: () => (row.status === 'active' ? '正常' : '禁用') }
      )
    },
  },
  {
    title: '注册时间',
    key: 'createdAt',
  },
  {
    title: '操作',
    key: 'actions',
    render: (row: any) => {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => handleEdit(row),
              },
              { default: () => '编辑' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row),
              },
              { default: () => '删除' }
            ),
          ],
        }
      )
    },
  },
]

const tableData = ref([
  {
    id: 1,
    username: 'admin',
    email: 'admin@example.com',
    role: 'admin',
    status: 'active',
    createdAt: '2025-01-01 10:00:00',
  },
  {
    id: 2,
    username: 'user001',
    email: 'user001@example.com',
    role: 'user',
    status: 'active',
    createdAt: '2025-11-15 14:30:00',
  },
  {
    id: 3,
    username: 'user002',
    email: 'user002@example.com',
    role: 'user',
    status: 'disabled',
    createdAt: '2025-11-20 09:15:00',
  },
])

const pagination = {
  pageSize: 10,
}

// 表单相关
const showModal = ref(false)
const formRef = ref()
const formValue = ref({
  username: '',
  email: '',
  role: 'user',
  status: true,
})

const rules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur',
  },
  email: {
    required: true,
    message: '请输入邮箱',
    trigger: 'blur',
  },
  role: {
    required: true,
    message: '请选择角色',
    trigger: 'change',
  },
}

// 操作方法
const handleSearch = () => {
  loading.value = true
  setTimeout(() => {
    message.success('搜索完成')
    loading.value = false
  }, 500)
}

const handleReset = () => {
  searchText.value = ''
  statusFilter.value = null
  message.info('已重置筛选条件')
}

const handleEdit = (row: any) => {
  message.info(`编辑用户: ${row.username}`)
  // 这里可以打开编辑弹窗并填充数据
}

const handleDelete = (row: any) => {
  message.warning(`删除用户: ${row.username}`)
  // 这里可以添加删除确认逻辑
}

const handleSubmit = () => {
  formRef.value?.validate((errors: any) => {
    if (!errors) {
      message.success('用户添加成功')
      showModal.value = false
      // 这里可以添加提交逻辑
    }
  })
}
</script>

