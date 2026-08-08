<template>
  <div>
    <n-space vertical :size="16">
      <div>
        <h1 class="apple-page-title">GPT业务</h1>
        <p class="apple-page-subtitle">卡密充值 Token、支付与账单链接相关工具</p>
      </div>

      <n-card :bordered="false" class="shadow-sm">
        <n-tabs v-model:value="activeTab" type="line" animated>
          <n-tab-pane name="token" tab="卡密充值token生成">
            <n-space vertical :size="16" class="max-w-3xl">
              <n-text depth="3" class="text-sm">
                ① 以 <code class="rounded bg-gray-100 px-1">eyJ</code> 开头则整段作为
                <code class="rounded bg-gray-100 px-1">accessToken</code>。② 格式为
                <code class="rounded bg-gray-100 px-1">account----token</code> 时，将 account 赋值到
                <code class="rounded bg-gray-100 px-1">user.email</code>，token 作为
                <code class="rounded bg-gray-100 px-1">accessToken</code>。③ 合法 JSON
                会填充 <code class="rounded bg-gray-100 px-1">accessToken</code>、<code class="rounded bg-gray-100 px-1">account.structure</code>、<code class="rounded bg-gray-100 px-1">user.email</code>。
                ④ 非合法 JSON 时用正则从文本中提取上述字段。
                <br />
                <strong>account.id / account.planType / user.id / user.email</strong> 会优先从
                <code class="rounded bg-gray-100 px-1">accessToken</code> 的 JWT claims 中解码填充（更准确）；JWT 未提供时才回退默认/随机值。
                <code class="rounded bg-gray-100 px-1">account.structure</code> 不在 JWT 内，默认
                <code class="rounded bg-gray-100 px-1">personal</code>。
                注意：<code class="rounded bg-gray-100 px-1">sessionToken</code> 由服务端加密签发（NextAuth JWE），本地无法伪造，如需真实值请走登录获取 cookie。
              </n-text>

              <n-form label-placement="left" label-width="88px">
                <n-form-item label="输入">
                  <n-input
                    v-model:value="tokenInput"
                    type="textarea"
                    :rows="8"
                    placeholder="粘贴 JWT、整段 JSON 或含 accessToken 的文本"
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-space>
                    <n-button type="primary" @click="handleGenerateTokenJson">生成 JSON</n-button>
                  </n-space>
                </n-form-item>
                <n-alert v-if="tokenError" type="warning" class="mt-1">{{ tokenError }}</n-alert>
                <n-form-item label="输出">
                  <n-input
                    v-model:value="tokenOutput"
                    type="textarea"
                    readonly
                    :rows="14"
                    placeholder="生成结果（固定格式 JSON）"
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-button secondary :disabled="!tokenOutput" @click="handleCopyTokenOutput">一键复制</n-button>
                </n-form-item>
              </n-form>
            </n-space>
          </n-tab-pane>

          <n-tab-pane name="payment-link" tab="提取支付链接">
            <n-space vertical :size="16" class="max-w-3xl">
              <n-text depth="3" class="text-sm">
                输入解析规则与「卡密充值token生成」一致；生成可在控制台执行的
                <code class="rounded bg-gray-100 px-1">fetch</code> 代码，其中
                <code class="rounded bg-gray-100 px-1">authorization</code> 已填入解析得到的 accessToken。
              </n-text>

              <n-form label-placement="left" label-width="88px">
                <n-form-item label="输入">
                  <n-input
                    v-model:value="paymentInput"
                    type="textarea"
                    :rows="8"
                    placeholder="粘贴 JWT、整段 JSON 或含 accessToken 的文本"
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-space>
                    <n-button type="primary" @click="handleGeneratePaymentFetch">生成 fetch 代码</n-button>
                  </n-space>
                </n-form-item>
                <n-alert v-if="paymentError" type="warning" class="mt-1">{{ paymentError }}</n-alert>
                <n-form-item label="输出">
                  <n-input
                    v-model:value="paymentOutput"
                    type="textarea"
                    readonly
                    :rows="18"
                    placeholder="checkout fetch 代码"
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-button secondary :disabled="!paymentOutput" @click="handleCopyPaymentOutput">一键复制</n-button>
                </n-form-item>
              </n-form>
            </n-space>
          </n-tab-pane>

          <n-tab-pane name="batch-token" tab="批量提token">
            <n-space vertical :size="16" class="max-w-3xl">
              <n-text depth="3" class="text-sm">
                每行一条
                <code class="rounded bg-gray-100 px-1">account----access_token</code>，
                生成规则与「卡密充值token生成」相同：将 account 写入
                <code class="rounded bg-gray-100 px-1">user.email</code>，token 作为
                <code class="rounded bg-gray-100 px-1">accessToken</code>，
                并从 JWT claims 补全 account / user 字段。
                输出格式为 <code class="rounded bg-gray-100 px-1">account----json</code>（每行一条）。
              </n-text>

              <n-form label-placement="left" label-width="88px">
                <n-form-item label="输入">
                  <n-input
                    v-model:value="batchTokenInput"
                    type="textarea"
                    :rows="10"
                    placeholder="每行一条：account----access_token&#10;user@example.com----eyJhbGciOi..."
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-space>
                    <n-button type="primary" @click="handleBatchExtractToken">批量生成</n-button>
                  </n-space>
                </n-form-item>
                <n-alert v-if="batchTokenError" type="warning" class="mt-1">{{ batchTokenError }}</n-alert>
                <n-form-item label="输出">
                  <n-input
                    v-model:value="batchTokenOutput"
                    type="textarea"
                    readonly
                    :rows="14"
                    placeholder="account----{...json...}（每行一条）"
                    class="font-mono text-sm"
                  />
                </n-form-item>
                <n-form-item>
                  <n-button secondary :disabled="!batchTokenOutput" @click="handleCopyBatchTokenOutput">
                    一键复制
                  </n-button>
                </n-form-item>
              </n-form>
            </n-space>
          </n-tab-pane>

          <n-tab-pane name="bill-link" tab="提取账单链接">
            <n-form label-placement="left" label-width="140px" class="max-w-2xl">
              <n-form-item label="说明">
                <n-text depth="3">从通知或原文中提取账单 / 发票相关链接。</n-text>
              </n-form-item>
              <n-form-item label="原始文本">
                <n-input type="textarea" :rows="6" placeholder="粘贴含账单链接的文本" />
              </n-form-item>
              <n-form-item>
                <n-space>
                  <n-button type="primary" disabled>提取链接（待接入）</n-button>
                  <n-button disabled>复制结果</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  NSpace,
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NText,
  NAlert,
  useMessage,
} from 'naive-ui'

const activeTab = ref<'token' | 'payment-link' | 'batch-token' | 'bill-link'>('token')
const message = useMessage()

/** 与后端输出一致的载荷结构 */
interface TokenPayload {
  accessToken: string
  account: { id: string; planType: string; structure: string }
  user: { id: string; name: string; email: string }
}

function generateUuid(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

/** 生成 user-XXXX 格式 ID，字符集与官方一致（大小写字母+数字，25位） */
function generateUserId(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let s = ''
  for (let i = 0; i < 25; i++) s += chars[Math.floor(Math.random() * chars.length)]
  return 'user-' + s
}

function defaultPayload(accessToken: string): TokenPayload {
  return {
    accessToken,
    account: { id: '', planType: 'free', structure: 'personal' },
    user: { id: '', name: '', email: '' },
  }
}

/** 解码 JWT 的 payload 段（base64url），失败返回 null */
function decodeJwtPayload(token: string): Record<string, unknown> | null {
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    let b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    while (b64.length % 4) b64 += '='
    const json = decodeURIComponent(
      atob(b64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join(''),
    )
    const obj = JSON.parse(json) as unknown
    if (obj !== null && typeof obj === 'object' && !Array.isArray(obj)) {
      return obj as Record<string, unknown>
    }
    return null
  } catch {
    return null
  }
}

/** 从 access_token(JWT) 的 claims 中提取真实账号信息（无 structure，该字段不在 JWT 内） */
function extractClaims(token: string): {
  accountId: string
  planType: string
  userId: string
  email: string
} {
  const result = { accountId: '', planType: '', userId: '', email: '' }
  const payload = decodeJwtPayload(token)
  if (!payload) return result

  const auth = payload['https://api.openai.com/auth']
  if (auth !== null && typeof auth === 'object' && !Array.isArray(auth)) {
    const a = auth as Record<string, unknown>
    if (a.chatgpt_account_id != null) result.accountId = String(a.chatgpt_account_id)
    if (a.chatgpt_plan_type != null) result.planType = String(a.chatgpt_plan_type)
    if (a.chatgpt_user_id != null) result.userId = String(a.chatgpt_user_id)
  }

  const profile = payload['https://api.openai.com/profile']
  if (profile !== null && typeof profile === 'object' && !Array.isArray(profile)) {
    const p = profile as Record<string, unknown>
    if (p.email != null) result.email = String(p.email)
  }

  return result
}

/**
 * 填充字段：优先用 accessToken(JWT) 中的真实信息（account.id / planType / user.id / user.email），
 * JWT 未提供的字段再回退到随机/派生值（account.id → UUID，user.id → user-XXXX，user.name → email 前缀）。
 */
function fillMissingId(payload: TokenPayload): TokenPayload {
  const claims = extractClaims(payload.accessToken)
  if (claims.accountId) payload.account.id = claims.accountId
  if (claims.planType) payload.account.planType = claims.planType
  if (claims.userId) payload.user.id = claims.userId
  if (claims.email && !payload.user.email) payload.user.email = claims.email

  if (!payload.account.id) {
    payload.account.id = generateUuid()
  }
  if (!payload.user.id) {
    payload.user.id = generateUserId()
  }
  if (!payload.user.name && payload.user.email) {
    payload.user.name = payload.user.email.split('@')[0]
  }
  return payload
}

/** 合法 JSON：按输出结构提取 accessToken、account、user */
function parseJsonToPayload(obj: unknown): TokenPayload | null {
  if (obj === null || typeof obj !== 'object' || Array.isArray(obj)) return null
  const rec = obj as Record<string, unknown>
  if (rec.accessToken == null || String(rec.accessToken).length === 0) return null

  const accessToken = String(rec.accessToken)

  let id = ''
  let planType = 'free'
  let structure = 'personal'
  if (rec.account !== null && typeof rec.account === 'object' && !Array.isArray(rec.account)) {
    const acc = rec.account as Record<string, unknown>
    if (acc.id != null) id = String(acc.id)
    if ('planType' in acc) {
      planType = acc.planType == null ? '' : String(acc.planType)
    }
    if ('structure' in acc) {
      structure = acc.structure == null ? 'personal' : String(acc.structure)
    }
  }

  let userId = ''
  let userName = ''
  let email = ''
  if (rec.user !== null && typeof rec.user === 'object' && !Array.isArray(rec.user)) {
    const u = rec.user as Record<string, unknown>
    if (u.id != null) userId = String(u.id)
    if (u.name != null) userName = String(u.name)
    if (u.email != null) email = String(u.email)
  }

  return {
    accessToken,
    account: { id, planType, structure },
    user: { id: userId, name: userName, email },
  }
}

/** 非合法 JSON：正则提取（与输出结构对应字段） */
function parseRegexToPayload(s: string): TokenPayload | null {
  const accessM = s.match(/"accessToken"\s*:\s*"([^"]*)"/)
  if (!accessM) return null
  const accessToken = accessM[1]

  let id = ''
  let planType = 'free'
  let structure = 'personal'
  const accountBlock = s.match(/"account"\s*:\s*\{([^}]*)\}/)
  if (accountBlock) {
    const inner = accountBlock[1]
    const idM = inner.match(/"id"\s*:\s*"([^"]*)"/)
    const ptM = inner.match(/"planType"\s*:\s*"([^"]*)"/)
    const stM = inner.match(/"structure"\s*:\s*"([^"]*)"/)
    if (idM) id = idM[1]
    if (ptM) planType = ptM[1]
    if (stM) structure = stM[1]
  }

  let userId = ''
  let userName = ''
  let email = ''
  const userBlock = s.match(/"user"\s*:\s*\{([^}]*)\}/)
  if (userBlock) {
    const inner = userBlock[1]
    const uIdM = inner.match(/"id"\s*:\s*"([^"]*)"/)
    const uNameM = inner.match(/"name"\s*:\s*"([^"]*)"/)
    const emM = inner.match(/"email"\s*:\s*"([^"]*)"/)
    if (uIdM) userId = uIdM[1]
    if (uNameM) userName = uNameM[1]
    if (emM) email = emM[1]
  }

  return {
    accessToken,
    account: { id, planType, structure },
    user: { id: userId, name: userName, email },
  }
}

/** 从输入解析为固定输出结构 */
function extractTokenPayload(raw: string): TokenPayload | null {
  const s = raw.trim()
  if (!s) return null

  // account----token 格式：左侧为邮箱，右侧为 accessToken
  if (s.includes('----')) {
    const idx = s.indexOf('----')
    const account = s.slice(0, idx).trim()
    const token = s.slice(idx + 4).trim()
    if (token) {
      const payload = defaultPayload(token)
      payload.user.email = account
      return fillMissingId(payload)
    }
  }

  if (s.startsWith('eyJ')) {
    return fillMissingId(defaultPayload(s))
  }

  try {
    const obj = JSON.parse(s) as unknown
    const payload = parseJsonToPayload(obj)
    return payload ? fillMissingId(payload) : null
  } catch {
    const payload = parseRegexToPayload(s)
    return payload ? fillMissingId(payload) : null
  }
}

const tokenInput = ref('')
const tokenOutput = ref('')
const tokenError = ref('')

const handleGenerateTokenJson = () => {
  tokenError.value = ''
  const payload = extractTokenPayload(tokenInput.value)
  if (!payload) {
    tokenOutput.value = ''
    tokenError.value =
      '未能解析：请使用 eyJ 开头 JWT、含 accessToken（及可选 account/user）的合法 JSON，或非 JSON 文本中匹配 "accessToken":"..." 等字段。'
    return
  }
  tokenOutput.value = JSON.stringify(payload, null, 2)
}

const handleCopyTokenOutput = async () => {
  if (!tokenOutput.value) {
    message.warning('暂无内容可复制')
    return
  }
  try {
    await navigator.clipboard.writeText(tokenOutput.value)
    message.success('已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动选中复制')
  }
}

/** 嵌入 JS 模板字符串时对 token 转义 */
function escapeForTemplateLiteral(s: string): string {
  return s.replace(/\\/g, '\\\\').replace(/`/g, '\\`').replace(/\$\{/g, '\\${')
}

/** 生成支付 checkout fetch 片段（authorization 使用模板字符串包裹 token） */
function buildPaymentCheckoutFetchSnippet(accessToken: string): string {
  const t = escapeForTemplateLiteral(accessToken)
  return `fetch("/backend-api/payments/checkout", {
    "method": "POST",
    "headers": {
      "authorization": \`${t}\`,
      "Content-Type": "application/json",
    },
    "body": JSON.stringify({
      "plan_name": "chatgptplusplan",
      "billing_details": {
        "country": "PH",
        "currency": "PHP"
      },
      "checkout_ui_mode": "redirect"
    })
  }).then(r => r.json()).then(d => window.open(d.url))`
}

const paymentInput = ref('')
const paymentOutput = ref('')
const paymentError = ref('')

const handleGeneratePaymentFetch = () => {
  paymentError.value = ''
  const payload = extractTokenPayload(paymentInput.value)
  if (!payload) {
    paymentOutput.value = ''
    paymentError.value =
      '未能解析：请使用 eyJ 开头 JWT、含 accessToken（及可选 account/user）的合法 JSON，或非 JSON 文本中匹配 "accessToken":"..." 等字段。'
    return
  }
  paymentOutput.value = buildPaymentCheckoutFetchSnippet(payload.accessToken)
}

const handleCopyPaymentOutput = async () => {
  if (!paymentOutput.value) {
    message.warning('暂无内容可复制')
    return
  }
  try {
    await navigator.clipboard.writeText(paymentOutput.value)
    message.success('已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动选中复制')
  }
}

const batchTokenInput = ref('')
const batchTokenOutput = ref('')
const batchTokenError = ref('')

const handleBatchExtractToken = () => {
  batchTokenError.value = ''
  const lines = batchTokenInput.value
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter((l) => l.length > 0)

  if (lines.length === 0) {
    batchTokenOutput.value = ''
    batchTokenError.value = '请先粘贴内容'
    return
  }

  const results: string[] = []
  const failLines: number[] = []

  lines.forEach((line, idx) => {
    const payload = extractTokenPayload(line)
    if (!payload?.accessToken) {
      failLines.push(idx + 1)
      return
    }
    const account = payload.user.email.trim()
    results.push(`${account}----${JSON.stringify(payload)}`)
  })

  if (results.length === 0) {
    batchTokenOutput.value = ''
    batchTokenError.value =
      '未能解析任何行：请使用 account----access_token 格式（每行一条）。'
    return
  }

  batchTokenOutput.value = results.join('\n')
  if (failLines.length > 0) {
    batchTokenError.value = `已生成 ${results.length} 条；第 ${failLines.join('、')} 行解析失败已跳过。`
  }
}

const handleCopyBatchTokenOutput = async () => {
  if (!batchTokenOutput.value) {
    message.warning('暂无内容可复制')
    return
  }
  try {
    await navigator.clipboard.writeText(batchTokenOutput.value)
    message.success('已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动选中复制')
  }
}
</script>

<style scoped>
:deep(.n-card) {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

:deep(.n-tabs .n-tab-pane) {
  padding-top: 24px;
}
</style>
