<template>
  <div>
    <n-space vertical :size="24">
      <!-- 页面标题 -->
      <div>
        <h1 class="apple-page-title">控制台</h1>
        <p class="apple-page-subtitle">欢迎使用后台管理系统</p>
      </div>

      <!-- 今日总览统计 -->
      <n-grid :x-gap="16" :y-gap="16" :cols="4" responsive="screen">
        <n-grid-item>
          <n-card :bordered="false" class="stat-card">
            <div class="stat-card-inner">
              <div class="stat-icon stat-icon-green">
                <span class="text-2xl">📦</span>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ loading ? '—' : totalSoldCount }}</div>
                <div class="stat-label">{{ dateLabel }} 总售出</div>
              </div>
            </div>
            <div class="stat-footer">
              所有卡密类型合计
            </div>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="stat-card">
            <div class="stat-card-inner">
              <div class="stat-icon stat-icon-blue">
                <span class="text-2xl">💵</span>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ loading ? '—' : `$${totalRevenueUSD.toFixed(2)}` }}</div>
                <div class="stat-label">{{ dateLabel }} 收入（USD）</div>
              </div>
            </div>
            <div class="stat-footer">
              售出价格按美元计算
            </div>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="stat-card">
            <div class="stat-card-inner">
              <div class="stat-icon stat-icon-orange">
                <span class="text-2xl">💰</span>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ loading ? '—' : `¥${totalRevenueCNY.toFixed(2)}` }}</div>
                <div class="stat-label">{{ dateLabel }} 收入（CNY）</div>
              </div>
            </div>
            <div class="stat-footer">
              按汇率 1 USD = 7 CNY 换算
            </div>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="stat-card">
            <div class="stat-card-inner">
              <div class="stat-icon stat-icon-purple">
                <span class="text-2xl">🎫</span>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ loading ? '—' : totalStockCount }}</div>
                <div class="stat-label">剩余在售库存</div>
              </div>
            </div>
            <div class="stat-footer">
              所有类型未售出合计
            </div>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 按卡密类型分类统计 -->
      <n-card :bordered="false" class="table-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-header-title">各类型卡密销售统计</span>
            <n-space align="center" :size="12">
              <n-date-picker
                v-model:value="selectedDate"
                type="date"
                clearable
                :placeholder="'选择日期（默认今天）'"
                style="width: 180px"
                @update:value="onDateChange"
              />
              <n-button size="small" :loading="loading" @click="fetchStats">
                <template #icon><span>🔄</span></template>
                刷新
              </n-button>
            </n-space>
          </div>
        </template>

        <n-spin :show="loading">
          <div v-if="!loading && statList.length === 0" class="empty-tip">
            暂无数据（今日尚无售出记录）
          </div>
          <n-grid v-else :x-gap="16" :y-gap="16" :cols="4" responsive="screen">
            <n-grid-item v-for="item in statList" :key="item.category">
              <n-card :bordered="false" class="type-stat-card">
                <div class="type-stat-header">
                  <span class="type-stat-badge">{{ item.category.toUpperCase() }}</span>
                </div>
                <div class="type-stat-body">
                  <div class="type-stat-row">
                    <span class="type-stat-row-label">今日售出</span>
                    <span class="type-stat-row-value sold">{{ item.sold_count }} 件</span>
                  </div>
                  <div class="type-stat-row">
                    <span class="type-stat-row-label">收入（USD）</span>
                    <span class="type-stat-row-value usd">${{ item.revenue_usd.toFixed(2) }}</span>
                  </div>
                  <div class="type-stat-row">
                    <span class="type-stat-row-label">收入（CNY）</span>
                    <span class="type-stat-row-value cny">¥{{ item.revenue_cny.toFixed(2) }}</span>
                  </div>
                  <!-- 库存按订阅类型细分 -->
                  <div class="type-stat-divider"></div>
                  <div class="type-stat-stock-header">
                    <span class="type-stat-row-label">库存明细</span>
                    <span class="type-stat-row-value" :class="item.stock_count > 0 ? 'stock-ok' : 'stock-empty'">
                      合计 {{ item.stock_count }} 件
                    </span>
                  </div>
                  <div
                    v-if="item.stock_by_type && item.stock_by_type.length > 0"
                    class="type-stat-stock-list"
                  >
                    <div
                      v-for="sub in item.stock_by_type"
                      :key="sub.subscription_type"
                      class="type-stat-stock-row"
                    >
                      <span class="stock-sub-type">{{ sub.subscription_type || '(未设置)' }}</span>
                      <span class="stock-sub-count" :class="sub.count > 0 ? 'stock-ok' : 'stock-empty'">
                        {{ sub.count }} 件
                      </span>
                    </div>
                  </div>
                  <div v-else class="type-stat-stock-empty">暂无库存</div>
                </div>
              </n-card>
            </n-grid-item>
          </n-grid>
        </n-spin>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  NSpace,
  NCard,
  NGrid,
  NGridItem,
  NButton,
  NSpin,
  NDatePicker,
} from 'naive-ui'
import { getDashboardStats, type CardTypeStat } from '@/api/card'

// 统计数据
const loading = ref(false)
const statList = ref<CardTypeStat[]>([])

// 日期选择（null 表示今天，值为毫秒时间戳）
const selectedDate = ref<number | null>(null)

// 将时间戳转为 YYYY-MM-DD 字符串（Asia/Shanghai）
const toDateStr = (ts: number | null): string | undefined => {
  if (!ts) return undefined
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  // 使用本地时间（浏览器通常与用户时区一致）
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 当前选中日期的显示文本
const dateLabel = computed(() => {
  if (!selectedDate.value) return '今天'
  return toDateStr(selectedDate.value) ?? '今天'
})

// 合计值
const totalSoldCount = computed(() =>
  statList.value.reduce((sum, item) => sum + item.sold_count, 0)
)
const totalRevenueUSD = computed(() =>
  statList.value.reduce((sum, item) => sum + item.revenue_usd, 0)
)
const totalRevenueCNY = computed(() =>
  statList.value.reduce((sum, item) => sum + item.revenue_cny, 0)
)
const totalStockCount = computed(() =>
  statList.value.reduce((sum, item) => sum + item.stock_count, 0)
)

// 获取统计数据
const fetchStats = async () => {
  loading.value = true
  try {
    const dateStr = toDateStr(selectedDate.value)
    const res = await getDashboardStats(dateStr)
    if (res.code === 200 && res.data) {
      statList.value = res.data
    }
  } catch (e) {
    console.error('获取统计数据失败', e)
  } finally {
    loading.value = false
  }
}

// 日期变化时自动刷新
const onDateChange = () => {
  fetchStats()
}

onMounted(() => {
  fetchStats()
})
</script>

<style scoped>
/* Apple 风格统计卡片 */
.stat-card {
  border-radius: 16px !important;
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08), 0 4px 8px rgba(0, 0, 0, 0.04) !important;
}

.stat-card-inner {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon-green {
  background: rgba(52, 199, 89, 0.1);
}

.stat-icon-blue {
  background: rgba(0, 122, 255, 0.1);
}

.stat-icon-orange {
  background: rgba(255, 149, 0, 0.1);
}

.stat-icon-purple {
  background: rgba(175, 82, 222, 0.1);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #1d1d1f;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #86868b;
  font-weight: 500;
  margin-top: 2px;
}

.stat-footer {
  font-size: 13px;
  color: #86868b;
  padding-top: 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.04);
}

/* 卡片标题行 */
.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.card-header-title {
  font-size: 17px;
  font-weight: 600;
  color: #1d1d1f;
  white-space: nowrap;
}

/* 表格卡片 */
.table-card {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

:deep(.table-card .n-card-header__main) {
  font-size: 17px;
  font-weight: 600;
  color: #1d1d1f;
}

/* 类型统计卡片 */
.type-stat-card {
  border-radius: 14px !important;
  border: 1px solid rgba(0, 0, 0, 0.06) !important;
  transition: all 0.3s ease !important;
}

.type-stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08) !important;
}

.type-stat-header {
  margin-bottom: 14px;
}

.type-stat-badge {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(0, 122, 255, 0.08);
  color: #007aff;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.type-stat-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.type-stat-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.type-stat-row-label {
  font-size: 13px;
  color: #86868b;
}

.type-stat-row-value {
  font-size: 15px;
  font-weight: 600;
}

.type-stat-row-value.sold {
  color: #1d1d1f;
}

.type-stat-row-value.stock-ok {
  color: #34c759;
}

.type-stat-row-value.stock-empty {
  color: #ff3b30;
}

.type-stat-row-value.usd {
  color: #007aff;
}

.type-stat-row-value.cny {
  color: #34c759;
}

.empty-tip {
  text-align: center;
  padding: 40px 0;
  color: #86868b;
  font-size: 14px;
}

/* 库存明细 */
.type-stat-divider {
  height: 1px;
  background: rgba(0, 0, 0, 0.05);
  margin: 4px 0;
}

.type-stat-stock-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.type-stat-stock-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-left: 4px;
}

.type-stat-stock-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stock-sub-type {
  font-size: 12px;
  color: #86868b;
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 8px;
  border-radius: 10px;
}

.stock-sub-count {
  font-size: 13px;
  font-weight: 600;
}

.type-stat-stock-empty {
  font-size: 12px;
  color: #c0c0c0;
  text-align: center;
  padding: 4px 0;
}
</style>
