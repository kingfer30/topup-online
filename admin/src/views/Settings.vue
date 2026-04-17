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
        </n-tabs>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NSpace,
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NCheckboxGroup,
  NCheckbox,
  NButton,
  NTable,
  NTag,
  NSpin,
  NText,
  useMessage,
} from 'naive-ui'
import { getDigisellerPrices, upsertDigisellerPrice } from '@/api/digiseller'
import { getAdminAISettings, updateAdminAISettings } from '@/api/ai'

const message = useMessage()

// 基础设置
const aiFormRef = ref()
const aiSaving = ref(false)
const aiSettings = ref({
  model_name: '',
  base_url: '',
  api_key: '',
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
const allSubscriptionTypes = ['pro', 'pro_plus', 'ultra', 'go', 'plus', 'team']

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

onMounted(() => {
  loadDigisellerPrices()
  loadAISettings()
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

