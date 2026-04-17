<template>
  <div class="ai-translate-page">
    <div class="page-head">
      <h1 class="apple-page-title">AI 翻译</h1>
      <p class="apple-page-subtitle">在左侧输入原文，选择语言后点击「翻译」获取译文（非实时）</p>
    </div>

    <n-card :bordered="false" class="shadow-sm translate-card">
      <div class="lang-toolbar">
        <div class="lang-side">
          <div class="lang-label">源语言</div>
          <div class="lang-chips">
            <button
              v-for="opt in sourceQuick"
              :key="opt.value"
              type="button"
              class="chip"
              :class="{ active: sourceLang === opt.value }"
              @click="sourceLang = opt.value"
            >
              {{ opt.label }}
            </button>
            <n-select
              v-model:value="sourceLang"
              :options="allLangOptions"
              size="small"
              class="lang-more"
              placeholder="更多语言"
            />
          </div>
        </div>

        <n-button quaternary circle class="swap-btn" @click="swapLangs" aria-label="交换语言">
          ⇄
        </n-button>

        <div class="lang-side">
          <div class="lang-label">目标语言</div>
          <div class="lang-chips">
            <button
              v-for="opt in targetQuick"
              :key="opt.value"
              type="button"
              class="chip"
              :class="{ active: targetLang === opt.value }"
              @click="targetLang = opt.value"
            >
              {{ opt.label }}
            </button>
            <n-select
              v-model:value="targetLang"
              :options="targetSelectOptions"
              size="small"
              class="lang-more"
              placeholder="更多语言"
            />
          </div>
        </div>
      </div>

      <div class="panels">
        <div class="panel panel-in">
          <div class="panel-actions">
            <n-button size="small" quaternary circle aria-label="复制原文" title="复制原文" @click="copyText(sourceText)">
              ⧉
            </n-button>
          </div>
          <n-input
            v-model:value="sourceText"
            type="textarea"
            :autosize="{ minRows: 14, maxRows: 24 }"
            placeholder="请输入文本"
            class="area-input"
          />
        </div>

        <div class="panel-mid">
          <n-button type="primary" size="large" round :loading="loading" @click="handleTranslate">
            翻译
          </n-button>
        </div>

        <div class="panel panel-out">
          <div class="panel-actions">
            <n-button size="small" quaternary circle aria-label="复制译文" title="复制译文" @click="copyText(outputText)">
              ⧉
            </n-button>
          </div>
          <n-input
            v-model:value="outputText"
            type="textarea"
            readonly
            placeholder="译文将显示在这里"
            :autosize="{ minRows: 14, maxRows: 24 }"
            class="area-output"
          />
        </div>
      </div>

      <div class="page-foot">
        <n-text depth="3" style="font-size: 12px">模型配置见：系统设置 → AI模型设置</n-text>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { NButton, NCard, NInput, NSelect, NText, useMessage } from 'naive-ui'
import { adminAITranslate } from '@/api/ai'

const message = useMessage()

const LANGS = [
  { label: '检测语言', value: 'auto' },
  { label: '中文（简体）', value: 'zh' },
  { label: '英语', value: 'en' },
  { label: '日语', value: 'ja' },
  { label: '俄语', value: 'ru' },
  { label: '韩语', value: 'ko' },
  { label: '法语', value: 'fr' },
  { label: '德语', value: 'de' },
  { label: '西班牙语', value: 'es' },
] as const

const sourceQuick = [
  { label: '检测语言', value: 'auto' },
  { label: '中文', value: 'zh' },
  { label: '俄语', value: 'ru' },
  { label: '英语', value: 'en' },
]

const targetQuick = [
  { label: '俄语', value: 'ru' },
  { label: '英语', value: 'en' },
  { label: '中文', value: 'zh' },
]

const allLangOptions = LANGS.map((x) => ({ label: x.label, value: x.value }))

const targetSelectOptions = computed(() =>
  LANGS.filter((x) => x.value !== 'auto').map((x) => ({ label: x.label, value: x.value }))
)

const sourceLang = ref<string>('auto')
const targetLang = ref<string>('ru')
const sourceText = ref('')
const outputText = ref('')
const loading = ref(false)

const copyText = async (text: string) => {
  const v = (text || '').trim()
  if (!v) {
    message.warning('没有可复制的内容')
    return
  }
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(v)
      message.success('已复制')
      return
    }
  } catch {
    // fallthrough to legacy method
  }
  try {
    const textarea = document.createElement('textarea')
    textarea.value = v
    textarea.setAttribute('readonly', 'true')
    textarea.style.position = 'fixed'
    textarea.style.top = '-9999px'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(textarea)
    if (ok) {
      message.success('已复制')
    } else {
      message.error('复制失败')
    }
  } catch {
    message.error('复制失败')
  }
}

const swapLangs = () => {
  const s = sourceLang.value
  const t = targetLang.value
  if (s === 'auto') {
    sourceLang.value = t
    targetLang.value = 'ru'
    return
  }
  sourceLang.value = t
  targetLang.value = s
}

const handleTranslate = async () => {
  const text = sourceText.value.trim()
  if (!text) {
    message.warning('请输入要翻译的内容')
    return
  }
  if (!targetLang.value || targetLang.value === 'auto') {
    message.warning('请选择目标语言')
    return
  }
  loading.value = true
  outputText.value = ''
  try {
    const res = await adminAITranslate({
      text,
      source_lang: sourceLang.value,
      target_lang: targetLang.value,
    })
    if (res.code === 200 && res.data?.translation) {
      outputText.value = res.data.translation
    } else {
      message.error(res.message || '翻译失败')
    }
  } catch {
    message.error('请求失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.ai-translate-page {
  max-width: 1200px;
}

.page-head {
  margin-bottom: 16px;
}

.translate-card {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.06) !important;
}

.lang-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.lang-side {
  flex: 1;
  min-width: 200px;
}

.lang-label {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 6px;
}

.lang-chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.chip {
  border: none;
  background: transparent;
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  border-bottom: 2px solid transparent;
}

.chip:hover {
  background: rgba(59, 130, 246, 0.08);
}

.chip.active {
  color: #2563eb;
  border-bottom-color: #2563eb;
  font-weight: 600;
}

.lang-more {
  width: 130px;
}

.swap-btn {
  font-size: 18px;
  margin-bottom: 2px;
  flex-shrink: 0;
}

.panels {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 16px;
  align-items: stretch;
}

.panel {
  min-height: 280px;
  position: relative;
}

.panel-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 2;
}

.panel-mid {
  display: flex;
  align-items: center;
  justify-content: center;
}

.area-input :deep(textarea) {
  border-radius: 12px;
  font-size: 15px;
  line-height: 1.6;
}

.area-output :deep(textarea) {
  border-radius: 12px;
  font-size: 15px;
  line-height: 1.6;
  background: #f8fafc !important;
}

.page-foot {
  margin-top: 12px;
  text-align: right;
}

@media (max-width: 900px) {
  .panels {
    grid-template-columns: 1fr;
  }

  .panel-mid {
    padding: 8px 0;
  }
}
</style>
