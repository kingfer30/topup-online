<template>
  <div class="gpt-cards-management">
    <n-card title="GPT卡密管理">
      <template #header-extra>
        <n-space>
          <n-button
            type="warning"
            :disabled="checkedRowKeys.length === 0"
            :loading="checkLoading"
            @click="handleBatchCheck"
          >
            <template #icon><span>🔍</span></template>
            批量检查 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button
            type="error"
            :disabled="checkedRowKeys.length === 0"
            @click="handleBatchDelete"
          >
            批量删除 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button type="primary" @click="showImportModal = true">
            <template #icon><span>📥</span></template>
            批量导入
          </n-button>
        </n-space>
      </template>

      <n-space vertical :size="16">
        <!-- 搜索栏 -->
        <n-space>
          <n-select
            v-model:value="searchSupplier"
            :options="[{ label: '全部供应商', value: '' }, ...supplierOptions]"
            placeholder="供应商"
            style="width: 150px"
          />
          <n-select
            v-model:value="searchStatus"
            :options="statusOptions"
            placeholder="状态"
            style="width: 130px"
          />
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索卡密/邮箱/买家"
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
          :data="cardList"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: GptCard) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>

    <!-- 批量导入弹窗 -->
    <n-modal
      v-model:show="showImportModal"
      preset="card"
      title="批量导入卡密"
      :style="{ width: '540px' }"
      :mask-closable="false"
    >
      <n-form ref="importFormRef" :model="importForm" label-placement="left" label-width="90px">
        <n-form-item label="供应商" path="supplier" :rule="{ required: true, message: '请选择供应商' }">
          <n-select
            v-model:value="importForm.supplier"
            :options="supplierOptions"
            placeholder="请选择供应商"
          />
        </n-form-item>
        <n-form-item label="卡密列表" path="keysText">
          <n-input
            v-model:value="importForm.keysText"
            type="textarea"
            placeholder="每行一条卡密，支持批量粘贴"
            :rows="10"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showImportModal = false">取消</n-button>
          <n-button type="primary" :loading="importLoading" @click="handleImport">
            导入 {{ importKeysCount > 0 ? `(${importKeysCount}条)` : '' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 批量检查结果弹窗 -->
    <n-modal
      v-model:show="showCheckResultModal"
      preset="card"
      title="批量检查结果"
      :style="{ width: '680px' }"
    >
      <n-space vertical>
        <n-alert type="info">
          共检查 {{ checkResult.checked }} 条，其中 {{ checkResult.updated }} 条状态已修正
        </n-alert>
        <n-data-table
          :columns="checkResultColumns"
          :data="checkResult.results"
          :bordered="false"
          :single-line="false"
          size="small"
          max-height="400"
        />
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button type="primary" @click="showCheckResultModal = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      preset="card"
      title="编辑卡密"
      :style="{ width: '480px' }"
      :mask-closable="false"
    >
      <n-form :model="editForm" label-placement="left" label-width="90px">
        <n-form-item label="目的账号">
          <n-input v-model:value="editForm.gpt_mail" placeholder="GPT邮箱账号" />
        </n-form-item>
        <n-form-item label="买家">
          <n-input v-model:value="editForm.buyer" placeholder="买家信息" />
        </n-form-item>
        <n-form-item label="状态">
          <n-select v-model:value="editForm.status" :options="editStatusOptions" />
        </n-form-item>
        <n-form-item label="订阅结果">
          <n-input v-model:value="editForm.sub_result" type="textarea" :rows="3" placeholder="订阅结果" />
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
import { ref, computed, onMounted, h } from 'vue'
import {
  NCard, NSpace, NButton, NInput, NSelect, NDataTable, NModal, NForm, NFormItem,
  NTag, NEllipsis, NAlert, useMessage, type PaginationProps, type FormInst, type DataTableColumns
} from 'naive-ui'
import {
  getSuppliers, getGptCardList, batchImportGptCards, updateGptCard,
  deleteGptCard, batchDeleteGptCards, batchCheckGptCards,
  type GptCard, type CheckCardResult
} from '@/api/gpt-cards'

const message = useMessage()

const loading = ref(false)
const cardList = ref<GptCard[]>([])
const checkedRowKeys = ref<number[]>([])

// 搜索条件
const searchSupplier = ref('')
const searchStatus = ref<number>(0)
const searchKeyword = ref('')

// 供应商列表
const suppliers = ref<string[]>([])
const supplierOptions = computed(() =>
  suppliers.value.map(s => ({ label: s, value: s }))
)

const statusOptions = [
  { label: '全部状态', value: 0 },
  { label: '待使用', value: 1 },
  { label: '已使用', value: 2 },
  { label: '作废', value: -1 },
]

const editStatusOptions = [
  { label: '待使用', value: 1 },
  { label: '已使用', value: 2 },
  { label: '作废', value: -1 },
]

const statusTagMap: Record<number, { type: 'success' | 'info' | 'error' | 'warning', text: string }> = {
  1: { type: 'info', text: '待使用' },
  2: { type: 'success', text: '已使用' },
  [-1]: { type: 'error', text: '作废' },
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
const columns = computed<DataTableColumns<GptCard>>(() => [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 70 },
  { title: '供应商', key: 'supplier', width: 90 },
  {
    title: '卡密',
    key: 'key',
    width: 200,
    render: (row) => h(NEllipsis, { style: 'max-width: 200px' }, { default: () => row.key }),
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row) => {
      const s = statusTagMap[row.status] ?? { type: 'warning' as const, text: '未知' }
      return h(NTag, { type: s.type, size: 'small' }, { default: () => s.text })
    },
  },
  {
    title: '目的账号',
    key: 'gpt_mail',
    width: 160,
    render: (row) => row.gpt_mail || '-',
  },
  { title: '买家', key: 'buyer', width: 120, render: (row) => row.buyer || '-' },
  { title: '导入时间', key: 'import_time', width: 160, render: (row) => formatTime(row.import_time) },
  { title: '使用时间', key: 'use_time', width: 160, render: (row) => formatTime(row.use_time) },
  { title: '订阅开始', key: 'sub_start_time', width: 160, render: (row) => formatTime(row.sub_start_time) },
  { title: '订阅结束', key: 'sub_end_time', width: 160, render: (row) => formatTime(row.sub_end_time) },
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
])

// 加载数据
const loadCards = async () => {
  loading.value = true
  try {
    const res = await getGptCardList({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      supplier: searchSupplier.value || undefined,
      status: searchStatus.value || undefined,
      keyword: searchKeyword.value || undefined,
    })
    if (res.code === 200) {
      cardList.value = res.data.list || []
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
  loadCards()
}

const handleReset = () => {
  searchSupplier.value = ''
  searchStatus.value = 0
  searchKeyword.value = ''
  pagination.value.page = 1
  loadCards()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadCards()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadCards()
}

// 批量导入
const showImportModal = ref(false)
const importLoading = ref(false)
const importFormRef = ref<FormInst | null>(null)
const importForm = ref({ supplier: '', keysText: '' })

const importKeysCount = computed(() =>
  importForm.value.keysText.split('\n').filter(k => k.trim()).length
)

const handleImport = () => {
  importFormRef.value?.validate(async (errors) => {
    if (errors) return
    const keys = importForm.value.keysText.split('\n').map(k => k.trim()).filter(Boolean)
    if (keys.length === 0) {
      message.warning('请输入至少一条卡密')
      return
    }
    importLoading.value = true
    try {
      const res = await batchImportGptCards({ supplier: importForm.value.supplier, keys })
      if (res.code === 200) {
        message.success(`成功导入 ${res.data.imported} 条卡密`)
        showImportModal.value = false
        importForm.value = { supplier: '', keysText: '' }
        await loadCards()
      } else {
        message.error(res.message || '导入失败')
      }
    } catch {
      message.error('导入失败')
    } finally {
      importLoading.value = false
    }
  })
}

// 编辑
const showEditModal = ref(false)
const editLoading = ref(false)
const editingId = ref(0)
const editForm = ref({ gpt_mail: '', buyer: '', status: 1, sub_result: '' })

const handleEdit = (row: GptCard) => {
  editingId.value = row.id
  editForm.value = {
    gpt_mail: row.gpt_mail || '',
    buyer: row.buyer || '',
    status: row.status,
    sub_result: row.sub_result || '',
  }
  showEditModal.value = true
}

const handleEditSave = async () => {
  editLoading.value = true
  try {
    const res = await updateGptCard(editingId.value, {
      gpt_mail: editForm.value.gpt_mail,
      buyer: editForm.value.buyer,
      status: editForm.value.status,
      sub_result: editForm.value.sub_result,
    })
    if (res.code === 200) {
      message.success('更新成功')
      showEditModal.value = false
      await loadCards()
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
const handleDelete = async (row: GptCard) => {
  try {
    const res = await deleteGptCard(row.id)
    if (res.code === 200) {
      message.success('删除成功')
      await loadCards()
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
    const res = await batchDeleteGptCards(checkedRowKeys.value)
    if (res.code === 200) {
      message.success(`已删除 ${checkedRowKeys.value.length} 条`)
      checkedRowKeys.value = []
      await loadCards()
    } else {
      message.error(res.message || '批量删除失败')
    }
  } catch {
    message.error('批量删除失败')
  }
}

// 批量检查
const checkLoading = ref(false)
const showCheckResultModal = ref(false)
const checkResult = ref<{ checked: number; updated: number; results: CheckCardResult[] }>({
  checked: 0,
  updated: 0,
  results: [],
})

const checkResultColumns: DataTableColumns<CheckCardResult> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '卡密',
    key: 'key',
    render: (row) => h(NEllipsis, { style: 'max-width: 200px' }, { default: () => row.key }),
  },
  {
    title: '原状态',
    key: 'old_status',
    width: 90,
    render: (row) => {
      const s = statusTagMap[row.old_status] ?? { type: 'warning' as const, text: '未知' }
      return h(NTag, { type: s.type, size: 'small' }, { default: () => s.text })
    },
  },
  {
    title: '新状态',
    key: 'new_status',
    width: 90,
    render: (row) => {
      const s = statusTagMap[row.new_status] ?? { type: 'warning' as const, text: '未知' }
      return h(NTag, { type: row.changed ? 'warning' : s.type, size: 'small' }, { default: () => s.text })
    },
  },
  {
    title: '是否变更',
    key: 'changed',
    width: 90,
    render: (row) =>
      h(NTag, { type: row.changed ? 'warning' : 'default', size: 'small' },
        { default: () => (row.changed ? '已修正' : '无变化') }),
  },
  { title: '说明', key: 'message' },
]

const handleBatchCheck = async () => {
  if (checkedRowKeys.value.length === 0) return
  checkLoading.value = true
  try {
    const res = await batchCheckGptCards(checkedRowKeys.value)
    if (res.code === 200) {
      checkResult.value = res.data
      showCheckResultModal.value = true
      if (res.data.updated > 0) {
        await loadCards()
      }
    } else {
      message.error(res.message || '批量检查失败')
    }
  } catch {
    message.error('批量检查失败')
  } finally {
    checkLoading.value = false
  }
}

onMounted(async () => {
  // 加载供应商列表
  try {
    const res = await getSuppliers()
    if (res.code === 200) {
      suppliers.value = res.data || []
    }
  } catch {}
  await loadCards()
})
</script>
