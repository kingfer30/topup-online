import { ref, computed, onMounted } from 'vue'

export interface WebmailFormatConfig {
  value: string
  label: string
  placeholder: string
  hint: string
}

export const WEBMAIL_FORMAT_CONFIGS: WebmailFormatConfig[] = [
  {
    value: '1',
    label: '格式1: 邮箱----邮箱密码',
    placeholder: 'user@hotmail.com----邮箱密码',
    hint: '每行一个账号；也兼容 邮箱|邮箱密码（竖线仅 2 段时）',
  },
  {
    value: '2',
    label: '格式2: 邮箱----GPT密码----邮箱密码',
    placeholder: 'user@hotmail.com----GPT密码----邮箱密码',
    hint: 'GPT密码字段会被忽略，取件使用邮箱密码',
  },
  {
    value: '3',
    label: '格式3: 邮箱----邮箱密码----GPT密码',
    placeholder: 'user@hotmail.com----邮箱密码----GPT密码',
    hint: 'GPT密码字段会被忽略，取件使用邮箱密码',
  },
]

function migrateWebmailFormat(saved: string | null) {
  if (!saved) return '1'
  const legacy: Record<string, string> = { f1: '1', f2: '2', f3: '3', f4: '2', f5: '3' }
  return legacy[saved] || saved
}

export function useWebmailAccountFormat(storageKey: string) {
  const accountFormat = ref('1')

  const formatOptions = WEBMAIL_FORMAT_CONFIGS.map(c => ({ label: c.label, value: c.value }))

  const currentFormatConfig = computed(() =>
    WEBMAIL_FORMAT_CONFIGS.find(c => c.value === accountFormat.value) || WEBMAIL_FORMAT_CONFIGS[0],
  )

  const formatPlaceholder = computed(() => currentFormatConfig.value.placeholder)
  const formatHint = computed(() => currentFormatConfig.value.hint)

  function onFormatChange(value: string) {
    localStorage.setItem(storageKey, value)
  }

  onMounted(() => {
    accountFormat.value = migrateWebmailFormat(localStorage.getItem(storageKey))
  })

  return {
    accountFormat,
    formatOptions,
    formatPlaceholder,
    formatHint,
    onFormatChange,
  }
}
