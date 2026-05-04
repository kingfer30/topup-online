<template>
  <div class="gpt-cdk-management">
    <n-card title="GPT-CDK管理">
      <template #header-extra>
        <n-space>
          <n-button
            type="error"
            :disabled="checkedRowKeys.length === 0"
            @click="handleBatchDelete"
          >
            批量删除 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button type="primary" @click="showGenerateModal = true">
            <template #icon><span>✨</span></template>
            批量生成
          </n-button>
        </n-space>
      </template>

      <n-space vertical :size="16">
        <!-- 搜索栏 -->
        <n-space>
          <n-select
            v-model:value="searchSellStatus"
            :options="sellStatusOptions"
            placeholder="出售状态"
            style="width: 130px"
          />
          <n-select
            v-model:value="searchUseStatus"
            :options="useStatusOptions"
            placeholder="使用状态"
            style="width: 130px"
          />
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索CDK/邮箱/买家"
            clearable
            style="width: 260px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <!-- 数据表格 -->
        <n-data-table
          remote
          :columns="columns"
          :data="cdkList"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: GptCdk) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
          :scroll-x="1400"
        />
      </n-space>
    </n-card>

    <!-- 批量生成弹窗 -->
    <n-modal
      v-model:show="showGenerateModal"
      preset="card"
      title="批量生成CDK"
      :style="{ width: '420px' }"
      :mask-closable="false"
    >
      <n-form ref="generateFormRef" :model="generateForm" label-placement="left" label-width="90px">
        <n-form-item
          label="生成数量"
          path="count"
          :rule="{ required: true, type: 'number', min: 1, max: 500, message: '数量须在 1~500 之间' }"
        >
          <n-input-number
            v-model:value="generateForm.count"
            :min="1"
            :max="500"
            placeholder="请输入生成数量"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="过期时间">
          <n-date-picker
            v-model:value="generateForm.expireTimeMs"
            type="datetime"
            clearable
            placeholder="不设置则永不过期"
            style="width: 100%"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showGenerateModal = false">取消</n-button>
          <n-button type="primary" :loading="generateLoading" @click="handleGenerate">生成</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      preset="card"
      title="编辑CDK"
      :style="{ width: '500px' }"
      :mask-closable="false"
    >
      <n-form :model="editForm" label-placement="left" label-width="90px">
        <n-form-item label="目的账号">
          <n-input v-model:value="editForm.gpt_mail" placeholder="GPT邮箱账号" />
        </n-form-item>
        <n-form-item label="买家">
          <n-input v-model:value="editForm.buyer" placeholder="买家信息" />
        </n-form-item>
        <n-form-item label="出售状态">
          <n-select v-model:value="editForm.sell_status" :options="editSellStatusOptions" />
        </n-form-item>
        <n-form-item label="使用状态">
          <n-select v-model:value="editForm.use_status" :options="editUseStatusOptions" />
        </n-form-item>
        <n-form-item label="关联卡密ID">
          <n-input-number v-model:value="editForm.card_id" :min="0" placeholder="gpt_cards.id" style="width:100%" />
        </n-form-item>
        <n-form-item label="过期时间">
          <n-date-picker
            v-model:value="editForm.expireTimeMs"
            type="datetime"
            clearable
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="订阅结果">
          <n-input v-model:value="editForm.sub_result" type="textarea" :rows="2" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" :loading="editLoading" @click="handleEditSave">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  NCard, NSpace, NButton, NInput, NSelect, NDataTable, NModal, NForm, NFormItem,
  NInputNumber, NDatePicker, NTag, NEllipsis, useMessage, type PaginationProps,
  type FormInst, type DataTableColumns
} from 'naive-ui'
import {
  getGptCdkList, batchGenerateGptCdk, updateGptCdk, deleteGptCdk,
  batchDeleteGptCdks, type GptCdk
} from '@/api/gpt-cdk'

const message = useMessage()

const loading = ref(false)
const cdkList = ref<GptCdk[]>([])
const checkedRowKeys = ref<number[]>([])

// 搜索条件
const searchSellStatus = ref<number>(0)
const searchUseStatus = ref<number>(0)
const searchKeyword = ref('')

const sellStatusOptions = [
  { label: '全部出售状态', value: 0 },
  { label: '待售', value: 1 },
  { label: '已售', value: 2 },
  { label: '作废', value: -1 },
]

const useStatusOptions = [
  { label: '全部使用状态', value: 0 },
  { label: '未使用', value: 1 },
  { label: '占用中', value: 2 },
  { label: '已使用', value: 3 },
]

const editSellStatusOptions = [
  { label: '待售', value: 1 },
  { label: '已售', value: 2 },
  { label: '作废', value: -1 },
]

const editUseStatusOptions = [
  { label: '未使用', value: 1 },
  { label: '占用中', value: 2 },
  { label: '已使用', value: 3 },
]

const sellStatusTagMap: Record<number, { type: 'info' | 'success' | 'error' | 'warning', text: string }> = {
  1: { type: 'info', text: '待售' },
  2: { type: 'success', text: '已售' },
  [-1]: { type: 'error', text: '作废' },
}

const useStatusTagMap: Record<number, { type: 'default' | 'warning' | 'success', text: string }> = {
  1: { type: 'default', text: '未使用' },
  2: { type: 'warning', text: '占用中' },
  3: { type: 'success', text: '已使用' },
}

// 分页
const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const formatTime = (ts?: number) => {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN', { hour12: false })
}

// 表格列定义
const columns: DataTableColumns<GptCdk> = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 70 },
  {
    title: 'CDK',
    key: 'key',
    width: 220,
    render: (row) => h(NEllipsis, { style: 'max-width: 220px' }, { default: () => row.key }),
  },
  {
    title: '出售状态',
    key: 'sell_status',
    width: 90,
    render: (row) => {
      const s = sellStatusTagMap[row.sell_status] ?? { type: 'warning' as const, text: '未知' }
      return h(NTag, { type: s.type, size: 'small' }, { default: () => s.text })
    },
  },
  {
    title: '使用状态',
    key: 'use_status',
    width: 90,
    render: (row) => {
      const s = useStatusTagMap[row.use_status] ?? { type: 'warning' as const, text: '未知' }
      return h(NTag, { type: s.type as any, size: 'small' }, { default: () => s.text })
    },
  },
  { title: '目的账号', key: 'gpt_mail', width: 160, render: (row) => row.gpt_mail || '-' },
  { title: '买家', key: 'buyer', width: 120, render: (row) => row.buyer || '-' },
  { title: '关联card_id', key: 'card_id', width: 110, render: (row) => row.card_id ?? '-' },
  { title: '过期时间', key: 'expire_time', width: 160, render: (row) => formatTime(row.expire_time) },
  { title: '使用时间', key: 'use_time', width: 160, render: (row) => formatTime(row.use_time) },
  { title: '创建时间', key: 'created_at', width: 160, render: (row) => row.created_at ? new Date(row.created_at).toLocaleString('zh-CN', { hour12: false }) : '-' },
  {
    title: 'IP地址',
    key: 'ip_addr',
    width: 130,
    render: (row) => row.ip_addr || '-',
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
        ],
      }),
  },
]

// 加载数据
const loadCdks = async () => {
  loading.value = true
  try {
    const res = await getGptCdkList({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      sell_status: searchSellStatus.value || undefined,
      use_status: searchUseStatus.value || undefined,
      keyword: searchKeyword.value || undefined,
    })
    if (res.code === 200) {
      cdkList.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '加载失败')
    }
  } catch {
    message.error('加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  loadCdks()
}

const handleReset = () => {
  searchSellStatus.value = 0
  searchUseStatus.value = 0
  searchKeyword.value = ''
  pagination.value.page = 1
  loadCdks()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadCdks()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadCdks()
}

// 批量生成
const showGenerateModal = ref(false)
const generateLoading = ref(false)
const generateFormRef = ref<FormInst | null>(null)
const generateForm = ref({ count: 10, expireTimeMs: null as number | null })

const handleGenerate = () => {
  generateFormRef.value?.validate(async (errors) => {
    if (errors) return
    generateLoading.value = true
    try {
      const expireTime = generateForm.value.expireTimeMs
        ? Math.floor(generateForm.value.expireTimeMs / 1000)
        : undefined
      const res = await batchGenerateGptCdk({
        count: generateForm.value.count,
        expire_time: expireTime,
      })
      if (res.code === 200) {
        message.success(`成功生成 ${res.data.generated} 条CDK`)
        showGenerateModal.value = false
        generateForm.value = { count: 10, expireTimeMs: null }
        await loadCdks()
      } else {
        message.error(res.message || '生成失败')
      }
    } catch {
      message.error('生成失败')
    } finally {
      generateLoading.value = false
    }
  })
}

// 编辑
const showEditModal = ref(false)
const editLoading = ref(false)
const editingId = ref(0)
const editForm = ref({
  gpt_mail: '',
  buyer: '',
  sell_status: 1,
  use_status: 1,
  card_id: null as number | null,
  sub_result: '',
  expireTimeMs: null as number | null,
})

const handleEdit = (row: GptCdk) => {
  editingId.value = row.id
  editForm.value = {
    gpt_mail: row.gpt_mail || '',
    buyer: row.buyer || '',
    sell_status: row.sell_status,
    use_status: row.use_status,
    card_id: row.card_id ?? null,
    sub_result: row.sub_result || '',
    expireTimeMs: row.expire_time ? row.expire_time * 1000 : null,
  }
  showEditModal.value = true
}

const handleEditSave = async () => {
  editLoading.value = true
  try {
    const expireTime = editForm.value.expireTimeMs
      ? Math.floor(editForm.value.expireTimeMs / 1000)
      : undefined
    const res = await updateGptCdk(editingId.value, {
      gpt_mail: editForm.value.gpt_mail,
      buyer: editForm.value.buyer,
      sell_status: editForm.value.sell_status,
      use_status: editForm.value.use_status,
      card_id: editForm.value.card_id ?? undefined,
      sub_result: editForm.value.sub_result,
      expire_time: expireTime,
    })
    if (res.code === 200) {
      message.success('更新成功')
      showEditModal.value = false
      await loadCdks()
    } else {
      message.error(res.message || '更新失败')
    }
  } catch {
    message.error('更新失败')
  } finally {
    editLoading.value = false
  }
}

// 删除
const handleDelete = async (row: GptCdk) => {
  try {
    const res = await deleteGptCdk(row.id)
    if (res.code === 200) {
      message.success('删除成功')
      await loadCdks()
    } else {
      message.error(res.message || '删除失败')
    }
  } catch {
    message.error('删除失败')
  }
}

const handleBatchDelete = async () => {
  if (checkedRowKeys.value.length === 0) return
  try {
    const res = await batchDeleteGptCdks(checkedRowKeys.value)
    if (res.code === 200) {
      message.success(`已删除 ${checkedRowKeys.value.length} 条`)
      checkedRowKeys.value = []
      await loadCdks()
    } else {
      message.error(res.message || '批量删除失败')
    }
  } catch {
    message.error('批量删除失败')
  }
}

onMounted(() => {
  loadCdks()
})
</script>
