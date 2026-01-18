<template>
  <div>
    <n-space vertical :size="16">
      <!-- 页面标题和操作 -->
      <div class="flex justify-between items-center">
        <div>
          <h1 class="text-2xl font-bold mb-2">镜像卡密管理</h1>
          <p class="text-gray-500">管理镜像账号的卡密信息</p>
        </div>
        <n-space>
          <n-button type="primary" @click="handleAdd">
            <template #icon>
              <n-icon>➕</n-icon>
            </template>
            添加卡密
          </n-button>
          <n-button type="info" @click="handleBatchImport">
            <template #icon>
              <n-icon>📥</n-icon>
            </template>
            批量导入
          </n-button>
        </n-space>
      </div>

      <!-- 搜索和筛选 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索卡密用户名、密码或绑定用户账号"
            clearable
            style="width: 350px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon>🔍</n-icon>
            </template>
          </n-input>
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>
      </n-card>

      <!-- 卡密表格 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-data-table
          :columns="columns"
          :data="tableData"
          :pagination="pagination"
          :bordered="false"
          :loading="loading"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-card>
    </n-space>

    <!-- 添加/编辑卡密弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="isEdit ? '编辑卡密' : '添加卡密'"
      style="width: 600px"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="用户名" path="username">
          <n-input
            v-model:value="formValue.username"
            placeholder="请输入用户名"
          />
        </n-form-item>
        
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="formValue.password"
            type="password"
            show-password-on="click"
            placeholder="请输入密码"
          />
        </n-form-item>

        <template v-if="isEdit">
          <n-form-item label="状态" path="status">
            <n-select v-model:value="formValue.status" :options="statusOptions" />
          </n-form-item>

          <n-form-item label="绑定用户" path="bind_user_id">
            <n-space vertical style="width: 100%">
              <n-input
                v-model:value="userSearchKeyword"
                placeholder="搜索用户（输入用户名/邮箱）"
                clearable
                @keyup.enter="searchUsers"
              >
                <template #suffix>
                  <n-button text @click="searchUsers">搜索</n-button>
                </template>
              </n-input>
              <n-select
                v-model:value="formValue.bind_user_id"
                :options="userOptions"
                placeholder="选择要绑定的用户（0表示解绑）"
                clearable
                filterable
                :loading="loadingUsers"
              />
            </n-space>
          </n-form-item>
        </template>
      </n-form>
      
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 批量导入弹窗 -->
    <n-modal
      v-model:show="showBatchImportModal"
      preset="card"
      title="批量导入卡密"
      style="width: 700px"
      :mask-closable="false"
    >
      <n-space vertical :size="12">
        <n-alert type="info" title="导入格式说明">
          每行一条记录，格式为：<strong>用户名----密码</strong><br />
          例如：<br />
          <code>user001----password123</code><br />
          <code>user002----password456</code>
        </n-alert>

        <n-input
          v-model:value="batchImportData"
          type="textarea"
          placeholder="请按格式输入要导入的卡密数据（每行一条）"
          :rows="10"
        />
      </n-space>
      
      <template #footer>
        <n-space justify="end">
          <n-button @click="showBatchImportModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleBatchImportSubmit">
            开始导入
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
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
  NPopconfirm,
  NAlert,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
} from 'naive-ui'
import type { MirrorCard } from '@/api/mirror-cards'
import {
  getMirrorCardList,
  createMirrorCard,
  updateMirrorCard,
  deleteMirrorCard,
  batchImportMirrorCards,
} from '@/api/mirror-cards'
import { getUserList } from '@/api/admin'

const message = useMessage()

// 搜索和筛选
const searchKeyword = ref('')

// 状态选项
const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 },
]

// 表格数据
const loading = ref(false)
const tableData = ref<MirrorCard[]>([])
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 表格列定义
const columns: DataTableColumns<MirrorCard> = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
  },
  {
    title: '用户名',
    key: 'username',
    width: 150,
  },
  {
    title: '密码',
    key: 'password',
    width: 150,
  },
  {
    title: '绑定状态',
    key: 'bind_status',
    width: 100,
    render: (row) => {
      const statusMap: Record<number, { text: string; type: any }> = {
        0: { text: '未绑定', type: 'default' },
        1: { text: '已绑定', type: 'success' },
      }
      const status = statusMap[row.bind_status] || { text: '未知', type: 'default' }
      return h(NTag, { type: status.type }, { default: () => status.text })
    },
  },
  {
    title: '绑定用户',
    key: 'bind_user_id',
    width: 180,
    render: (row) => {
      if (!row.bind_user_id) return '-'
      const username = row.bind_user_name ? ` (${row.bind_user_name})` : ''
      return `ID: ${row.bind_user_id}${username}`
    },
  },
  {
    title: '绑定时间',
    key: 'bind_time',
    width: 180,
    render: (row) => row.bind_time ? new Date(row.bind_time).toLocaleString('zh-CN') : '-',
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap: Record<number, { text: string; type: any }> = {
        1: { text: '启用', type: 'success' },
        2: { text: '禁用', type: 'error' },
      }
      const status = statusMap[row.status] || { text: '未知', type: 'default' }
      return h(NTag, { type: status.type }, { default: () => status.text })
    },
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row) => new Date(row.created_at).toLocaleString('zh-CN'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    fixed: 'right',
    render: (row) => {
      return h('div', { class: 'flex gap-2' }, [
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => handleCopy(row),
          },
          { default: () => '复制' }
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            onClick: () => handleEdit(row),
          },
          { default: () => '编辑' }
        ),
        h(
          NPopconfirm,
          {
            onPositiveClick: () => handleDelete(row.id),
          },
          {
            default: () => '确定要删除这个卡密吗？',
            trigger: () =>
              h(
                NButton,
                {
                  size: 'small',
                  type: 'error',
                },
                { default: () => '删除' }
              ),
          }
        ),
      ])
    },
  },
]

// 表单
const showModal = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInst | null>(null)
const formValue = ref({
  id: 0,
  username: '',
  password: '',
  status: 1,
  bind_user_id: 0,
})

// 用户搜索
const userSearchKeyword = ref('')
const userOptions = ref<Array<{ label: string; value: number }>>([
  { label: '解绑（不绑定任何用户）', value: 0 },
])
const loadingUsers = ref(false)

// 批量导入
const showBatchImportModal = ref(false)
const batchImportData = ref('')

// 表单验证规则
const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: !isEdit.value, message: '请输入密码', trigger: 'blur' },
  ],
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const res = await getMirrorCardList({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      keyword: searchKeyword.value,
    })
    if (res.code === 200) {
      tableData.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '获取数据失败')
    }
  } catch (error: any) {
    message.error(error.message || '获取数据失败')
  } finally {
    loading.value = false
  }
}

// 搜索用户
const searchUsers = async () => {
  if (!userSearchKeyword.value) {
    message.warning('请输入搜索关键词')
    return
  }
  
  loadingUsers.value = true
  try {
    const res = await getUserList({
      page: 1,
      page_size: 50,
      keyword: userSearchKeyword.value,
    })
    if (res.code === 200) {
      const users = res.data.list || []
      userOptions.value = [
        { label: '解绑（不绑定任何用户）', value: 0 },
        ...users.map((user: any) => ({
          label: `${user.username} (${user.email || 'ID: ' + user.id})`,
          value: user.id,
        })),
      ]
      if (users.length === 0) {
        message.info('未找到匹配的用户')
      }
    } else {
      message.error(res.message || '搜索用户失败')
    }
  } catch (error: any) {
    message.error(error.message || '搜索用户失败')
  } finally {
    loadingUsers.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.value.page = 1
  loadData()
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  pagination.value.page = 1
  loadData()
}

// 翻页
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadData()
}

// 改变每页数量
const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadData()
}

// 添加
const handleAdd = () => {
  isEdit.value = false
  formValue.value = {
    id: 0,
    username: '',
    password: '',
    status: 1,
    bind_user_id: 0,
  }
  userSearchKeyword.value = ''
  userOptions.value = [{ label: '解绑（不绑定任何用户）', value: 0 }]
  showModal.value = true
}

// 编辑
const handleEdit = (row: MirrorCard) => {
  isEdit.value = true
  formValue.value = {
    id: row.id,
    username: row.username,
    password: '', // 编辑时密码留空
    status: row.status,
    bind_user_id: row.bind_user_id || 0,
  }
  userSearchKeyword.value = ''
  // 如果已绑定用户，添加到选项中
  if (row.bind_user_id > 0) {
    userOptions.value = [
      { label: '解绑（不绑定任何用户）', value: 0 },
      { label: `用户 ID: ${row.bind_user_id}`, value: row.bind_user_id },
    ]
  } else {
    userOptions.value = [{ label: '解绑（不绑定任何用户）', value: 0 }]
  }
  showModal.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    submitting.value = true

    if (isEdit.value) {
      // 编辑
      const data: any = {
        status: formValue.value.status,
        bind_user_id: formValue.value.bind_user_id,
      }
      if (formValue.value.username) {
        data.username = formValue.value.username
      }
      if (formValue.value.password) {
        data.password = formValue.value.password
      }

      const res = await updateMirrorCard(formValue.value.id, data)
      if (res.code === 200) {
        message.success('更新成功')
        showModal.value = false
        loadData()
      } else {
        message.error(res.message || '更新失败')
      }
    } else {
      // 新增
      const res = await createMirrorCard({
        username: formValue.value.username,
        password: formValue.value.password,
      })
      if (res.code === 200) {
        message.success('创建成功')
        showModal.value = false
        loadData()
      } else {
        message.error(res.message || '创建失败')
      }
    }
  } catch (error: any) {
    console.error('表单验证失败', error)
  } finally {
    submitting.value = false
  }
}

// 复制卡密
const handleCopy = (row: MirrorCard) => {
  const copyText = `${row.username}----${row.password}`
  
  // 使用 Clipboard API
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(copyText).then(() => {
      message.success('已复制到剪贴板')
    }).catch(() => {
      // 如果失败，使用备用方案
      fallbackCopy(copyText)
    })
  } else {
    // 不支持 Clipboard API，使用备用方案
    fallbackCopy(copyText)
  }
}

// 备用复制方案
const fallbackCopy = (text: string) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    document.execCommand('copy')
    message.success('已复制到剪贴板')
  } catch (err) {
    message.error('复制失败，请手动复制')
  }
  document.body.removeChild(textarea)
}

// 删除
const handleDelete = async (id: number) => {
  try {
    const res = await deleteMirrorCard(id)
    if (res.code === 200) {
      message.success('删除成功')
      loadData()
    } else {
      message.error(res.message || '删除失败')
    }
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 批量导入
const handleBatchImport = () => {
  batchImportData.value = ''
  showBatchImportModal.value = true
}

// 批量导入提交
const handleBatchImportSubmit = async () => {
  if (!batchImportData.value.trim()) {
    message.warning('请输入要导入的数据')
    return
  }

  submitting.value = true
  try {
    const res = await batchImportMirrorCards({
      data: batchImportData.value,
    })
    if (res.code === 200) {
      message.success(res.message || '导入成功')
      showBatchImportModal.value = false
      loadData()
    } else {
      message.error(res.message || '导入失败')
    }
  } catch (error: any) {
    message.error(error.message || '导入失败')
  } finally {
    submitting.value = false
  }
}

// 初始化
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.flex {
  display: flex;
}

.gap-2 {
  gap: 0.5rem;
}

.justify-between {
  justify-content: space-between;
}

.items-center {
  align-items: center;
}

code {
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}
</style>

