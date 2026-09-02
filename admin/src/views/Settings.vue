<template>
  <div>
    <n-space vertical :size="16">
      <!-- 页面标题 -->
      <div>
        <h1 class="apple-page-title">系统设置</h1>
        <p class="apple-page-subtitle">配置系统基本参数</p>
      </div>

      <!-- 设置表单 -->
      <n-card :bordered="false" class="shadow-sm">
        <n-tabs type="line" animated>
          <n-tab-pane name="basic" tab="基础设置">
            <n-form
              ref="basicFormRef"
              :model="basicSettings"
              label-placement="left"
              label-width="120px"
              class="max-w-2xl"
            >
              <n-form-item label="网站名称">
                <n-input v-model:value="basicSettings.siteName" placeholder="请输入网站名称" />
              </n-form-item>
              <n-form-item label="网站标题">
                <n-input v-model:value="basicSettings.siteTitle" placeholder="请输入网站标题" />
              </n-form-item>
              <n-form-item label="网站描述">
                <n-input
                  v-model:value="basicSettings.siteDescription"
                  type="textarea"
                  placeholder="请输入网站描述"
                  :rows="3"
                />
              </n-form-item>
              <n-form-item label="联系邮箱">
                <n-input v-model:value="basicSettings.contactEmail" placeholder="请输入联系邮箱" />
              </n-form-item>
              <n-form-item label="维护模式">
                <n-switch v-model:value="basicSettings.maintenanceMode">
                  <template #checked>开启</template>
                  <template #unchecked>关闭</template>
                </n-switch>
              </n-form-item>
              <n-form-item>
                <n-space>
                  <n-button type="primary" @click="handleSaveBasic">保存设置</n-button>
                  <n-button @click="handleResetBasic">重置</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>

          <n-tab-pane name="ai-model" tab="AI模型设置">
            <n-form
              ref="aiFormRef"
              :model="aiSettings"
              label-placement="left"
              label-width="120px"
              class="max-w-2xl"
            >
              <n-form-item label="模型名称">
                <n-input v-model:value="aiSettings.model_name" placeholder="例如 gpt-4o-mini" />
              </n-form-item>
              <n-form-item label="Base URL">
                <n-input
                  v-model:value="aiSettings.base_url"
                  placeholder="https://api.openai.com（不含 /v1/chat/completions）"
                />
              </n-form-item>
              <n-form-item label="API Key">
                <n-input
                  v-model:value="aiSettings.api_key"
                  type="password"
                  show-password-on="click"
                  placeholder="留空则不修改已保存的密钥"
                />
              </n-form-item>
              <n-form-item>
                <n-space vertical :size="8">
                  <n-text depth="3" style="font-size: 12px">
                    供「AI翻译」等后台能力使用，需为 OpenAI 兼容的 Chat Completions 接口。
                  </n-text>
                  <n-button type="primary" :loading="aiSaving" @click="handleSaveAI">保存 AI 配置</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>

          <n-tab-pane name="cursor-pay" tab="Cursor付款设置">
            <n-form
              ref="cursorPayFormRef"
              :model="cursorPaySettings"
              label-placement="left"
              label-width="120px"
              class="max-w-2xl"
            >
              <n-text depth="3" style="font-size: 13px; display: block; margin-bottom: 16px">
                用于自动提交 Stripe 结账并提取支付宝付款页。
              </n-text>
              <n-form-item label="账单姓名">
                <n-input v-model:value="cursorPaySettings.billing_name" placeholder="例如 AIGuoGuo" />
              </n-form-item>
              <n-form-item label="账单邮编">
                <n-input v-model:value="cursorPaySettings.billing_postal" placeholder="例如 536546" />
              </n-form-item>
              <n-form-item label="账单省份">
                <n-input v-model:value="cursorPaySettings.billing_state" placeholder="例如 Zhejiang" />
              </n-form-item>
              <n-form-item label="账单城市">
                <n-input v-model:value="cursorPaySettings.billing_city" placeholder="例如 Huzhou" />
              </n-form-item>
              <n-form-item label="账单地址">
                <n-input v-model:value="cursorPaySettings.billing_line1" placeholder="例如 清河路177号" />
              </n-form-item>
              <n-form-item label="账单国家">
                <n-input v-model:value="cursorPaySettings.billing_country" placeholder="例如 CN" />
              </n-form-item>
              <n-form-item label="代理协议">
                <n-select
                  v-model:value="cursorPaySettings.proxy_scheme"
                  :options="cursorPayProxySchemeOptions"
                  style="width: 160px"
                />
              </n-form-item>
              <n-form-item label="代理主机">
                <n-input v-model:value="cursorPaySettings.proxy_host" placeholder="例如 hk.stormip.cn:1000" />
              </n-form-item>
              <n-form-item label="代理用户名">
                <n-input v-model:value="cursorPaySettings.proxy_username" placeholder="例如 abc" />
              </n-form-item>
              <n-form-item label="代理密码">
                <n-input
                  v-model:value="cursorPaySettings.proxy_password"
                  type="password"
                  show-password-on="click"
                  :placeholder="cursorPaySettings.proxy_password_configured ? '已配置，留空则不修改' : '留空表示无密码'"
                />
              </n-form-item>
              <n-form-item>
                <n-space vertical :size="8">
                  <n-text depth="3" style="font-size: 12px">
                    代理只用于提取 Stripe / Alipay 付款页。分开填写主机、账号、密码更方便，保存时由后端组装。主机留空则走系统环境代理。
                  </n-text>
                  <n-button type="primary" :loading="cursorPaySaving" @click="handleSaveCursorPay">保存 Cursor 付款设置</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>

          <n-tab-pane name="payment" tab="支付设置">
            <n-form
              ref="paymentFormRef"
              :model="paymentSettings"
              label-placement="left"
              label-width="120px"
              class="max-w-2xl"
            >
              <n-form-item label="支付方式">
                <n-checkbox-group v-model:value="paymentSettings.methods">
                  <n-space>
                    <n-checkbox value="alipay" label="支付宝" />
                    <n-checkbox value="wechat" label="微信支付" />
                    <n-checkbox value="stripe" label="Stripe" />
                  </n-space>
                </n-checkbox-group>
              </n-form-item>
              <n-form-item label="最小充值金额">
                <n-input-number v-model:value="paymentSettings.minAmount" :min="1" :step="1">
                  <template #suffix>元</template>
                </n-input-number>
              </n-form-item>
              <n-form-item label="最大充值金额">
                <n-input-number v-model:value="paymentSettings.maxAmount" :min="1" :step="1">
                  <template #suffix>元</template>
                </n-input-number>
              </n-form-item>
              <n-form-item label="手续费率">
                <n-input-number
                  v-model:value="paymentSettings.feeRate"
                  :min="0"
                  :max="100"
                  :step="0.1"
                >
                  <template #suffix>%</template>
                </n-input-number>
              </n-form-item>
              <n-form-item>
                <n-space>
                  <n-button type="primary" @click="handleSavePayment">保存设置</n-button>
                  <n-button @click="handleResetPayment">重置</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>

          <n-tab-pane name="notification" tab="通知设置">
            <n-form
              ref="notificationFormRef"
              :model="notificationSettings"
              label-placement="left"
              label-width="120px"
              class="max-w-2xl"
            >
              <n-form-item label="邮件通知">
                <n-switch v-model:value="notificationSettings.emailEnabled">
                  <template #checked>开启</template>
                  <template #unchecked>关闭</template>
                </n-switch>
              </n-form-item>
              <n-form-item label="短信通知">
                <n-switch v-model:value="notificationSettings.smsEnabled">
                  <template #checked>开启</template>
                  <template #unchecked>关闭</template>
                </n-switch>
              </n-form-item>
              <n-form-item label="SMTP服务器">
                <n-input
                  v-model:value="notificationSettings.smtpHost"
                  placeholder="smtp.example.com"
                />
              </n-form-item>
              <n-form-item label="SMTP端口">
                <n-input-number v-model:value="notificationSettings.smtpPort" :min="1" :max="65535" />
              </n-form-item>
              <n-form-item label="发件人邮箱">
                <n-input
                  v-model:value="notificationSettings.smtpFrom"
                  placeholder="noreply@example.com"
                />
              </n-form-item>
              <n-form-item>
                <n-space>
                  <n-button type="primary" @click="handleSaveNotification">保存设置</n-button>
                  <n-button @click="handleTestEmail">测试邮件</n-button>
                </n-space>
              </n-form-item>
            </n-form>
          </n-tab-pane>

          <n-tab-pane name="digiseller" tab="Digiseller设置">
            <div class="max-w-2xl">
              <p class="text-sm text-gray-500 mb-4">配置各订阅类型的今日售价，取货时将自动填充对应售价。</p>
              <n-spin :show="digisellerLoading">
                <n-table :bordered="false" :single-line="false" size="small">
                  <thead>
                    <tr>
                      <th style="width: 160px">订阅类型</th>
                      <th>今日售价（USD）</th>
                      <th style="width: 100px">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in digisellerPriceList" :key="item.subscription_type">
                      <td>
                        <n-tag :bordered="false" type="info" size="small">{{ item.subscription_type }}</n-tag>
                      </td>
                      <td>
                        <n-input-number
                          v-model:value="item.price"
                          :min="0"
                          :precision="2"
                          :step="0.5"
                          placeholder="请输入售价"
                          style="width: 160px"
                        >
                          <template #prefix>$</template>
                        </n-input-number>
                      </td>
                      <td>
                        <n-button
                          type="primary"
                          size="small"
                          :loading="item.saving"
                          @click="handleSaveDigisellerPrice(item)"
                        >
                          保存
                        </n-button>
                      </td>
                    </tr>
                  </tbody>
                </n-table>
              </n-spin>
            </div>
          </n-tab-pane>

          <!-- 已登录设备 -->
          <n-tab-pane name="sessions" tab="已登录设备">
            <div class="mb-4 flex items-center justify-between">
              <span class="text-sm text-gray-500">当前所有活跃登录设备</span>
              <n-popconfirm
                @positive-click="handleKickAll"
                positive-text="确认踢出"
                negative-text="取消"
              >
                <template #trigger>
                  <n-button type="error" size="small" :loading="kickAllLoading">
                    一键踢出所有设备（含自己）
                  </n-button>
                </template>
                此操作将踢出所有设备（包括您自己），您将需要重新登录。确认继续？
              </n-popconfirm>
            </div>
            <n-spin :show="sessionsLoading">
              <n-data-table
                :columns="sessionColumns"
                :data="sessions"
                :bordered="false"
                size="small"
                :row-class-name="(row: any) => row.is_current ? 'current-session-row' : ''"
              />
            </n-spin>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import {
  NSpace,
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NSwitch,
  NCheckboxGroup,
  NCheckbox,
  NButton,
  NTable,
  NTag,
  NSpin,
  NText,
  NDataTable,
  NPopconfirm,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { getDigisellerPrices, upsertDigisellerPrice } from '@/api/digiseller'
import { getAdminAISettings, updateAdminAISettings } from '@/api/ai'
import { getAdminCursorPaySettings, updateAdminCursorPaySettings } from '@/api/cursorPay'
import { getAdminSessions, kickSession, kickAllSessions, type AdminSession } from '@/api/admin'

const router = useRouter()

const message = useMessage()

// 基础设置
const aiFormRef = ref()
const aiSaving = ref(false)
const aiSettings = ref({
  model_name: '',
  base_url: '',
  api_key: '',
})

const cursorPayFormRef = ref()
const cursorPaySaving = ref(false)
const cursorPayProxySchemeOptions = [
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'SOCKS5', value: 'socks5' },
]
const cursorPaySettings = ref({
  billing_name: '',
  billing_postal: '',
  billing_state: '',
  billing_city: '',
  billing_line1: '',
  billing_country: '',
  proxy_scheme: 'http',
  proxy_host: '',
  proxy_username: '',
  proxy_password: '',
  proxy_password_configured: false,
})

const basicFormRef = ref()
const basicSettings = ref({
  siteName: 'ChatGPT充值平台',
  siteTitle: 'ChatGPT 自动化充值平台',
  siteDescription: '快速、安全、自动化的 ChatGPT 充值服务',
  contactEmail: 'support@example.com',
  maintenanceMode: false,
})

// 支付设置
const paymentFormRef = ref()
const paymentSettings = ref({
  methods: ['alipay', 'wechat'],
  minAmount: 10,
  maxAmount: 10000,
  feeRate: 2.5,
})

// 通知设置
const notificationFormRef = ref()
const notificationSettings = ref({
  emailEnabled: true,
  smsEnabled: false,
  smtpHost: 'smtp.example.com',
  smtpPort: 587,
  smtpFrom: 'noreply@example.com',
})

// 保存方法
const loadAISettings = async () => {
  try {
    const res = await getAdminAISettings()
    if (res.code === 200 && res.data) {
      aiSettings.value.model_name = res.data.model_name || ''
      aiSettings.value.base_url = res.data.base_url || ''
      aiSettings.value.api_key = ''
    }
  } catch {
    message.error('加载 AI 模型配置失败')
  }
}

const handleSaveAI = async () => {
  aiSaving.value = true
  try {
    const res = await updateAdminAISettings({
      model_name: aiSettings.value.model_name,
      base_url: aiSettings.value.base_url,
      api_key: aiSettings.value.api_key,
    })
    if (res.code === 200) {
      message.success('AI 配置已保存')
      aiSettings.value.api_key = ''
      await loadAISettings()
    } else {
      message.error(res.message || '保存失败')
    }
  } catch {
    message.error('保存失败')
  } finally {
    aiSaving.value = false
  }
}

const loadCursorPaySettings = async () => {
  try {
    const res = await getAdminCursorPaySettings()
    if (res.code === 200 && res.data) {
      cursorPaySettings.value = {
        billing_name: res.data.billing_name || '',
        billing_postal: res.data.billing_postal || '',
        billing_state: res.data.billing_state || '',
        billing_city: res.data.billing_city || '',
        billing_line1: res.data.billing_line1 || '',
        billing_country: res.data.billing_country || '',
        proxy_scheme: res.data.proxy_scheme || 'http',
        proxy_host: res.data.proxy_host || '',
        proxy_username: res.data.proxy_username || '',
        proxy_password: '',
        proxy_password_configured: !!res.data.proxy_password_configured,
      }
    }
  } catch {
    message.error('加载 Cursor 付款配置失败')
  }
}

const handleSaveCursorPay = async () => {
  cursorPaySaving.value = true
  try {
    const res = await updateAdminCursorPaySettings({
      billing_name: cursorPaySettings.value.billing_name,
      billing_postal: cursorPaySettings.value.billing_postal,
      billing_state: cursorPaySettings.value.billing_state,
      billing_city: cursorPaySettings.value.billing_city,
      billing_line1: cursorPaySettings.value.billing_line1,
      billing_country: cursorPaySettings.value.billing_country,
      proxy_scheme: cursorPaySettings.value.proxy_scheme,
      proxy_host: cursorPaySettings.value.proxy_host,
      proxy_username: cursorPaySettings.value.proxy_username,
      proxy_password: cursorPaySettings.value.proxy_password,
    })
    if (res.code === 200) {
      message.success('Cursor 付款设置已保存')
      cursorPaySettings.value.proxy_password = ''
      await loadCursorPaySettings()
    } else {
      message.error(res.message || '保存失败')
    }
  } catch {
    message.error('保存失败')
  } finally {
    cursorPaySaving.value = false
  }
}

const handleSaveBasic = () => {
  message.success('基础设置保存成功')
}

const handleResetBasic = () => {
  message.info('已重置为默认值')
}

const handleSavePayment = () => {
  message.success('支付设置保存成功')
}

const handleResetPayment = () => {
  message.info('已重置为默认值')
}

const handleSaveNotification = () => {
  message.success('通知设置保存成功')
}

const handleTestEmail = () => {
  message.info('正在发送测试邮件...')
  setTimeout(() => {
    message.success('测试邮件发送成功')
  }, 1000)
}

// 所有订阅类型
const allSubscriptionTypes = ['free', 'pro', 'pro_plus', 'pro_x5', 'pro_x20', 'ultra', 'go', 'plus', 'team']

interface DigisellerPriceRow {
  subscription_type: string
  price: number
  saving: boolean
}

const digisellerLoading = ref(false)
const digisellerPriceList = ref<DigisellerPriceRow[]>(
  allSubscriptionTypes.map((t) => ({ subscription_type: t, price: 0, saving: false }))
)

// 加载已保存的价格配置
const loadDigisellerPrices = async () => {
  digisellerLoading.value = true
  try {
    const res = await getDigisellerPrices()
    if (res.code === 200 && res.data) {
      const map = new Map(res.data.map((item) => [item.subscription_type, item.price]))
      digisellerPriceList.value.forEach((row) => {
        if (map.has(row.subscription_type)) {
          row.price = map.get(row.subscription_type)!
        }
      })
    }
  } catch (e) {
    message.error('加载Digiseller价格配置失败')
  } finally {
    digisellerLoading.value = false
  }
}

// 保存单行价格
const handleSaveDigisellerPrice = async (item: DigisellerPriceRow) => {
  item.saving = true
  try {
    const res = await upsertDigisellerPrice({
      subscription_type: item.subscription_type,
      price: item.price,
    })
    if (res.code === 200) {
      message.success(`${item.subscription_type} 售价保存成功`)
    } else {
      message.error(res.message || '保存失败')
    }
  } catch (e) {
    message.error('保存失败')
  } finally {
    item.saving = false
  }
}

// ==================== 已登录设备 ====================
const sessions = ref<AdminSession[]>([])
const sessionsLoading = ref(false)
const kickAllLoading = ref(false)

const loadSessions = async () => {
  sessionsLoading.value = true
  try {
    const res = await getAdminSessions()
    if (res.code === 200) {
      sessions.value = res.data || []
    }
  } catch {
    message.error('加载设备列表失败')
  } finally {
    sessionsLoading.value = false
  }
}

const handleKickOne = async (uuid: string) => {
  try {
    const res = await kickSession(uuid)
    if (res.code === 200) {
      message.success('已踢出该设备')
      await loadSessions()
    } else {
      message.error(res.message || '操作失败')
    }
  } catch {
    message.error('操作失败')
  }
}

const handleKickAll = async () => {
  kickAllLoading.value = true
  try {
    await kickAllSessions()
    // 踢出所有设备（含自己），清除本地 token 并跳转登录页
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
    router.push('/login')
  } catch {
    message.error('操作失败')
  } finally {
    kickAllLoading.value = false
  }
}

const sessionColumns: DataTableColumns<AdminSession> = [
  {
    title: '状态',
    key: 'is_current',
    width: 80,
    render: (row) => row.is_current ? h(NTag, { type: 'success', size: 'small' }, { default: () => '当前' }) : null,
  },
  { title: 'IP 地址', key: 'ip_address', width: 140 },
  {
    title: '登录时间',
    key: 'created_at',
    width: 180,
    render: (row) => new Date(row.created_at * 1000).toLocaleString('zh-CN'),
  },
  {
    title: 'User-Agent',
    key: 'user_agent',
    ellipsis: { tooltip: true },
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(
        NButton,
        { size: 'small', type: 'error', ghost: true, onClick: () => handleKickOne(row.session_uuid) },
        { default: () => '踢出' },
      ),
  },
]

onMounted(() => {
  loadDigisellerPrices()
  loadAISettings()
  loadCursorPaySettings()
  loadSessions()
})
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

