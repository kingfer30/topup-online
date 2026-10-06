<template>
  <div class="digiseller-orders">
    <n-card title="Digiseller 订单管理">
      <template #header-extra>
        <n-button type="primary" :loading="syncing" @click="handleSync">同步订单</n-button>
      </template>
      <n-space vertical :size="16">
        <n-space align="center">
          <n-input
            v-model:value="searchKeyword"
            placeholder="发票号 / 唯一码 / 邮箱 / 商品ID"
            clearable
            style="width: 280px"
            @keyup.enter="handleSearch"
          />
          <n-select
            v-model:value="searchState"
            :options="stateOptions"
            clearable
            placeholder="唯一码状态"
            style="width: 180px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <n-data-table
          remote
          :columns="columns"
          :data="list"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: DigisellerOrder) => row.id"
          :scroll-x="1680"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>

    <n-modal v-model:show="showDetail" preset="card" title="订单详情" :style="{ width: '760px' }">
      <n-descriptions v-if="current" :column="2" bordered label-placement="left" size="small">
        <n-descriptions-item label="发票号">{{ current.inv }}</n-descriptions-item>
        <n-descriptions-item label="唯一码">{{ current.unique_code }}</n-descriptions-item>
        <n-descriptions-item label="商品 ID">{{ current.id_goods || '—' }}</n-descriptions-item>
        <n-descriptions-item label="数量">{{ current.cnt_goods || '—' }}</n-descriptions-item>
        <n-descriptions-item label="实收金额">{{ formatAmount(current) }}</n-descriptions-item>
        <n-descriptions-item label="等值 USD">{{ formatUsd(current.amount_usd) }}</n-descriptions-item>
        <n-descriptions-item label="净收益">{{ current.profit || '—' }}</n-descriptions-item>
        <n-descriptions-item label="买家邮箱">{{ current.email || '—' }}</n-descriptions-item>
        <n-descriptions-item label="代理商">{{ current.agent_id || '—' }}</n-descriptions-item>
        <n-descriptions-item label="代理佣金">{{ current.agent_percent || '—' }}</n-descriptions-item>
        <n-descriptions-item label="优惠码">{{ current.promo_code || '—' }}</n-descriptions-item>
        <n-descriptions-item label="赠送码">{{ current.bonus_code || '—' }}</n-descriptions-item>
        <n-descriptions-item label="购物车 UID" :span="2">{{ current.cart_uid || '—' }}</n-descriptions-item>
        <n-descriptions-item label="唯一码状态">{{ stateLabel(current.uc_state) }}</n-descriptions-item>
        <n-descriptions-item label="支付时间">{{ fmtTime(current.date_pay) }}</n-descriptions-item>
        <n-descriptions-item label="核验时间">{{ fmtTime(current.uc_date_check) }}</n-descriptions-item>
        <n-descriptions-item label="交付时间">{{ fmtTime(current.uc_date_delivery) }}</n-descriptions-item>
        <n-descriptions-item label="确认时间">{{ fmtTime(current.uc_date_confirmed) }}</n-descriptions-item>
        <n-descriptions-item label="驳回时间">{{ fmtTime(current.uc_date_refuted) }}</n-descriptions-item>
        <n-descriptions-item label="入库时间">{{ fmtTime(current.created_at) }}</n-descriptions-item>
        <n-descriptions-item label="更新时间">{{ fmtTime(current.updated_at) }}</n-descriptions-item>
        <n-descriptions-item label="附加参数" :span="2">
          <pre class="options-json">{{ prettyOptions(current.options_json) }}</pre>
        </n-descriptions-item>
      </n-descriptions>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NTag,
  useMessage,
  type DataTableColumns,
  type PaginationProps,
} from 'naive-ui'
import { getDigisellerOrders, syncDigisellerOrders, type DigisellerOrder } from '@/api/digiseller'

const message = useMessage()
const loading = ref(false)
const syncing = ref(false)
const list = ref<DigisellerOrder[]>([])
const searchKeyword = ref('')
const searchState = ref<number | null>(null)
const showDetail = ref(false)
const current = ref<DigisellerOrder | null>(null)

const stateOptions = [
  { label: '未验证', value: 1 },
  { label: '已交付待确认', value: 2 },
  { label: '已确认', value: 3 },
  { label: '已驳回', value: 4 },
  { label: '已验证未交付', value: 5 },
]

const stateMap: Record<number, { label: string; type: 'default' | 'info' | 'success' | 'warning' | 'error' }> = {
  1: { label: '未验证', type: 'warning' },
  2: { label: '已交付待确认', type: 'info' },
  3: { label: '已确认', type: 'success' },
  4: { label: '已驳回', type: 'error' },
  5: { label: '已验证未交付', type: 'default' },
}

const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

const stateLabel = (state: number) => stateMap[state]?.label || (state ? String(state) : '—')

const fmtTime = (value?: string | null) => {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN')
}

const formatAmount = (row: DigisellerOrder) => {
  const amount = row.amount ?? 0
  return row.type_curr ? `${amount} ${row.type_curr}` : String(amount)
}

const formatUsd = (value?: number) => (value == null ? '—' : value.toFixed(2))

const prettyOptions = (raw?: string) => {
  if (!raw) return '—'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

const openDetail = (row: DigisellerOrder) => {
  current.value = row
  showDetail.value = true
}

const columns: DataTableColumns<DigisellerOrder> = [
  { title: '发票号', key: 'inv', width: 120 },
  { title: '唯一码', key: 'unique_code', width: 170 },
  { title: '商品 ID', key: 'id_goods', width: 110 },
  {
    title: '实收',
    key: 'amount',
    width: 130,
    render: (row) => formatAmount(row),
  },
  {
    title: 'USD',
    key: 'amount_usd',
    width: 90,
    render: (row) => formatUsd(row.amount_usd),
  },
  { title: '净收益', key: 'profit', width: 100 },
  { title: '买家邮箱', key: 'email', width: 200, ellipsis: { tooltip: true } },
  { title: '数量', key: 'cnt_goods', width: 70 },
  {
    title: '状态',
    key: 'uc_state',
    width: 140,
    render: (row) => {
      const item = stateMap[row.uc_state] || { label: stateLabel(row.uc_state), type: 'default' as const }
      return h(NTag, { type: item.type, size: 'small' }, { default: () => item.label })
    },
  },
  {
    title: '支付时间',
    key: 'date_pay',
    width: 170,
    render: (row) => fmtTime(row.date_pay),
  },
  {
    title: '入库时间',
    key: 'created_at',
    width: 170,
    render: (row) => fmtTime(row.created_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    fixed: 'right',
    render: (row) =>
      h(NButton, { size: 'small', onClick: () => openDetail(row) }, { default: () => '详情' }),
  },
]

const loadList = async () => {
  loading.value = true
  try {
    const res = await getDigisellerOrders({
      page: pagination.value.page as number,
      page_size: pagination.value.pageSize as number,
      keyword: searchKeyword.value.trim(),
      uc_state: searchState.value,
    })
    if (res.code === 200 && res.data) {
      list.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '加载失败')
    }
  } catch (error: any) {
    message.error(error?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const handleSync = async () => {
  syncing.value = true
  try {
    const res = await syncDigisellerOrders()
    if (res.code === 200 && res.data) {
      const failedText = res.data.failed > 0 ? `，失败 ${res.data.failed} 条` : ''
      message.success(`已同步 ${res.data.date_start} 至 ${res.data.date_finish} 的订单，写入 ${res.data.saved} 条${failedText}`)
      await loadList()
    } else {
      message.error(res.message || '同步失败')
    }
  } catch (error: any) {
    message.error(error?.message || '同步失败')
  } finally {
    syncing.value = false
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  loadList()
}

const handleReset = () => {
  searchKeyword.value = ''
  searchState.value = null
  pagination.value.page = 1
  loadList()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadList()
}

const handlePageSizeChange = (size: number) => {
  pagination.value.pageSize = size
  pagination.value.page = 1
  loadList()
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.options-json {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  line-height: 1.5;
}
</style>
