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
                会按会话结构提取字段。④ 非合法 JSON 时用正则从文本中提取上述字段。
                <br />
                输出为完整会话 JSON（含
                <code class="rounded bg-gray-100 px-1">WARNING_BANNER</code>、
                <code class="rounded bg-gray-100 px-1">user</code>、
                <code class="rounded bg-gray-100 px-1">expires</code>、
                <code class="rounded bg-gray-100 px-1">account</code>、
                <code class="rounded bg-gray-100 px-1">accessToken</code>、
                <code class="rounded bg-gray-100 px-1">authProvider</code>、
                <code class="rounded bg-gray-100 px-1">sessionToken</code>、
                <code class="rounded bg-gray-100 px-1">rumViewTags</code>）。
                JWT claims 能解出的字段优先使用，其余缺失字段用默认/随机值补齐。
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
                生成规则与「卡密充值token生成」相同，输出完整会话 JSON。
                格式为 <code class="rounded bg-gray-100 px-1">account----json</code>（每行一条）。
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

const WARNING_BANNER =
  '!!!!!!!!!!!!!!!!!!!! DO NOT SHARE ANY PART OF THE INFORMATION YOU SEE HERE. THIS INFORMATION IS SENSITIVE AND CAN GRANT ACCESS TO YOUR ACCOUNT. SHARING THIS INFORMATION IS LIKE SHARING YOUR PASSWORD. !!!!!!!!!!!!!!!!!!!!'

/** 与 ChatGPT 会话 JSON 一致的载荷结构 */
interface TokenPayload {
  WARNING_BANNER: string
  user: {
    id: string
    name: string
    email: string
    iat: number
    amr: string[]
    mfa: boolean
  }
  expires: string
  account: {
    id: string
    createdTime: number
    planType: string
    structure: string
    isUsageBasedSeatEnabled: boolean
    isConversationClassifierEnabledForWorkspace: boolean
    hasFloraFeature: boolean
    isFedrampCompliantWorkspace: boolean
    isDelinquent: boolean
    residencyRegion: string
    computeResidency: string
  }
  accessToken: string
  authProvider: string
  sessionToken: string
  rumViewTags: { light_account: { fetched: boolean } }
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

function randomBase64Url(byteLen: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_'
  let s = ''
  for (let i = 0; i < byteLen; i++) s += chars[Math.floor(Math.random() * chars.length)]
  return s
}

/** 占位 sessionToken（JWE 形态），输入未提供时补齐 */
function generateSessionToken(): string {
  return `eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..${randomBase64Url(16)}.${randomBase64Url(800)}.${randomBase64Url(22)}`
}

function unixToIso(sec: number): string {
  const ms = sec > 1e12 ? sec : sec * 1000
  const d = new Date(ms)
  return Number.isNaN(d.getTime()) ? new Date(Date.now() + 10 * 24 * 3600 * 1000).toISOString() : d.toISOString()
}

function defaultPayload(accessToken: string): TokenPayload {
  return {
    WARNING_BANNER,
    user: {
      id: '',
      name: '',
      email: '',
      iat: 0,
      amr: ['otp', 'urn:openai:amr:otp_email'],
      mfa: false,
    },
    expires: '',
    account: {
      id: '',
      createdTime: 0,
      planType: 'free',
      structure: 'personal',
      isUsageBasedSeatEnabled: false,
      isConversationClassifierEnabledForWorkspace: true,
      hasFloraFeature: false,
      isFedrampCompliantWorkspace: false,
      isDelinquent: false,
      residencyRegion: 'no_constraint',
      computeResidency: 'no_constraint',
    },
    accessToken,
    authProvider: 'openai',
    sessionToken: '',
    rumViewTags: { light_account: { fetched: false } },
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

/** 从 access_token(JWT) 的 claims 中提取真实账号信息 */
function extractClaims(token: string): {
  accountId: string
  planType: string
  userId: string
  email: string
  name: string
  iat: number
  exp: number
  amr: string[]
  computeResidency: string
} {
  const result = {
    accountId: '',
    planType: '',
    userId: '',
    email: '',
    name: '',
    iat: 0,
    exp: 0,
    amr: [] as string[],
    computeResidency: '',
  }
  const payload = decodeJwtPayload(token)
  if (!payload) return result

  const auth = payload['https://api.openai.com/auth']
  if (auth !== null && typeof auth === 'object' && !Array.isArray(auth)) {
    const a = auth as Record<string, unknown>
    if (a.chatgpt_account_id != null) result.accountId = String(a.chatgpt_account_id)
    if (a.chatgpt_plan_type != null) result.planType = String(a.chatgpt_plan_type)
    if (a.chatgpt_user_id != null) result.userId = String(a.chatgpt_user_id)
    if (a.chatgpt_compute_residency != null) result.computeResidency = String(a.chatgpt_compute_residency)
    if (Array.isArray(a.amr)) result.amr = a.amr.map((x) => String(x))
  }

  const profile = payload['https://api.openai.com/profile']
  if (profile !== null && typeof profile === 'object' && !Array.isArray(profile)) {
    const p = profile as Record<string, unknown>
    if (p.email != null) result.email = String(p.email)
    if (p.name != null) result.name = String(p.name)
  }

  if (typeof payload.iat === 'number') result.iat = payload.iat
  if (typeof payload.exp === 'number') result.exp = payload.exp

  return result
}

/**
 * 填充字段：优先用 accessToken(JWT) / 输入已有值，缺失项用默认或随机值补齐。
 */
function fillMissingId(payload: TokenPayload): TokenPayload {
  const claims = extractClaims(payload.accessToken)
  if (!payload.account.id && claims.accountId) payload.account.id = claims.accountId
  if (payload.account.planType === 'free' && claims.planType) payload.account.planType = claims.planType
  if (!payload.user.id && claims.userId) payload.user.id = claims.userId
  if (!payload.user.email && claims.email) payload.user.email = claims.email
  if (!payload.user.name && claims.name) payload.user.name = claims.name
  if (!payload.user.iat && claims.iat) payload.user.iat = claims.iat
  if (claims.amr.length > 0) payload.user.amr = claims.amr
  if (payload.account.computeResidency === 'no_constraint' && claims.computeResidency) {
    payload.account.computeResidency = claims.computeResidency
  }
  if (!payload.expires && claims.exp) payload.expires = unixToIso(claims.exp)

  const nowSec = Math.floor(Date.now() / 1000)
  if (!payload.account.id) payload.account.id = generateUuid()
  if (!payload.user.id) payload.user.id = generateUserId()
  if (!payload.user.name && payload.user.email) payload.user.name = payload.user.email.split('@')[0]
  if (!payload.user.name) payload.user.name = 'user' + Math.floor(Math.random() * 10000)
  if (!payload.user.email) payload.user.email = `${payload.user.name}@example.com`
  if (!payload.user.iat) payload.user.iat = nowSec
  if (!payload.sessionToken) payload.sessionToken = generateSessionToken()
  if (!payload.WARNING_BANNER) payload.WARNING_BANNER = WARNING_BANNER
  if (!payload.authProvider) payload.authProvider = 'openai'
  if (!payload.expires) payload.expires = unixToIso(payload.user.iat + 10 * 24 * 3600)
  if (!payload.account.createdTime) {
    payload.account.createdTime = Number(((claims.iat || payload.user.iat) - 2965384 + Math.random()).toFixed(3))
  }
  return payload
}

function asBool(v: unknown, fallback: boolean): boolean {
  if (typeof v === 'boolean') return v
  if (v === 'true') return true
  if (v === 'false') return false
  return fallback
}

function asNum(v: unknown, fallback: number): number {
  if (typeof v === 'number' && !Number.isNaN(v)) return v
  if (typeof v === 'string' && v.trim() !== '') {
    const n = Number(v)
    if (!Number.isNaN(n)) return n
  }
  return fallback
}

function asAmr(v: unknown, fallback: string[]): string[] {
  if (Array.isArray(v) && v.length > 0) return v.map((x) => String(x))
  return fallback
}

/** 合法 JSON：按会话结构提取字段 */
function parseJsonToPayload(obj: unknown): TokenPayload | null {
  if (obj === null || typeof obj !== 'object' || Array.isArray(obj)) return null
  const rec = obj as Record<string, unknown>
  if (rec.accessToken == null || String(rec.accessToken).length === 0) return null

  const payload = defaultPayload(String(rec.accessToken))

  if (rec.WARNING_BANNER != null) payload.WARNING_BANNER = String(rec.WARNING_BANNER)
  if (rec.expires != null) payload.expires = String(rec.expires)
  if (rec.authProvider != null) payload.authProvider = String(rec.authProvider)
  if (rec.sessionToken != null) payload.sessionToken = String(rec.sessionToken)

  if (rec.account !== null && typeof rec.account === 'object' && !Array.isArray(rec.account)) {
    const acc = rec.account as Record<string, unknown>
    if (acc.id != null) payload.account.id = String(acc.id)
    if ('planType' in acc) payload.account.planType = acc.planType == null ? '' : String(acc.planType)
    if ('structure' in acc) payload.account.structure = acc.structure == null ? 'personal' : String(acc.structure)
    if ('createdTime' in acc) payload.account.createdTime = asNum(acc.createdTime, payload.account.createdTime)
    if ('isUsageBasedSeatEnabled' in acc) {
      payload.account.isUsageBasedSeatEnabled = asBool(acc.isUsageBasedSeatEnabled, false)
    }
    if ('isConversationClassifierEnabledForWorkspace' in acc) {
      payload.account.isConversationClassifierEnabledForWorkspace = asBool(
        acc.isConversationClassifierEnabledForWorkspace,
        true,
      )
    }
    if ('hasFloraFeature' in acc) payload.account.hasFloraFeature = asBool(acc.hasFloraFeature, false)
    if ('isFedrampCompliantWorkspace' in acc) {
      payload.account.isFedrampCompliantWorkspace = asBool(acc.isFedrampCompliantWorkspace, false)
    }
    if ('isDelinquent' in acc) payload.account.isDelinquent = asBool(acc.isDelinquent, false)
    if (acc.residencyRegion != null) payload.account.residencyRegion = String(acc.residencyRegion)
    if (acc.computeResidency != null) payload.account.computeResidency = String(acc.computeResidency)
  }

  if (rec.user !== null && typeof rec.user === 'object' && !Array.isArray(rec.user)) {
    const u = rec.user as Record<string, unknown>
    if (u.id != null) payload.user.id = String(u.id)
    if (u.name != null) payload.user.name = String(u.name)
    if (u.email != null) payload.user.email = String(u.email)
    if ('iat' in u) payload.user.iat = asNum(u.iat, payload.user.iat)
    if ('amr' in u) payload.user.amr = asAmr(u.amr, payload.user.amr)
    if ('mfa' in u) payload.user.mfa = asBool(u.mfa, false)
  }

  if (rec.rumViewTags !== null && typeof rec.rumViewTags === 'object' && !Array.isArray(rec.rumViewTags)) {
    const tags = rec.rumViewTags as Record<string, unknown>
    const light = tags.light_account
    if (light !== null && typeof light === 'object' && !Array.isArray(light)) {
      const l = light as Record<string, unknown>
      payload.rumViewTags = { light_account: { fetched: asBool(l.fetched, false) } }
    }
  }

  return payload
}

/** 非合法 JSON：正则提取（与输出结构对应字段） */
function parseRegexToPayload(s: string): TokenPayload | null {
  const accessM = s.match(/"accessToken"\s*:\s*"([^"]*)"/)
  if (!accessM) return null
  const payload = defaultPayload(accessM[1])

  const sessionM = s.match(/"sessionToken"\s*:\s*"([^"]*)"/)
  if (sessionM) payload.sessionToken = sessionM[1]
  const expiresM = s.match(/"expires"\s*:\s*"([^"]*)"/)
  if (expiresM) payload.expires = expiresM[1]
  const authM = s.match(/"authProvider"\s*:\s*"([^"]*)"/)
  if (authM) payload.authProvider = authM[1]

  const accountBlock = s.match(/"account"\s*:\s*\{([^}]*)\}/)
  if (accountBlock) {
    const inner = accountBlock[1]
    const idM = inner.match(/"id"\s*:\s*"([^"]*)"/)
    const ptM = inner.match(/"planType"\s*:\s*"([^"]*)"/)
    const stM = inner.match(/"structure"\s*:\s*"([^"]*)"/)
    const ctM = inner.match(/"createdTime"\s*:\s*([0-9.]+)/)
    const crM = inner.match(/"computeResidency"\s*:\s*"([^"]*)"/)
    const rrM = inner.match(/"residencyRegion"\s*:\s*"([^"]*)"/)
    if (idM) payload.account.id = idM[1]
    if (ptM) payload.account.planType = ptM[1]
    if (stM) payload.account.structure = stM[1]
    if (ctM) payload.account.createdTime = Number(ctM[1])
    if (crM) payload.account.computeResidency = crM[1]
    if (rrM) payload.account.residencyRegion = rrM[1]
  }

  const userBlock = s.match(/"user"\s*:\s*\{([^}]*)\}/)
  if (userBlock) {
    const inner = userBlock[1]
    const uIdM = inner.match(/"id"\s*:\s*"([^"]*)"/)
    const uNameM = inner.match(/"name"\s*:\s*"([^"]*)"/)
    const emM = inner.match(/"email"\s*:\s*"([^"]*)"/)
    const iatM = inner.match(/"iat"\s*:\s*([0-9.]+)/)
    const mfaM = inner.match(/"mfa"\s*:\s*(true|false)/)
    if (uIdM) payload.user.id = uIdM[1]
    if (uNameM) payload.user.name = uNameM[1]
    if (emM) payload.user.email = emM[1]
    if (iatM) payload.user.iat = Number(iatM[1])
    if (mfaM) payload.user.mfa = mfaM[1] === 'true'
  }

  return payload
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
