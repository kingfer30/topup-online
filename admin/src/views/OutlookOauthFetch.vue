<template>
  <div class="outlook-fetch-page">
    <div class="page-head">
      <h1 class="apple-page-title">Oauth 取件</h1>
      <p class="apple-page-subtitle">输入 Outlook 账号信息，通过微软授权拉取收件箱与垃圾箱邮件</p>
    </div>

    <n-card :bordered="false" class="shadow-sm input-card">
      <template #header>
        <span class="card-title">账号信息</span>
      </template>

      <div class="form-area">
        <div class="format-row">
          <span class="format-label">账号格式</span>
          <n-select
            v-model:value="accountFormat"
            :options="formatOptions"
            style="flex: 1"
            :disabled="loading"
            @update:value="onFormatChange"
          />
        </div>
        <p class="hint-text">{{ formatHint }}</p>

        <n-input
          v-model:value="accountLine"
          type="textarea"
          :placeholder="formatPlaceholder"
          :autosize="{ minRows: 3, maxRows: 6 }"
          :disabled="loading"
        />

        <div class="action-row">
          <n-text v-if="fetchData" depth="3" style="font-size: 13px">
            当前邮箱：<strong>{{ fetchData.email }}</strong>
          </n-text>
          <div v-else />
          <n-button type="primary" size="large" round :loading="loading" @click="handleFetch">
            取件
          </n-button>
        </div>
      </div>
    </n-card>

    <div v-if="fetchData" class="mail-panels">
      <n-card
        v-for="folder in folders"
        :key="folder.title"
        :bordered="false"
        class="shadow-sm mail-card"
      >
        <template #header>
          <div class="folder-header">
            <span class="folder-title">{{ folder.title }}</span>
            <n-tag size="small" :type="folder.items.length ? 'info' : 'default'">
              {{ folder.items.length }} 封
            </n-tag>
          </div>
        </template>

        <n-empty v-if="folder.items.length === 0" description="暂无邮件" />
        <n-list v-else hoverable clickable>
          <n-list-item
            v-for="(mail, idx) in folder.items"
            :key="`${folder.title}-${idx}-${mail.received_at}`"
            @click="openDetail(mail)"
          >
            <div class="mail-row">
              <div class="mail-meta">
                <span class="mail-subject">{{ mail.subject }}</span>
                <span class="mail-from">{{ mail.from }}</span>
                <span class="mail-time">{{ mail.received_at }}</span>
              </div>
              <div class="mail-actions" @click.stop>
                <n-button
                  v-if="mail.code"
                  type="success"
                  size="small"
                  round
                  @click="copyText(mail.code)"
                >
                  {{ mail.code }}
                </n-button>
                <n-button
                  v-else
                  type="warning"
                  size="small"
                  round
                  @click="openDetail(mail)"
                >
                  详情
                </n-button>
              </div>
            </div>
          </n-list-item>
        </n-list>
      </n-card>
    </div>

    <n-modal
      v-model:show="showDetail"
      preset="card"
      :title="detailMail?.subject || '邮件详情'"
      style="width: 860px; max-width: 96vw"
      :bordered="false"
    >
      <template v-if="detailMail">
        <div class="detail-meta">
          <div class="detail-meta-row">
            <span class="detail-label">发件人</span>
            <span class="detail-value">{{ detailMail.from }}</span>
          </div>
          <div class="detail-meta-row">
            <span class="detail-label">时间</span>
            <span class="detail-value">{{ detailMail.received_at }}</span>
          </div>
          <div v-if="detailMail.code" class="detail-meta-row">
            <span class="detail-label">验证码</span>
            <n-button type="success" size="small" round @click="copyText(detailMail.code)">
              {{ detailMail.code }}
            </n-button>
          </div>
        </div>

        <div class="detail-body-wrap">
          <div v-if="detailLoading" style="display:flex;align-items:center;justify-content:center;padding:40px 0;">
            <n-spin size="small" />
            <span style="margin-left:8px;color:#999">正在加载正文…</span>
          </div>
          <template v-else>
            <iframe
              v-if="detailMail.html_body"
              :srcdoc="detailMail.html_body"
              sandbox="allow-same-origin"
              class="detail-iframe"
            />
            <pre v-else-if="detailMail.body" class="detail-body">{{ detailMail.body }}</pre>
            <n-empty v-else description="暂无正文内容" />
          </template>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  NButton, NCard, NEmpty, NInput, NList, NListItem,
  NModal, NSelect, NSpin, NTag, NText, useMessage,
} from 'naive-ui'
import {
  fetchOutlookMails,
  fetchOutlookDetail,
  type OutlookMailItem,
  type OutlookFetchData,
} from '@/api/outlook-oauth'

const message = useMessage()
const STORAGE_KEY = 'outlookOauthAccountFormat'

interface FormatConfig {
  value: string
  label: string
  placeholder: string
  hint: string
}

const FORMAT_CONFIGS: FormatConfig[] = [
  {
    value: '1',
    label: '格式1: 邮箱|邮箱密码|refresh_token|client_id',
    placeholder: 'user@outlook.com|邮箱密码|refresh_token|client_id',
    hint: '竖线分隔，字段顺序同格式2',
  },
  {
    value: '2',
    label: '格式2: 邮箱----邮箱密码----refresh_token----client_id',
    placeholder: 'user@outlook.com----邮箱密码----refresh_token----client_id',
    hint: 'refresh_token 与 client_id 顺序与格式3 相反',
  },
  {
    value: '3',
    label: '格式3: 邮箱----邮箱密码----client_id----refresh_token',
    placeholder: 'user@outlook.com----邮箱密码----client_id----refresh_token',
    hint: '邮箱密码用于 IMAP，client_id / refresh_token 为 Outlook OAuth 凭据',
  },
  {
    value: '4',
    label: '格式4: 邮箱----GPT密码----邮箱密码----client_id----refresh_token',
    placeholder: 'user@outlook.com----GPT密码----邮箱密码----client_id----refresh_token',
    hint: 'GPT密码字段会被忽略，取件使用邮箱密码',
  },
  {
    value: '5',
    label: '格式5: 邮箱----邮箱密码----GPT密码----client_id----refresh_token',
    placeholder: 'user@outlook.com----邮箱密码----GPT密码----client_id----refresh_token',
    hint: 'GPT密码字段会被忽略，取件使用邮箱密码',
  },
]

const formatOptions = FORMAT_CONFIGS.map(c => ({ label: c.label, value: c.value }))

function migrateAccountFormat(saved: string | null) {
  if (!saved) return '1'
  const legacy: Record<string, string> = { f5: '1', f4: '2', f1: '3', f2: '4', f3: '5' }
  return legacy[saved] || saved
}

const accountFormat = ref('1')
const accountLine = ref('')
const loading = ref(false)
const fetchData = ref<OutlookFetchData | null>(null)
const showDetail = ref(false)
const detailMail = ref<OutlookMailItem | null>(null)
const detailLoading = ref(false)

const currentFormatConfig = computed(() =>
  FORMAT_CONFIGS.find(c => c.value === accountFormat.value) || FORMAT_CONFIGS[0],
)

const formatPlaceholder = computed(() => currentFormatConfig.value.placeholder)
const formatHint = computed(() => currentFormatConfig.value.hint)

function onFormatChange(value: string) {
  localStorage.setItem(STORAGE_KEY, value)
}

onMounted(() => {
  accountFormat.value = migrateAccountFormat(localStorage.getItem(STORAGE_KEY))
})

const folders = computed(() => {
  if (!fetchData.value) return []
  return [
    { title: '📥 收件箱', items: fetchData.value.inbox },
    { title: '🗑️ 垃圾箱', items: fetchData.value.junk },
  ]
})

async function handleFetch() {
  const line = accountLine.value.trim()
  if (!line) {
    message.warning('请输入账号信息')
    return
  }
  loading.value = true
  fetchData.value = null

  try {
    const res = await fetchOutlookMails(line, accountFormat.value)
    if (res.code !== 200) {
      message.error(res.message || '取件失败')
      return
    }
    fetchData.value = res.data
  } catch (e: any) {
    message.error(e?.message || '请求失败')
  } finally {
    loading.value = false
  }
}

async function openDetail(mail: OutlookMailItem) {
  detailMail.value = mail
  showDetail.value = true

  // 正文未加载时按需拉取
  if (!mail.body && !mail.html_body) {
    detailLoading.value = true
    try {
      const res = await fetchOutlookDetail(
        accountLine.value.trim(),
        mail.folder,
        mail.seq_num,
        accountFormat.value,
        mail.id || '',
      )
      if (res.code === 200 && res.data) {
        mail.body = res.data.body
        mail.html_body = res.data.html_body
        // 不从详情正文中回写 code，避免误匹配正文里的无关数字
        detailMail.value = { ...mail }
      }
    } catch {
      // 静默失败，弹窗仍展示已有信息
    } finally {
      detailLoading.value = false
    }
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    message.success('已复制：' + text)
  } catch {
    message.error('复制失败')
  }
}
</script>

<style scoped>
.outlook-fetch-page {
  padding: 24px;
  max-width: 100%;
}

.page-head {
  margin-bottom: 24px;
}

.apple-page-title {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.5px;
  margin: 0 0 6px;
  color: var(--n-text-color);
}

.apple-page-subtitle {
  font-size: 14px;
  color: var(--n-text-color-3);
  margin: 0;
}

.input-card {
  border-radius: 16px;
  margin-bottom: 24px;
}

.card-title {
  font-weight: 600;
  font-size: 15px;
}

.form-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hint-text {
  margin: 0;
  font-size: 13px;
  color: var(--n-text-color-3);
}

.format-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.format-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-2);
  flex-shrink: 0;
}

.action-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}

.mail-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

@media (max-width: 900px) {
  .mail-panels {
    grid-template-columns: 1fr;
  }
}

.mail-card {
  border-radius: 16px;
}

.folder-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.folder-title {
  font-weight: 600;
  font-size: 15px;
}

.mail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 2px 0;
}

.mail-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.mail-subject {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--n-text-color);
}

.mail-from {
  font-size: 12px;
  color: var(--n-text-color-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mail-time {
  font-size: 11px;
  color: var(--n-text-color-4, #aaa);
}

.mail-actions {
  flex-shrink: 0;
}

.detail-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--n-divider-color);
}

.detail-meta-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-3);
  width: 56px;
  flex-shrink: 0;
}

.detail-value {
  font-size: 13px;
  color: var(--n-text-color);
}

.detail-body-wrap {
  min-height: 200px;
}

.detail-iframe {
  width: 100%;
  height: 480px;
  border: none;
  border-radius: 8px;
  background: #fff;
}

.detail-body {
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--n-text-color);
  margin: 0;
  background: var(--n-color-embedded, rgba(127, 127, 127, 0.08));
  border-radius: 8px;
  padding: 16px;
  max-height: 480px;
  overflow-y: auto;
}
</style>
