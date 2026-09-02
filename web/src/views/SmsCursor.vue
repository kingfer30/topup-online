<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <!-- 顶部预留区域：后续用于展示与 web 主站关联的菜单，暂不实现 -->
    <div class="h-14 bg-white border-b border-gray-200"></div>

    <div class="flex-1 flex items-center justify-center px-4 py-10">
      <n-card class="w-full max-w-lg shadow-md" :bordered="false">
        <h1 class="text-2xl font-bold text-gray-800 mb-1">Cursor 短信验证码查询</h1>
        <p class="text-gray-400 text-sm mb-6" v-if="account">账号：{{ account }}</p>

        <!-- 参数缺失 -->
        <n-alert v-if="paramError" type="error" :bordered="false">
          {{ paramError }}
        </n-alert>

        <template v-else>
          <!-- 验证码展示 -->
          <div
            v-if="result?.status === 'received'"
            class="rounded-lg border-l-4 border-green-500 bg-green-50 p-5 text-center"
          >
            <div class="text-sm text-gray-500 mb-2">验证码</div>
            <div class="text-4xl font-bold tracking-widest text-green-700 font-mono">
              {{ result.code || '——' }}
            </div>
            <n-button size="small" class="mt-4" @click="handleCopy" v-if="result.code">
              {{ copied ? '已复制' : '复制验证码' }}
            </n-button>
          </div>

          <!-- 等待 / 错误状态 -->
          <div
            v-else
            class="rounded-lg border-l-4 p-4 leading-relaxed whitespace-pre-wrap"
            :class="statusBoxClass"
          >
            {{ statusMessage }}
          </div>

          <!-- 有效期 -->
          <div v-if="result?.expires_at" class="mt-4 rounded-md border border-amber-300 bg-amber-50 px-4 py-3 text-amber-800">
            当前号码有效期至
            <div class="mt-1 text-base font-bold text-amber-900">{{ result.expires_at }}</div>
          </div>

          <!-- 操作区 -->
          <div class="flex items-center gap-3 mt-5">
            <n-button type="primary" :loading="loading" :disabled="result?.status === 'received'" @click="manualRefresh">
              立即刷新
            </n-button>
            <span v-if="result?.status !== 'received'" class="text-gray-400 text-sm">
              {{ countdown > 0 ? `${countdown} 秒后自动刷新` : '' }}
            </span>
          </div>
        </template>
      </n-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { NCard, NButton, NAlert, useMessage } from 'naive-ui'
import { queryCursorSms } from '@/api/sms'
import type { CursorSmsQueryResult } from '@/api/sms'

const POLL_SECONDS = 10

const message = useMessage()

const account = ref('')
const pass = ref('')
const paramError = ref('')

const loading = ref(false)
const copied = ref(false)
const result = ref<CursorSmsQueryResult | null>(null)

const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

// 从地址栏解析 account----pass，例如 /sms/cursor?account----pass
function parseParams() {
  const search = window.location.search.replace(/^\?/, '')
  if (!search) {
    paramError.value = '查询链接缺少账号信息，请检查链接后重试。'
    return
  }

  let raw = search
  try {
    raw = decodeURIComponent(search)
  } catch {
    raw = search
  }

  const sepIndex = raw.indexOf('----')
  if (sepIndex === -1) {
    paramError.value = '查询链接格式错误，请检查链接后重试。'
    return
  }

  account.value = raw.slice(0, sepIndex).trim()
  pass.value = raw.slice(sepIndex + 4).trim()

  if (!account.value || !pass.value) {
    paramError.value = '查询链接缺少账号或密码，请检查链接后重试。'
  }
}

const statusMessage = computed(() => {
  if (!result.value) return '正在查询短信，请稍候……'
  return result.value.message || (result.value.status === 'waiting' ? '暂未收到短信，请等待。' : '查询失败，请稍后重试。')
})

const statusBoxClass = computed(() => {
  if (!result.value || result.value.status === 'waiting') {
    return 'border-blue-500 bg-blue-50 text-blue-800'
  }
  return 'border-red-500 bg-red-50 text-red-700'
})

function stopCountdown() {
  if (countdownTimer !== null) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  countdown.value = 0
}

function startCountdown() {
  stopCountdown()
  countdown.value = POLL_SECONDS
  countdownTimer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      stopCountdown()
      fetchCode()
    }
  }, 1000)
}

async function fetchCode() {
  if (paramError.value) return
  stopCountdown()
  loading.value = true
  try {
    const res = await queryCursorSms(account.value, pass.value)
    result.value = res.data
    if (res.data.status === 'received') {
      return
    }
    startCountdown()
  } catch (err: any) {
    result.value = {
      status: 'error',
      code: '',
      message: err?.message || '网络异常，请稍后重试。',
      expires_at: '',
    }
    startCountdown()
  } finally {
    loading.value = false
  }
}

function manualRefresh() {
  fetchCode()
}

async function handleCopy() {
  if (!result.value?.code) return
  try {
    await navigator.clipboard.writeText(result.value.code)
    copied.value = true
    message.success('已复制到剪贴板')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    message.error('复制失败，请手动选择')
  }
}

onMounted(() => {
  parseParams()
  if (!paramError.value) {
    fetchCode()
  }
})

onBeforeUnmount(() => {
  stopCountdown()
})
</script>
