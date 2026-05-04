<template>
  <div class="min-h-screen bg-gray-50 py-10 px-4">
    <div class="max-w-2xl mx-auto">
      <!-- 页面标题 -->
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">GPT Plus 充值</h1>
        <p class="text-gray-500 mt-2">请按步骤完成充值流程</p>
      </div>

      <!-- 步骤条 -->
      <n-steps :current="currentStep" class="mb-8">
        <n-step title="验证令牌" description="填写 GPT 账号信息" />
        <n-step title="验证CDK" description="输入充值码" />
        <n-step title="确认充值" description="核对信息" />
        <n-step title="充值进行中" description="等待完成" />
      </n-steps>

      <!-- 步骤内容卡片 -->
      <n-card class="shadow-md">
        <!-- Step 1: 验证会话令牌 -->
        <div v-if="currentStep === 1">
          <h2 class="text-lg font-semibold text-gray-700 mb-4">Step 1 — 验证会话令牌</h2>
          <n-form
            ref="step1FormRef"
            :model="step1Form"
            :rules="step1Rules"
            label-placement="top"
          >
            <n-form-item label="GPT 账号邮箱" path="userEmail">
              <n-input
                v-model:value="step1Form.userEmail"
                placeholder="请输入 ChatGPT 账号邮箱"
                clearable
              />
            </n-form-item>
            <n-form-item label="Session Token" path="userGptToken">
              <n-input
                v-model:value="step1Form.userGptToken"
                type="textarea"
                :rows="4"
                placeholder="请粘贴 __Secure-next-auth.session-token Cookie 值"
                clearable
              />
            </n-form-item>
            <n-form-item label="完整认证数据（可选）" path="fullAuthData">
              <n-input
                v-model:value="step1Form.fullAuthData"
                type="textarea"
                :rows="3"
                placeholder="粘贴完整认证 JSON（选填，提高成功率）"
                clearable
              />
            </n-form-item>
          </n-form>
          <div class="flex justify-end mt-4">
            <n-button type="primary" @click="handleStep1Next">
              下一步
            </n-button>
          </div>
        </div>

        <!-- Step 2: 验证CDK -->
        <div v-if="currentStep === 2">
          <h2 class="text-lg font-semibold text-gray-700 mb-4">Step 2 — 验证CDK</h2>
          <n-form
            ref="step2FormRef"
            :model="step2Form"
            :rules="step2Rules"
            label-placement="top"
          >
            <n-form-item label="充值码（CDK）" path="cdkKey">
              <n-input
                v-model:value="step2Form.cdkKey"
                placeholder="请输入充值码"
                clearable
                :disabled="cdkVerified"
              />
            </n-form-item>
          </n-form>

          <!-- CDK 验证结果 -->
          <n-alert
            v-if="cdkVerified"
            type="success"
            class="mb-4"
          >
            <template #icon>
              <span>✓</span>
            </template>
            CDK 验证通过！
            <span v-if="cdkInfo.expire_time" class="ml-2 text-gray-500 text-sm">
              有效期至：{{ formatTime(cdkInfo.expire_time) }}
            </span>
            <span v-else class="ml-2 text-gray-500 text-sm">永久有效</span>
          </n-alert>

          <div class="flex justify-between mt-4">
            <n-button @click="currentStep = 1">上一步</n-button>
            <div class="flex gap-2">
              <n-button
                v-if="!cdkVerified"
                type="primary"
                :loading="verifyingCdk"
                @click="handleVerifyCdk"
              >
                验证CDK
              </n-button>
              <n-button
                v-if="cdkVerified"
                type="primary"
                @click="currentStep = 3"
              >
                下一步
              </n-button>
            </div>
          </div>
        </div>

        <!-- Step 3: 确认充值 -->
        <div v-if="currentStep === 3">
          <h2 class="text-lg font-semibold text-gray-700 mb-6">Step 3 — 确认充值信息</h2>
          <div class="bg-gray-50 rounded-lg p-4 space-y-3">
            <div class="flex justify-between">
              <span class="text-gray-500">充值账号</span>
              <span class="font-medium">{{ step1Form.userEmail }}</span>
            </div>
            <n-divider class="my-2" />
            <div class="flex justify-between">
              <span class="text-gray-500">充值码（CDK）</span>
              <span class="font-medium font-mono text-sm">{{ maskCdk(step2Form.cdkKey) }}</span>
            </div>
            <n-divider class="my-2" />
            <div class="flex justify-between">
              <span class="text-gray-500">CDK 有效期</span>
              <span class="font-medium">
                {{ cdkInfo.expire_time ? formatTime(cdkInfo.expire_time) : '永久有效' }}
              </span>
            </div>
          </div>
          <n-alert type="warning" class="mt-4">
            请确认以上信息无误。充值开始后将无法撤销，CDK 将被锁定使用。
          </n-alert>
          <div class="flex justify-between mt-6">
            <n-button @click="currentStep = 2">上一步</n-button>
            <n-button type="primary" @click="handleConfirmTopup">
              确认充值
            </n-button>
          </div>
        </div>

        <!-- Step 4: 充值进行中 -->
        <div v-if="currentStep === 4">
          <h2 class="text-lg font-semibold text-gray-700 mb-6">Step 4 — 充值进行中</h2>

          <!-- 处理中 -->
          <div v-if="taskStatus === 0 || taskStatus === 1" class="text-center py-10">
            <n-spin size="large" />
            <p class="mt-4 text-gray-600">正在处理充值，请耐心等待…</p>
            <p class="text-gray-400 text-sm mt-1">预计需要 1~3 分钟</p>
          </div>

          <!-- 充值成功 -->
          <div v-else-if="taskStatus === 2" class="text-center py-10">
            <div class="text-6xl mb-4">🎉</div>
            <p class="text-2xl font-bold text-green-600">充值成功！</p>
            <p v-if="taskMessage" class="text-gray-500 mt-2 text-sm">{{ taskMessage }}</p>
            <n-button type="primary" class="mt-6" @click="handleReset">
              重新充值
            </n-button>
          </div>

          <!-- 充值失败 -->
          <div v-else-if="taskStatus === 3" class="text-center py-10">
            <div class="text-6xl mb-4">❌</div>
            <p class="text-2xl font-bold text-red-600">充值失败</p>
            <p v-if="taskMessage" class="text-gray-500 mt-2 text-sm max-w-sm mx-auto">{{ taskMessage }}</p>
            <div class="flex gap-3 justify-center mt-6">
              <n-button @click="handleReset">重新开始</n-button>
              <n-button type="primary" @click="currentStep = 3">再次尝试</n-button>
            </div>
          </div>
        </div>
      </n-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import {
  NSteps,
  NStep,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NAlert,
  NDivider,
  NSpin,
  type FormInst,
  type FormRules,
  useMessage,
} from 'naive-ui'
import { verifyCdk, startTopup, getTopupTaskStatus } from '@/api/topup'
import type { VerifyCdkResponse } from '@/api/topup'

const message = useMessage()

const currentStep = ref(1)

// Step1 表单
const step1FormRef = ref<FormInst | null>(null)
const step1Form = ref({
  userEmail: '',
  userGptToken: '',
  fullAuthData: '',
})
const step1Rules: FormRules = {
  userEmail: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  userGptToken: [
    { required: true, message: '请输入 Session Token', trigger: 'blur' },
    { min: 20, message: 'Token 格式不正确', trigger: 'blur' },
  ],
}

const handleStep1Next = () => {
  step1FormRef.value?.validate((errors) => {
    if (!errors) {
      currentStep.value = 2
    }
  })
}

// Step2 CDK 验证
const step2FormRef = ref<FormInst | null>(null)
const step2Form = ref({ cdkKey: '' })
const step2Rules: FormRules = {
  cdkKey: [{ required: true, message: '请输入充值码', trigger: 'blur' }],
}
const verifyingCdk = ref(false)
const cdkVerified = ref(false)
const cdkInfo = ref<VerifyCdkResponse>({ valid: false, cdk_id: 0, expire_time: null })

const handleVerifyCdk = async () => {
  step2FormRef.value?.validate(async (errors) => {
    if (errors) return
    verifyingCdk.value = true
    try {
      const res = await verifyCdk(step2Form.value.cdkKey)
      cdkInfo.value = res.data
      cdkVerified.value = true
    } catch (e: any) {
      message.error(e?.message || 'CDK 验证失败，请检查是否已使用或输入有误')
    } finally {
      verifyingCdk.value = false
    }
  })
}

// Step3 确认充值
const handleConfirmTopup = () => {
  currentStep.value = 4
  doStartTopup()
}

// Step4 充值流程
const taskId = ref<number | null>(null)
const taskStatus = ref<number>(1)
const taskMessage = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const doStartTopup = async () => {
  try {
    const res = await startTopup({
      cdk_key: step2Form.value.cdkKey,
      user_email: step1Form.value.userEmail,
      user_gpt_token: step1Form.value.userGptToken,
      full_auth_data: step1Form.value.fullAuthData,
    })
    taskId.value = res.data.task_id
    taskStatus.value = 1
    startPolling()
  } catch (e: any) {
    taskStatus.value = 3
    taskMessage.value = e?.message || '充值请求失败，请稍后重试'
  }
}

const startPolling = () => {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(async () => {
    if (!taskId.value) return
    try {
      const res = await getTopupTaskStatus(taskId.value)
      taskStatus.value = res.data.status
      taskMessage.value = res.data.message
      if (taskStatus.value === 2 || taskStatus.value === 3) {
        stopPolling()
      }
    } catch (_) {
      // 忽略轮询错误，继续轮询
    }
  }, 3000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onUnmounted(() => {
  stopPolling()
})

// 重置所有状态，重新开始
const handleReset = () => {
  stopPolling()
  currentStep.value = 1
  step1Form.value = { userEmail: '', userGptToken: '', fullAuthData: '' }
  step2Form.value = { cdkKey: '' }
  cdkVerified.value = false
  cdkInfo.value = { valid: false, cdk_id: 0, expire_time: null }
  taskId.value = null
  taskStatus.value = 1
  taskMessage.value = ''
}

// 格式化时间戳（毫秒）
const formatTime = (ms: number | null) => {
  if (!ms) return '永久有效'
  return new Date(ms).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

// 部分遮掩 CDK
const maskCdk = (key: string) => {
  if (key.length <= 8) return key
  return key.slice(0, 4) + '****' + key.slice(-4)
}
</script>
