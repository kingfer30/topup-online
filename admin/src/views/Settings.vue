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
  NInputNumber,
  NSwitch,
  NCheckboxGroup,
  NCheckbox,
  NButton,
  useMessage,
} from 'naive-ui'

const message = useMessage()

// 基础设置
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

