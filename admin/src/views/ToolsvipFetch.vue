<template>
  <div class="webmail-fetch-page">
    <div class="page-head">
      <h1 class="page-title">toolsvip 取件</h1>
      <p class="page-subtitle">通过 toolsvip API 拉取收件箱与垃圾箱邮件</p>
    </div>

    <n-card :bordered="false" class="input-card">
      <template #header><span class="card-title">账号信息</span></template>
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
          :autosize="{ minRows: 2, maxRows: 5 }"
          :disabled="loading"
        />
        <div class="action-row">
          <n-text v-if="fetchData" depth="3" style="font-size:13px">
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
        class="mail-card"
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
            :key="idx"
            @click="openDetail(mail)"
          >
            <div class="mail-row">
              <div class="mail-meta">
                <span class="mail-subject">{{ mail.subject }}</span>
                <span class="mail-time">{{ mail.date }}</span>
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
            <span class="detail-label">时间</span>
            <span class="detail-value">{{ detailMail.date }}</span>
          </div>
          <div v-if="detailDisplayCode" class="detail-meta-row">
            <span class="detail-label">验证码</span>
            <n-button type="success" size="small" round @click="copyText(detailDisplayCode)">
              {{ detailDisplayCode }}
            </n-button>
          </div>
        </div>
        <div class="detail-body-wrap">
          <div v-if="detailLoading" class="detail-loading">
            <n-spin size="small" />
            <span>正在加载正文…</span>
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
import { ref, computed } from 'vue'
import { NButton, NCard, NEmpty, NInput, NList, NListItem, NModal, NSelect, NSpin, NTag, NText, useMessage } from 'naive-ui'
import { fetchToolsvipMails, type ToolsvipFetchData, type ToolsvipMailItem } from '@/api/webmail'
import { useWebmailAccountFormat } from '@/composables/useWebmailAccountFormat'

const message = useMessage()
const accountLine = ref('')
const loading = ref(false)
const fetchData = ref<ToolsvipFetchData | null>(null)
const showDetail = ref(false)
const detailMail = ref<ToolsvipMailItem | null>(null)

const {
  accountFormat,
  formatOptions,
  formatPlaceholder,
  formatHint,
  onFormatChange,
} = useWebmailAccountFormat('toolsvipAccountFormat')

const detailLoading = ref(false)

const detailDisplayCode = computed(() => {
  if (!detailMail.value) return ''
  if (detailMail.value.code) return detailMail.value.code
  const text = `${detailMail.value.subject} ${detailMail.value.body} ${detailMail.value.html_body}`
  const m = text.match(/\b(\d{6})\b/)
  return m?.[1] || ''
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
  if (!line) { message.warning('请输入账号信息'); return }
  loading.value = true
  fetchData.value = null
  try {
    const res = await fetchToolsvipMails(line, accountFormat.value)
    if (res.code !== 200) { message.error(res.message || '取件失败'); return }
    fetchData.value = res.data
  } catch (e: any) {
    message.error(e?.message || '请求失败')
  } finally {
    loading.value = false
  }
}

function openDetail(mail: ToolsvipMailItem) {
  detailMail.value = mail
  showDetail.value = true
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
.webmail-fetch-page { padding: 24px; max-width: 100%; }
.page-head { margin-bottom: 24px; }
.page-title { font-size: 28px; font-weight: 700; letter-spacing: -0.5px; margin: 0 0 6px; color: var(--n-text-color); }
.page-subtitle { font-size: 14px; color: var(--n-text-color-3); margin: 0; }
.input-card { border-radius: 16px; margin-bottom: 24px; }
.card-title { font-weight: 600; font-size: 15px; }
.form-area { display: flex; flex-direction: column; gap: 12px; }
.hint-text { margin: 0; font-size: 13px; color: var(--n-text-color-3); }
.format-row { display: flex; align-items: center; gap: 12px; }
.format-label { font-size: 13px; font-weight: 600; color: var(--n-text-color-2); flex-shrink: 0; }
.action-row { display: flex; justify-content: space-between; align-items: center; margin-top: 4px; }
.mail-panels { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
@media (max-width: 900px) { .mail-panels { grid-template-columns: 1fr; } }
.mail-card { border-radius: 16px; }
.folder-header { display: flex; align-items: center; gap: 8px; }
.folder-title { font-weight: 600; font-size: 15px; }
.mail-row { display: flex; justify-content: space-between; align-items: center; gap: 12px; width: 100%; padding: 2px 0; }
.mail-meta { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; }
.mail-subject { font-size: 14px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--n-text-color); }
.mail-time { font-size: 11px; color: var(--n-text-color-4, #aaa); }
.mail-actions { flex-shrink: 0; }
.detail-meta { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--n-divider-color); }
.detail-meta-row { display: flex; align-items: center; gap: 12px; }
.detail-label { font-size: 13px; font-weight: 600; color: var(--n-text-color-3); width: 56px; flex-shrink: 0; }
.detail-value { font-size: 13px; color: var(--n-text-color); }
.detail-body-wrap { min-height: 200px; }
.detail-loading { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 40px 0; color: #999; }
.detail-iframe { width: 100%; height: 480px; border: none; border-radius: 8px; background: #fff; }
.detail-body { font-size: 13px; line-height: 1.7; white-space: pre-wrap; word-break: break-all; color: var(--n-text-color); margin: 0; background: var(--n-color-embedded, rgba(127,127,127,.08)); border-radius: 8px; padding: 16px; max-height: 480px; overflow-y: auto; }
</style>
