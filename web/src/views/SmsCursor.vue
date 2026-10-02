<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <div class="h-14 bg-white border-b border-gray-200"></div>

    <div class="flex-1 flex flex-col items-center justify-center px-4 py-10">
      <n-card class="cursor-page-card w-full max-w-5xl shadow-md" :bordered="false">
        <div class="cursor-page-header">
          <h1 class="cursor-page-title">{{ pageTitle }}</h1>
          <div class="cursor-tabs">
            <div class="cursor-tabs__track">
              <button
                v-for="tab in tabItems"
                :key="tab.key"
                type="button"
                class="cursor-tab"
                :class="{ 'is-active': activeTab === tab.key }"
                @click="onTabChange(tab.key)"
              >
                <span class="cursor-tab__icon" aria-hidden="true">
                  <svg v-if="tab.key === 'sms'" viewBox="0 0 24 24" fill="none">
                    <path d="M5 6.5h14A1.5 1.5 0 0 1 20.5 8v7A1.5 1.5 0 0 1 19 16.5H9.2L5 19.2V8A1.5 1.5 0 0 1 5 6.5Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/>
                    <path d="M8.5 10.5h7M8.5 13.2h4.5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>
                  </svg>
                  <svg v-else-if="tab.key === 'email'" viewBox="0 0 24 24" fill="none">
                    <rect x="3.6" y="5.8" width="16.8" height="12.4" rx="2" stroke="currentColor" stroke-width="1.7"/>
                    <path d="M4.4 7.4 12 12.6l7.6-5.2" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none">
                    <rect x="5" y="10.2" width="14" height="8.6" rx="2" stroke="currentColor" stroke-width="1.7"/>
                    <path d="M8.2 10.2V8.4a3.8 3.8 0 0 1 7.6 0v1.8" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>
                  </svg>
                </span>
                <span class="cursor-tab__label">{{ tab.label }}</span>
              </button>
            </div>
          </div>
        </div>
        <p class="text-gray-400 text-sm mb-6" v-if="account">Account: {{ account }}</p>

        <div class="cursor-page-body">
        <n-alert v-if="paramError" type="error" :bordered="false">
          {{ paramError }}
        </n-alert>

        <template v-else-if="activeTab === 'sms'">
          <div
            v-if="result?.status === 'received'"
            class="rounded-2xl border border-green-100 bg-green-50 px-8 py-8 text-center"
          >
            <div class="text-sm text-green-700/70 mb-3">Verification code</div>
            <div class="text-5xl font-bold tracking-[0.28em] text-green-700 font-mono pl-[0.28em]">
              {{ result.code || '——' }}
            </div>
            <n-button type="success" class="mt-6" @click="handleCopySms" v-if="result.code">
              {{ copied ? 'Copied' : 'Copy code' }}
            </n-button>
          </div>

          <n-alert
            v-else-if="result?.status === 'error'"
            type="error"
            :bordered="false"
          >
            {{ statusMessage }}
          </n-alert>

          <n-alert
            v-else
            type="info"
            :bordered="false"
          >
            {{ statusMessage }}
          </n-alert>

          <div v-if="result?.expires_at" class="mt-4 rounded-md border border-amber-300 bg-amber-50 px-4 py-3 text-amber-800">
            This number is valid until
            <div class="mt-1 text-base font-bold text-amber-900">{{ result.expires_at }}</div>
          </div>
        </template>

        <template v-else>
          <n-spin :show="mailLoading">
            <p v-if="mailResult?.email" class="text-sm text-gray-500 mb-4">
              Current mailbox: <strong class="text-gray-700">{{ mailResult.email }}</strong>
            </p>
            <n-alert v-if="mailError" type="error" :bordered="false" class="mb-4">
              {{ mailError }}
            </n-alert>
            <n-alert v-else-if="mailResult?.message && !mailResult.inbox.length && !mailResult.junk.length" type="info" :bordered="false" class="mb-4">
              {{ localizeMessage(mailResult.message) }}
            </n-alert>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 min-h-[360px]">
              <div
                v-for="folder in mailFolders"
                :key="folder.key"
                class="rounded-xl border border-gray-100 bg-gray-50/80 p-4 flex flex-col min-h-[320px]"
              >
                <div class="flex items-center justify-between mb-3">
                  <span class="font-semibold text-gray-800">{{ folder.title }}</span>
                  <n-tag size="small" :type="folder.items.length ? 'info' : 'default'">
                    {{ folder.items.length }}
                  </n-tag>
                </div>
                <n-empty v-if="!mailLoading && folder.items.length === 0" description="No emails" class="flex-1" />
                <div v-else class="flex flex-col gap-2 overflow-y-auto max-h-[520px] pr-1">
                  <button
                    v-for="(mail, idx) in folder.items"
                    :key="`${folder.key}-${idx}-${mail.id || mail.received_at}`"
                    type="button"
                    class="mail-item"
                    :class="{ 'has-code': Boolean(mail.code) }"
                    @click="openMailDetail(mail)"
                  >
                    <span class="mail-item__avatar" aria-hidden="true">{{ mailInitial(mail) }}</span>
                    <div class="mail-item__main">
                      <div class="mail-item__top">
                        <span class="mail-item__subject">{{ mail.subject || '(No subject)' }}</span>
                        <span class="mail-item__time">{{ formatMailTime(mail.received_at) }}</span>
                      </div>
                      <div class="mail-item__from">{{ formatMailFrom(mail.from) }}</div>
                      <div v-if="mailPreview(mail)" class="mail-item__preview">{{ mailPreview(mail) }}</div>
                    </div>
                    <n-button
                      v-if="mail.code"
                      type="success"
                      size="tiny"
                      round
                      class="mail-item__code"
                      @click.stop="copyText(mail.code)"
                    >
                      {{ mail.code }}
                    </n-button>
                  </button>
                </div>
              </div>
            </div>
          </n-spin>
        </template>
        </div>

        <div class="cursor-page-footer">
          <n-button
            type="primary"
            :loading="refreshLoading"
            :disabled="refreshDisabled"
            @click="manualRefresh"
          >
            Refresh
          </n-button>
          <span v-if="showCountdown" class="text-gray-400 text-sm">
            {{ countdown > 0 ? `Auto refresh in ${countdown}s` : '' }}
          </span>
        </div>
      </n-card>

      <div class="w-full max-w-5xl text-center text-sm text-gray-400 mt-4">
        <a href="https://plati.market/itm/5957989" target="_blank" rel="noopener noreferrer" class="hover:text-primary-600">Buy</a>
        <span class="mx-2">|</span>
        <a href="https://t.me/aiguoguo199" target="_blank" rel="noopener noreferrer" class="hover:text-primary-600">Contact us</a>
      </div>
    </div>

    <n-modal v-model:show="showDetail" preset="card" :title="detailMail?.subject || 'Message'" style="width: 860px; max-width: 96vw" :bordered="false">
      <template v-if="detailMail">
        <div class="flex flex-col gap-2 mb-4 pb-4 border-b border-gray-100">
          <div class="flex items-center gap-3 text-sm">
            <span class="w-14 shrink-0 text-gray-400">From</span>
            <span>{{ formatMailFrom(detailMail.from) }}</span>
          </div>
          <div class="flex items-center gap-3 text-sm">
            <span class="w-14 shrink-0 text-gray-400">Time</span>
            <span>{{ formatMailTime(detailMail.received_at) }}</span>
          </div>
          <div v-if="detailMail.code" class="flex items-center gap-3 text-sm">
            <span class="w-14 shrink-0 text-gray-400">Code</span>
            <n-button type="success" size="small" round @click="copyText(detailMail.code)">
              {{ detailMail.code }}
            </n-button>
          </div>
        </div>
        <div class="min-h-[200px]">
          <div v-if="detailLoading" class="flex items-center justify-center py-10 text-gray-400">
            <n-spin size="small" />
            <span class="ml-2">Loading message…</span>
          </div>
          <template v-else>
            <iframe
              v-if="detailMail.html_body"
              :srcdoc="prepareMailHtml(detailMail.html_body)"
              sandbox="allow-same-origin"
              class="w-full h-[480px] border-0 rounded-lg bg-white"
            />
            <pre
              v-else-if="detailMail.body"
              class="text-sm leading-relaxed whitespace-pre-wrap break-all m-0 bg-gray-50 rounded-lg p-4 max-h-[480px] overflow-y-auto"
            >{{ detailMail.body }}</pre>
            <n-empty v-else description="No message content" />
          </template>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NButton, NAlert, NSpin, NTag, NEmpty, NModal, useMessage } from 'naive-ui'
import {
  queryCursorSms,
  queryCursorEmail,
  queryCursorEmailDetail,
  queryCursorRecoveryMail,
  queryCursorRecoveryMailDetail,
} from '@/api/sms'
import type { CursorSmsQueryResult, CursorMailQueryResult, CursorMailItem } from '@/api/sms'

type CursorTab = 'sms' | 'email' | 'recovery'

const POLL_SECONDS = 10

const TAB_PATH: Record<CursorTab, string> = {
  sms: '/sms/cursor',
  email: '/email/cursor',
  recovery: '/recovery-mail/cursor',
}

const tabItems: { key: CursorTab; label: string }[] = [
  { key: 'sms', label: 'SMS' },
  { key: 'email', label: 'Email' },
  { key: 'recovery', label: 'Recovery Mail' },
]

const message = useMessage()
const route = useRoute()
const router = useRouter()

const account = ref('')
const rawQuery = ref('')
const queryToken = ref('')
const paramError = ref('')

const loading = ref(false)
const copied = ref(false)
const result = ref<CursorSmsQueryResult | null>(null)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const mailLoading = ref(false)
const mailError = ref('')
const mailResult = ref<CursorMailQueryResult | null>(null)
const showDetail = ref(false)
const detailLoading = ref(false)
const detailMail = ref<CursorMailItem | null>(null)

const activeTab = computed<CursorTab>(() => {
  if (route.path.startsWith('/email/')) return 'email'
  if (route.path.startsWith('/recovery-mail/')) return 'recovery'
  return 'sms'
})

const isMailTab = computed(() => activeTab.value !== 'sms')

const pageTitle = computed(() => {
  if (activeTab.value === 'email') return 'Cursor Email Inbox'
  if (activeTab.value === 'recovery') return 'Cursor Recovery Mail'
  return 'Cursor SMS Lookup'
})

const mailFolders = computed(() => [
  { key: 'inbox', title: 'Inbox', items: mailResult.value?.inbox || [] },
  { key: 'junk', title: 'Junk', items: mailResult.value?.junk || [] },
])

const refreshLoading = computed(() => (activeTab.value === 'sms' ? loading.value : mailLoading.value))
const refreshDisabled = computed(() => Boolean(paramError.value) || result.value?.status === 'received')
const showCountdown = computed(() => activeTab.value === 'sms' && result.value?.status !== 'received' && !paramError.value)

function localizeMessage(raw?: string) {
  const value = String(raw || '').trim()
  if (!value) return ''
  const map: Record<string, string> = {
    查询参数不能为空: 'Query parameter is required.',
    '查询参数格式错误，缺少 ---- 分隔符': 'Invalid query format.',
    账号或密码不能为空: 'Account or password is required.',
    账号不存在或已失效: 'Account not found or inactive.',
    账号或密码错误: 'Incorrect account or password.',
    该账号未配置接码地址: 'This account has no SMS inbox configured.',
    该账号没有可用的原生取件: 'Native mailbox is not available.',
    无法获取邮件详情: 'Unable to load this message.',
    暂无可取件的邮箱: 'No mailbox is available.',
    成功: 'OK',
  }
  if (map[value]) return map[value]
  if (value.startsWith('取码失败:')) return value.replace('取码失败:', 'Failed to fetch code:')
  if (value.startsWith('取件失败:')) return value.replace('取件失败:', 'Failed to fetch mail:')
  if (value.startsWith('拉取正文失败:')) return value.replace('拉取正文失败:', 'Failed to load message:')
  return value
}

function parseParams() {
  const params = new URLSearchParams(window.location.search)
  const token = (params.get('t') || '').trim()
  if (token) {
    queryToken.value = token
    rawQuery.value = ''
    paramError.value = ''
    return
  }

  const raw = window.location.search.replace(/^\?/, '')
  if (!raw) {
    paramError.value = 'This link is missing account information.'
    return
  }

  const sepIndex = raw.indexOf('----')
  if (sepIndex === -1) {
    paramError.value = 'Invalid link format. Please check and try again.'
    return
  }

  rawQuery.value = raw
  queryToken.value = ''

  const accountPart = raw.slice(0, sepIndex)
  try {
    account.value = decodeURIComponent(accountPart).trim()
  } catch {
    account.value = accountPart.trim()
  }

  if (!account.value || raw.slice(sepIndex + 4).trim() === '') {
    paramError.value = 'This link is missing an account or password.'
  }
}

function onTabChange(name: string | number) {
  const tab = String(name) as CursorTab
  if (!TAB_PATH[tab] || tab === activeTab.value) return
  router.push(TAB_PATH[tab] + window.location.search)
}

const statusMessage = computed(() => {
  if (!result.value) return 'Looking up SMS…'
  return localizeMessage(result.value.message) || (result.value.status === 'waiting' ? 'No SMS yet. Please wait.' : 'Lookup failed. Please try again.')
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
      fetchSms()
    }
  }, 1000)
}

async function fetchSms() {
  if (paramError.value || activeTab.value !== 'sms') return
  stopCountdown()
  loading.value = true
  try {
    const res = await queryCursorSms(rawQuery.value, queryToken.value)
    result.value = res.data
    if (res.data.account) {
      account.value = res.data.account
    }
    if (res.data.status === 'received') {
      return
    }
    startCountdown()
  } catch (err: any) {
    result.value = {
      account: account.value,
      status: 'error',
      code: '',
      message: localizeMessage(err?.message) || 'Network error. Please try again.',
      expires_at: '',
    }
    startCountdown()
  } finally {
    loading.value = false
  }
}

async function fetchMails() {
  if (paramError.value || !isMailTab.value) return
  stopCountdown()
  mailLoading.value = true
  mailError.value = ''
  try {
    const res = activeTab.value === 'recovery'
      ? await queryCursorRecoveryMail(rawQuery.value, queryToken.value)
      : await queryCursorEmail(rawQuery.value, queryToken.value)
    mailResult.value = {
      ...res.data,
      inbox: res.data.inbox || [],
      junk: res.data.junk || [],
    }
    if (res.data.account) {
      account.value = res.data.account
    }
  } catch (err: any) {
    mailResult.value = {
      account: account.value,
      email: '',
      source: '',
      inbox: [],
      junk: [],
      message: '',
    }
    mailError.value = localizeMessage(err?.message) || 'Network error. Please try again.'
  } finally {
    mailLoading.value = false
  }
}

function loadCurrentTab() {
  if (paramError.value) return
  if (activeTab.value === 'sms') {
    fetchSms()
    return
  }
  fetchMails()
}

function manualRefresh() {
  if (activeTab.value === 'sms') {
    fetchSms()
    return
  }
  fetchMails()
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    message.success('Copied to clipboard')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    message.error('Copy failed')
  }
}

async function handleCopySms() {
  if (!result.value?.code) return
  await copyText(result.value.code)
}

function canLoadNativeDetail(mail: CursorMailItem) {
  return Boolean(mail.id || (mail.folder && mail.seq_num))
}

function pad2(n: number) {
  return String(n).padStart(2, '0')
}

function formatMailTime(raw?: string) {
  const value = String(raw || '').trim()
  if (!value) return ''
  const date = new Date(value)
  if (!Number.isNaN(date.getTime())) {
    return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`
  }
  return value.replace('T', ' ').replace(/\+\d{2}:\d{2}$/, '').slice(0, 16)
}

function formatMailFrom(raw?: string) {
  const value = String(raw || '').trim()
  return value || 'Unknown sender'
}

function mailInitial(mail: CursorMailItem) {
  const source = (mail.from || mail.subject || '?').trim()
  return source.slice(0, 1).toUpperCase()
}

function isCssLikeText(text: string) {
  const sample = text.slice(0, 400)
  if (/\/\*|\.Root\b|color-scheme|font-family:|@media/.test(sample)) return true
  const braces = (sample.match(/[{}]/g) || []).length
  return braces >= 4 && sample.includes(';')
}

function usableText(text?: string) {
  const value = String(text || '').replace(/\s+/g, ' ').trim()
  if (!value || isCssLikeText(value)) return ''
  return value
}

function mailPreview(mail: CursorMailItem) {
  return usableText(mail.preview || mail.body)
}

function prepareMailHtml(html?: string) {
  const raw = String(html || '').trim()
  if (!raw) return ''
  const match = raw.match(/<(html|body|div|table|p|span|center|h[1-6])\b/i)
  const idx = match?.index ?? -1
  let next = raw
  if (idx > 0) {
    const prefix = raw.slice(0, idx).trim()
    if (prefix && !/<style/i.test(prefix) && prefix.includes('{') && prefix.includes('}')) {
      next = `<style>\n${prefix}\n</style>\n${raw.slice(idx)}`
    }
  }
  if (!/<html[\s>]/i.test(next)) {
    return `<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>${next}</body></html>`
  }
  return next
}

function hasMailHtml(mail: CursorMailItem) {
  return Boolean(String(mail.html_body || '').trim())
}

function hasMailText(mail: CursorMailItem) {
  const text = String(mail.body || '').trim()
  return Boolean(text) && !isCssLikeText(text)
}

async function openMailDetail(mail: CursorMailItem) {
  detailMail.value = { ...mail }
  showDetail.value = true

  if (hasMailHtml(mail) || hasMailText(mail)) {
    return
  }

  const canFetch = (activeTab.value === 'recovery' && Boolean(mail.id))
    || (activeTab.value === 'email' && canLoadNativeDetail(mail))
  if (!canFetch) {
    return
  }

  detailLoading.value = true
  try {
    if (activeTab.value === 'recovery' && mail.id) {
      const res = await queryCursorRecoveryMailDetail(rawQuery.value, mail.id, queryToken.value)
      if (res.data) {
        mail.body = res.data.body || ''
        mail.html_body = res.data.html_body || ''
        detailMail.value = {
          ...mail,
          code: res.data.code || mail.code,
          subject: res.data.subject || mail.subject,
          from: res.data.from || mail.from,
          received_at: res.data.received_at || mail.received_at,
        }
      }
    } else {
      const res = await queryCursorEmailDetail(rawQuery.value, mail, queryToken.value)
      if (res.data) {
        mail.body = res.data.body || ''
        mail.html_body = res.data.html_body || ''
        detailMail.value = { ...mail, code: res.data.code || mail.code }
      }
    }
  } catch (err: any) {
    message.error(localizeMessage(err?.message) || 'Failed to load message')
  } finally {
    detailLoading.value = false
  }
}

watch(
  () => route.path,
  () => {
    showDetail.value = false
    detailMail.value = null
    mailError.value = ''
    loadCurrentTab()
  }
)

onMounted(() => {
  parseParams()
  loadCurrentTab()
})

onBeforeUnmount(() => {
  stopCountdown()
})
</script>

<style scoped>
.cursor-page-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  min-height: 620px;
}

.cursor-page-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.cursor-page-footer {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: auto;
  padding-top: 16px;
}

.mail-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #edf1f5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.mail-item:hover {
  border-color: #bae6fd;
  box-shadow: 0 8px 20px rgba(14, 165, 233, 0.08);
  transform: translateY(-1px);
}

.mail-item.has-code {
  border-color: #bbf7d0;
}

.mail-item__avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(180deg, #e0f2fe 0%, #f0f9ff 100%);
  color: #0369a1;
  font-size: 14px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.mail-item__main {
  min-width: 0;
  flex: 1;
}

.mail-item__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.mail-item__subject {
  min-width: 0;
  color: #111827;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mail-item__time {
  flex-shrink: 0;
  color: #9ca3af;
  font-size: 11px;
}

.mail-item__from,
.mail-item__preview {
  margin-top: 3px;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mail-item__preview {
  color: #9ca3af;
}

.mail-item__code {
  flex-shrink: 0;
  margin-top: 2px;
}

.cursor-page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 4px;
}

.cursor-page-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.3;
  min-width: 0;
}

.cursor-tabs {
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}

.cursor-tabs__track {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  border-radius: 999px;
  background: linear-gradient(180deg, #f3f6fa 0%, #eef2f6 100%);
  border: 1px solid #e5eaf0;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.cursor-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-height: 36px;
  padding: 0 16px;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.01em;
  line-height: 1;
  cursor: pointer;
  transition: color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.cursor-tab:hover {
  color: #374151;
  background: rgba(255, 255, 255, 0.65);
}

.cursor-tab.is-active {
  color: #0369a1;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06), 0 6px 16px rgba(14, 165, 233, 0.12);
}

.cursor-tab__icon {
  display: inline-flex;
  width: 16px;
  height: 16px;
}

.cursor-tab__icon svg {
  width: 16px;
  height: 16px;
}

.cursor-tab__label {
  white-space: nowrap;
}

@media (max-width: 640px) {
  .cursor-page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .cursor-tabs {
    width: 100%;
    justify-content: flex-start;
  }

  .cursor-tabs__track {
    width: 100%;
  }

  .cursor-tab {
    flex: 1;
    padding: 0 8px;
    font-size: 12px;
    gap: 5px;
  }
}
</style>
