<template>
  <div>
    <n-space vertical :size="16">
      <!-- 页面标题和操作 -->
      <div class="flex justify-between items-center">
        <div>
          <h1 class="apple-page-title">用户列表</h1>
          <p class="apple-page-subtitle">管理系统所有用户</p>
        </div>
        <n-button type="primary" @click="handleAdd">
          <template #icon>
            <n-icon>➕</n-icon>
          </template>
          添加用户
        </n-button>
      </div>

      <!-- 搜索和筛选 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索用户名、邮箱或显示名"
            clearable
            style="width: 300px"
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

      <!-- 用户表格 -->
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

    <!-- 添加/编辑用户弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="isEdit ? '编辑用户' : '添加用户'"
      style="width: 600px"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="用户名" path="username">
          <n-input
            v-model:value="formValue.username"
            placeholder="请输入用户名（最多12字符）"
            :disabled="isEdit"
            maxlength="12"
          />
        </n-form-item>
        
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="formValue.password"
            type="password"
            show-password-on="click"
            :placeholder="isEdit ? '留空则不修改密码' : '请输入密码（8-20字符）'"
            minlength="8"
            maxlength="20"
          />
        </n-form-item>

        <n-form-item label="显示名" path="display_name">
          <n-input
            v-model:value="formValue.display_name"
            placeholder="请输入显示名（最多20字符）"
            maxlength="20"
          />
        </n-form-item>

        <n-form-item label="邮箱" path="email">
          <n-input
            v-model:value="formValue.email"
            placeholder="请输入邮箱"
            maxlength="50"
          />
        </n-form-item>

        <n-form-item label="状态" path="status">
          <n-select v-model:value="formValue.status" :options="statusOptions" />
        </n-form-item>

        <n-form-item label="用户来源" path="source">
          <n-select v-model:value="formValue.source" :options="sourceOptions" placeholder="请选择用户来源" />
        </n-form-item>

        <template v-if="isEdit">
          <n-form-item label="绑定卡密" path="mirror_card_id">
            <n-space vertical style="width: 100%">
              <n-input
                v-model:value="cardSearchKeyword"
                placeholder="搜索卡密（输入用户名）"
                clearable
                @keyup.enter="searchMirrorCards"
              >
                <template #suffix>
                  <n-button text @click="searchMirrorCards">搜索</n-button>
                </template>
              </n-input>
              <n-select
                v-model:value="formValue.mirror_card_id"
                :options="cardOptions"
                placeholder="选择要绑定的镜像卡密（0表示解绑）"
                clearable
                filterable
                :loading="loadingCards"
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
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
} from 'naive-ui'
import type { User, UpdateUserRequest } from '@/types'
import {
  getUserList,
  createUser,
  updateUser,
  deleteUser,
} from '@/api/admin'
import { getMirrorCardList } from '@/api/mirror-cards'

const message = useMessage()

// 搜索和筛选
const searchKeyword = ref('')

// 状态选项
const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 },
]

// 用户来源选项
const sourceOptions = [
  { label: 'chat', value: 'chat' },
  { label: '充值站', value: '充值站' },
  { label: '手动添加', value: '手动添加' },
]

// 表格数据
const loading = ref(false)
const tableData = ref<User[]>([])
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 表格列定义
const columns: DataTableColumns<User> = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
    sorter: 'default',
  },
  {
    title: '用户名',
    key: 'username',
    width: 150,
  },
  {
    title: '显示名',
    key: 'display_name',
    width: 150,
  },
  {
    title: '邮箱',
    key: 'email',
    width: 200,
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap: Record<number, { text: string; type: any }> = {
        1: { text: '启用', type: 'success' },
        2: { text: '禁用', type: 'error' },
        3: { text: '已删除', type: 'default' },
      }
      const status = statusMap[row.status] || { text: '未知', type: 'default' }
      return h(NTag, { type: status.type }, { default: () => status.text })
    },
  },
  {
    title: '来源',
    key: 'source',
    width: 120,
    render: (row) => row.source || '-',
  },
  {
    title: '最后登录',
    key: 'last_login',
    width: 180,
    render: (row) => row.last_login ? new Date(row.last_login).toLocaleString('zh-CN') : '-',
  },
  {
    title: '邀请码',
    key: 'aff_code',
    width: 120,
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: (row) => {
      return h(
        NSpace,
        { size: 'small' },
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
              NPopconfirm,
              {
                onPositiveClick: () => handleDeleteConfirm(row),
              },
              {
                default: () => '确定删除该用户吗？删除后用户将无法登录。',
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
          ],
        }
      )
    },
  },
]

// 表单相关
const showModal = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInst | null>(null)
const submitting = ref(false)
const currentUserId = ref<number>(0)

const formValue = ref<any>({
  username: '',
  password: '',
  display_name: '',
  email: '',
  status: 1,
  source: '',
  mirror_card_id: 0,
})

// 卡密搜索
const cardSearchKeyword = ref('')
const cardOptions = ref<Array<{ label: string; value: number }>>([
  { label: '解绑（不绑定镜像卡密）', value: 0 },
])
const loadingCards = ref(false)

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 1, max: 12, message: '用户名长度为1-12字符', trigger: 'blur' },
  ],
  password: [
    {
      required: false,
      validator: (_rule, value) => {
        if (isEdit.value && !value) {
          return true // 编辑时密码可为空
        }
        if (!isEdit.value && !value) {
          return new Error('请输入密码')
        }
        if (value && (value.length < 8 || value.length > 20)) {
          return new Error('密码长度为8-20字符')
        }
        return true
      },
      trigger: 'blur',
    },
  ],
  email: [
    {
      required: false,
      validator: (_rule, value) => {
        if (!value) return true
        const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        if (!emailReg.test(value)) {
          return new Error('请输入正确的邮箱格式')
        }
        return true
      },
      trigger: 'blur',
    },
  ],
}

// 加载用户列表
async function loadUserList() {
  loading.value = true
  try {
    const res = await getUserList({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      keyword: searchKeyword.value || undefined,
    }) as any

    if (res.code === 200) {
      tableData.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '获取用户列表失败')
    }
  } catch (error: any) {
    message.error(error.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 操作方法
const handleSearch = () => {
  pagination.value.page = 1
  loadUserList()
}

const handleReset = () => {
  searchKeyword.value = ''
  pagination.value.page = 1
  loadUserList()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadUserList()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadUserList()
}

// 搜索镜像卡密
const searchMirrorCards = async () => {
  if (!cardSearchKeyword.value) {
    message.warning('请输入搜索关键词')
    return
  }
  
  loadingCards.value = true
  try {
    const res = await getMirrorCardList({
      page: 1,
      page_size: 50,
      keyword: cardSearchKeyword.value,
    })
    if (res.code === 200) {
      const cards = res.data.list || []
      cardOptions.value = [
        { label: '解绑（不绑定镜像卡密）', value: 0 },
        ...cards.map((card: any) => ({
          label: `${card.username} (ID: ${card.id})`,
          value: card.id,
        })),
      ]
      if (cards.length === 0) {
        message.info('未找到匹配的卡密')
      }
    } else {
      message.error(res.message || '搜索卡密失败')
    }
  } catch (error: any) {
    message.error(error.message || '搜索卡密失败')
  } finally {
    loadingCards.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  currentUserId.value = 0
  formValue.value = {
    username: '',
    password: '',
    display_name: '',
    email: '',
    status: 1,
    source: '',
    mirror_card_id: 0,
  }
  cardSearchKeyword.value = ''
  cardOptions.value = [{ label: '解绑（不绑定镜像卡密）', value: 0 }]
  showModal.value = true
}

const handleEdit = (row: User) => {
  isEdit.value = true
  currentUserId.value = row.id
  formValue.value = {
    username: row.username,
    password: '', // 编辑时密码留空
    display_name: row.display_name,
    email: row.email,
    status: row.status,
    source: row.source || '',
    mirror_card_id: row.mirror_card_id || 0,
  }
  cardSearchKeyword.value = ''
  // 如果已绑定卡密，添加到选项中
  if (row.mirror_card_id && row.mirror_card_id > 0) {
    cardOptions.value = [
      { label: '解绑（不绑定镜像卡密）', value: 0 },
      { label: `卡密 ID: ${row.mirror_card_id}`, value: row.mirror_card_id },
    ]
  } else {
    cardOptions.value = [{ label: '解绑（不绑定镜像卡密）', value: 0 }]
  }
  showModal.value = true
}

const handleDeleteConfirm = async (row: User) => {
  try {
    const res = await deleteUser(row.id) as any
    if (res.code === 200) {
      message.success('删除用户成功')
      loadUserList()
    } else {
      message.error(res.message || '删除用户失败')
    }
  } catch (error: any) {
    message.error(error.message || '删除用户失败')
  }
}

const handleSubmit = () => {
  formRef.value?.validate(async (errors) => {
    if (errors) {
      message.error('请检查表单填写')
      return
    }

    submitting.value = true
    try {
      if (isEdit.value) {
        // 编辑用户
        const updateData: UpdateUserRequest = {
          display_name: formValue.value.display_name,
          email: formValue.value.email,
          status: formValue.value.status,
        }
        
        // 如果填写了密码，则更新密码
        if (formValue.value.password) {
          updateData.password = formValue.value.password
        }

        const res = await updateUser(currentUserId.value, updateData) as any
        if (res.code === 200) {
          message.success('更新用户成功')
          showModal.value = false
          loadUserList()
        } else {
          message.error(res.message || '更新用户失败')
        }
      } else {
        // 创建用户
        const res = await createUser(formValue.value) as any
        if (res.code === 200) {
          message.success('创建用户成功')
          showModal.value = false
          loadUserList()
        } else {
          message.error(res.message || '创建用户失败')
        }
      }
    } catch (error: any) {
      message.error(error.message || (isEdit.value ? '更新用户失败' : '创建用户失败'))
    } finally {
      submitting.value = false
    }
  })
}

// 初始化加载
onMounted(() => {
  loadUserList()
})
</script>

<style scoped>
:deep(.n-card) {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

:deep(.n-data-table-td) {
  padding: 14px 16px;
}

:deep(.n-modal .n-card) {
  border-radius: 16px !important;
}
</style>
