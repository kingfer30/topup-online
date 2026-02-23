<template>
  <div>
    <n-space vertical :size="16">
      <!-- 页面标题 -->
      <div>
        <h1 class="apple-page-title">订单列表</h1>
        <p class="apple-page-subtitle">查看和管理所有订单</p>
      </div>

      <!-- 搜索和筛选 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-space>
          <n-input
            v-model:value="searchText"
            placeholder="搜索订单号或用户"
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
            placeholder="订单状态"
            clearable
            style="width: 150px"
          />
          <n-date-picker v-model:value="dateRange" type="daterange" clearable />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleExport">导出</n-button>
        </n-space>
      </n-card>

      <!-- 订单表格 -->
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

    <!-- 订单详情弹窗 -->
    <n-modal v-model:show="showDetailModal" preset="card" title="订单详情" style="width: 700px">
      <n-descriptions :column="2" bordered v-if="currentOrder">
        <n-descriptions-item label="订单号">
          {{ currentOrder.orderNo }}
        </n-descriptions-item>
        <n-descriptions-item label="用户">
          {{ currentOrder.user }}
        </n-descriptions-item>
        <n-descriptions-item label="金额">
          ¥{{ currentOrder.amount }}
        </n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :type="getStatusType(currentOrder.status)">
            {{ getStatusText(currentOrder.status) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="创建时间">
          {{ currentOrder.createdAt }}
        </n-descriptions-item>
        <n-descriptions-item label="完成时间">
          {{ currentOrder.completedAt || '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="备注" :span="2">
          {{ currentOrder.note || '无' }}
        </n-descriptions-item>
      </n-descriptions>
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
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'

const message = useMessage()

// 搜索和筛选
const searchText = ref('')
const statusFilter = ref<string | null>(null)
const dateRange = ref<[number, number] | null>(null)

const statusOptions = [
  { label: '待支付', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
  { label: '已退款', value: 'refunded' },
]

// 表格数据
const loading = ref(false)
const showDetailModal = ref(false)
const currentOrder = ref<any>(null)

const getStatusType = (status: string) => {
  const map: Record<string, any> = {
    pending: 'warning',
    processing: 'info',
    completed: 'success',
    cancelled: 'default',
    refunded: 'error',
  }
  return map[status] || 'default'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待支付',
    processing: '处理中',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款',
  }
  return map[status] || status
}

const columns: DataTableColumns = [
  {
    title: '订单号',
    key: 'orderNo',
    width: 150,
  },
  {
    title: '用户',
    key: 'user',
  },
  {
    title: '金额',
    key: 'amount',
    render: (row: any) => `¥${row.amount}`,
  },
  {
    title: '状态',
    key: 'status',
    render: (row: any) => {
      return h(
        NTag,
        { type: getStatusType(row.status) },
        { default: () => getStatusText(row.status) }
      )
    },
  },
  {
    title: '创建时间',
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
                onClick: () => handleViewDetail(row),
              },
              { default: () => '详情' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'warning',
                disabled: row.status !== 'pending',
                onClick: () => handleCancel(row),
              },
              { default: () => '取消' }
            ),
          ],
        }
      )
    },
  },
]

const tableData = ref([
  {
    orderNo: 'ORD20250001',
    user: 'user001@example.com',
    amount: 99,
    status: 'completed',
    createdAt: '2025-11-29 10:30:25',
    completedAt: '2025-11-29 10:31:00',
    note: '正常订单',
  },
  {
    orderNo: 'ORD20250002',
    user: 'user002@example.com',
    amount: 199,
    status: 'processing',
    createdAt: '2025-11-29 10:28:15',
    completedAt: null,
    note: '',
  },
  {
    orderNo: 'ORD20250003',
    user: 'user003@example.com',
    amount: 299,
    status: 'completed',
    createdAt: '2025-11-29 10:25:10',
    completedAt: '2025-11-29 10:26:30',
    note: '',
  },
  {
    orderNo: 'ORD20250004',
    user: 'user004@example.com',
    amount: 99,
    status: 'refunded',
    createdAt: '2025-11-29 10:20:05',
    completedAt: '2025-11-29 10:00:00',
    note: '用户申请退款',
  },
  {
    orderNo: 'ORD20250005',
    user: 'user005@example.com',
    amount: 399,
    status: 'pending',
    createdAt: '2025-11-29 10:15:30',
    completedAt: null,
    note: '',
  },
])

const pagination = {
  pageSize: 10,
}

// 操作方法
const handleSearch = () => {
  loading.value = true
  setTimeout(() => {
    message.success('搜索完成')
    loading.value = false
  }, 500)
}

const handleExport = () => {
  message.info('导出功能开发中...')
}

const handleViewDetail = (row: any) => {
  currentOrder.value = row
  showDetailModal.value = true
}

const handleCancel = (row: any) => {
  message.warning(`取消订单: ${row.orderNo}`)
}
</script>

<style scoped>
:deep(.n-card) {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}
</style>

