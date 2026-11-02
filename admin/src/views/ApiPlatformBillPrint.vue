<template>
  <div class="bill-page">
    <div class="toolbar">
      <n-button type="primary" :loading="downloading" :disabled="!bill" @click="downloadPdf">下载 PDF</n-button>
      <n-button :loading="shooting" :disabled="!bill" @click="copyShot">生成截图</n-button>
    </div>
    <div ref="sheetRef" class="sheet">
      <header v-if="bill">
        <h1>结算快照 {{ bill.bill_no }}</h1>
        <p>{{ platform.base_url }}<span v-if="filterNames">（{{ filterNames }}）</span></p>
        <p>{{ snapshotTime }} · {{ statusLabel }}</p>
        <p class="note">消耗总额、已结算、待支付额度为美元。成本价、销售价和金额为人民币。金额 = 销售价 × 待支付额度。</p>
      </header>
      <table v-if="bill">
        <thead>
          <tr>
            <th>ID</th>
            <th>渠道</th>
            <th>消耗总额</th>
            <th>已结算</th>
            <th>待支付额度</th>
            <th>成本价</th>
            <th>销售价</th>
            <th>金额</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in bill.items || []" :key="item.id">
            <td>{{ item.channel_id }}</td>
            <td>{{ item.name }}</td>
            <td>{{ usd(item.used_amount) }}</td>
            <td>{{ usd(item.settled_amount) }}</td>
            <td>{{ usd(item.pay_quota) }}</td>
            <td>{{ cny(item.cost_price) }}</td>
            <td>{{ cny(item.sale_price) }}</td>
            <td>{{ cny(item.amount_cny) }}</td>
          </tr>
        </tbody>
      </table>
      <footer v-if="bill">
        <p>合计：{{ bill.channel_count || bill.items?.length || 0 }} 个渠道 · 待支付 {{ usd(bill.pay_quota) }} · 金额 {{ cny(bill.amount_cny) }}</p>
        <p>已结算合计 {{ usd(settledSum) }}</p>
      </footer>
      <p v-if="loadFailed" class="note">账单加载失败，请重新从后台打开。</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { NButton, useMessage } from 'naive-ui'
import html2canvas from 'html2canvas'
import { jsPDF } from 'jspdf'
import { getApiPlatformBill, type ApiPlatformBill } from '@/api/api-platform'

const route = useRoute()
const message = useMessage()
const sheetRef = ref<HTMLElement | null>(null)
const bill = ref<ApiPlatformBill | null>(null)
const loadFailed = ref(false)
const downloading = ref(false)
const shooting = ref(false)
const statusLabel = ref('')
const platform = ref({ base_url: '', account: '', remark: '' })

const usd = (value?: number | null) =>
  `$${Number(value || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
const cny = (value?: number | null) =>
  `¥${Number(value || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`

const filterNames = computed(() => {
  const names = new Set((bill.value?.items || []).map((item) => item.filter_name).filter(Boolean))
  return [...names].join(' / ')
})
const snapshotTime = computed(() => {
  if (!bill.value?.created_at) return ''
  return new Date(bill.value.created_at).toLocaleString('zh-CN', { hour12: false, timeZone: 'Asia/Shanghai' }) + ' 中国时间'
})
const settledSum = computed(() => (bill.value?.items || []).reduce((sum, item) => sum + Number(item.settled_amount || 0), 0))

const captureSheet = () => {
  if (!sheetRef.value) return Promise.reject(new Error('页面未准备好'))
  return html2canvas(sheetRef.value, {
    backgroundColor: '#ffffff',
    scale: 2,
    useCORS: true,
  })
}

const downloadPdf = async () => {
  if (!bill.value) return
  downloading.value = true
  try {
    const canvas = await captureSheet()
    const image = canvas.toDataURL('image/png')
    const pdf = new jsPDF('p', 'mm', 'a4')
    const pageWidth = pdf.internal.pageSize.getWidth()
    const pageHeight = pdf.internal.pageSize.getHeight()
    const imageHeight = (canvas.height * pageWidth) / canvas.width
    let remaining = imageHeight
    let offset = 0
    pdf.addImage(image, 'PNG', 0, offset, pageWidth, imageHeight)
    remaining -= pageHeight
    while (remaining > 0) {
      offset = remaining - imageHeight
      pdf.addPage()
      pdf.addImage(image, 'PNG', 0, offset, pageWidth, imageHeight)
      remaining -= pageHeight
    }
    pdf.save(`${bill.value.bill_no}.pdf`)
  } catch {
    message.error('下载 PDF 失败')
  } finally {
    downloading.value = false
  }
}

const copyShot = async () => {
  if (!sheetRef.value) return
  shooting.value = true
  const image = captureSheet().then(
    (canvas) =>
      new Promise<Blob>((resolve, reject) => {
        canvas.toBlob((blob) => (blob ? resolve(blob) : reject(new Error('empty'))), 'image/png')
      })
  )
  try {
    await navigator.clipboard.write([new ClipboardItem({ 'image/png': image })])
    message.success('结算快照已复制到剪贴板')
  } catch {
    message.error('复制到剪贴板失败，请允许浏览器访问剪贴板')
  } finally {
    shooting.value = false
  }
}

onMounted(async () => {
  const platformId = Number(route.params.platformId)
  const billId = Number(route.params.billId)
  try {
    const res = await getApiPlatformBill(platformId, billId)
    if (res.code === 200 && res.data) {
      bill.value = res.data.bill
      statusLabel.value = res.data.status_label
      platform.value = res.data.platform
      return
    }
  } catch {
    // 下面统一显示失败
  }
  loadFailed.value = true
})
</script>

<style scoped>
.bill-page {
  max-width: 1100px;
  margin: 24px auto;
  padding: 0 16px 32px;
  color: #1f2933;
  font-family: "Microsoft YaHei", sans-serif;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.sheet {
  background: #fff;
  padding: 8px 4px 16px;
}
header h1 {
  margin: 0 0 8px;
  font-size: 22px;
}
header p,
footer p {
  margin: 4px 0;
}
.note {
  color: #52606d;
  font-size: 13px;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
  font-size: 13px;
}
th,
td {
  border-bottom: 1px solid #d9e2ec;
  padding: 8px 6px;
  text-align: left;
}
th {
  background: #f0f4f8;
}
</style>
