<template>
  <div>
    <n-space vertical :size="24">
      <!-- 页面标题 -->
      <div>
        <h1 class="text-2xl font-bold mb-2">控制台</h1>
        <p class="text-gray-500">欢迎使用后台管理系统</p>
      </div>

      <!-- 统计卡片 -->
      <n-grid :x-gap="16" :y-gap="16" :cols="4" responsive="screen">
        <n-grid-item>
          <n-card :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
            <n-statistic label="今日订单" value="125">
              <template #prefix>
                <n-icon size="24" color="#18a058">
                  <span class="text-2xl">📦</span>
                </n-icon>
              </template>
            </n-statistic>
            <template #footer>
              <n-text depth="3">
                较昨日 
                <n-text type="success">+12%</n-text>
              </n-text>
            </template>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
            <n-statistic label="今日收入" value="¥12,580">
              <template #prefix>
                <n-icon size="24" color="#2080f0">
                  <span class="text-2xl">💰</span>
                </n-icon>
              </template>
            </n-statistic>
            <template #footer>
              <n-text depth="3">
                较昨日 
                <n-text type="success">+8%</n-text>
              </n-text>
            </template>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
            <n-statistic label="总用户数" value="3,842">
              <template #prefix>
                <n-icon size="24" color="#f0a020">
                  <span class="text-2xl">👥</span>
                </n-icon>
              </template>
            </n-statistic>
            <template #footer>
              <n-text depth="3">
                新增用户 
                <n-text type="info">+25</n-text>
              </n-text>
            </template>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
            <n-statistic label="可用卡密" value="568">
              <template #prefix>
                <n-icon size="24" color="#d03050">
                  <span class="text-2xl">🎫</span>
                </n-icon>
              </template>
            </n-statistic>
            <template #footer>
              <n-text depth="3">
                库存充足
              </n-text>
            </template>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 图表区域 -->
      <n-grid :x-gap="16" :y-gap="16" :cols="2" responsive="screen">
        <n-grid-item>
          <n-card title="订单趋势" :bordered="false" class="shadow-sm">
            <div class="h-64 flex items-center justify-center text-gray-400">
              图表区域（可集成 ECharts 或其他图表库）
            </div>
          </n-card>
        </n-grid-item>

        <n-grid-item>
          <n-card title="收入统计" :bordered="false" class="shadow-sm">
            <div class="h-64 flex items-center justify-center text-gray-400">
              图表区域（可集成 ECharts 或其他图表库）
            </div>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 最近订单 -->
      <n-card title="最近订单" :bordered="false" class="shadow-sm">
        <n-data-table
          :columns="columns"
          :data="data"
          :pagination="pagination"
          :bordered="false"
        />
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { h } from 'vue'
import {
  NSpace,
  NCard,
  NStatistic,
  NIcon,
  NText,
  NGrid,
  NGridItem,
  NDataTable,
  NTag,
  type DataTableColumns,
} from 'naive-ui'

// 表格列定义
const columns: DataTableColumns = [
  {
    title: '订单号',
    key: 'orderNo',
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
      const statusMap: Record<string, { type: any; text: string }> = {
        success: { type: 'success', text: '成功' },
        pending: { type: 'warning', text: '处理中' },
        failed: { type: 'error', text: '失败' },
      }
      const status = statusMap[row.status]
      return h(NTag, { type: status.type }, { default: () => status.text })
    },
  },
  {
    title: '时间',
    key: 'time',
  },
]

// 表格数据
const data = [
  {
    orderNo: 'ORD20250001',
    user: 'user001@example.com',
    amount: 99,
    status: 'success',
    time: '2025-11-29 10:30:25',
  },
  {
    orderNo: 'ORD20250002',
    user: 'user002@example.com',
    amount: 199,
    status: 'pending',
    time: '2025-11-29 10:28:15',
  },
  {
    orderNo: 'ORD20250003',
    user: 'user003@example.com',
    amount: 299,
    status: 'success',
    time: '2025-11-29 10:25:10',
  },
  {
    orderNo: 'ORD20250004',
    user: 'user004@example.com',
    amount: 99,
    status: 'failed',
    time: '2025-11-29 10:20:05',
  },
  {
    orderNo: 'ORD20250005',
    user: 'user005@example.com',
    amount: 399,
    status: 'success',
    time: '2025-11-29 10:15:30',
  },
]

// 分页配置
const pagination = {
  pageSize: 5,
}
</script>

