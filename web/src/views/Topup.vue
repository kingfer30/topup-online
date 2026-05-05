<template>
  <div>
    <!-- Header -->
    <header class="bg-white p-4 sticky top-0 z-40 shadow-xl border-b border-black/20">
      <div class="max-w-7xl mx-auto flex justify-between items-center flex-col sm:flex-row gap-3 sm:gap-0">
        <div>
          <n-button
            @click="router.push('/')"
            class="nav-link-active"
            :bordered="false"
          >
            {{ t.nav_home }}
          </n-button>
        </div>
        <nav class="flex gap-2 items-center justify-center">
          <a class="nav-link cursor-pointer" @click="router.push('/topup')">
            {{ t.hero_btn }}
          </a>
          <a class="nav-link cursor-pointer" @click="router.push('/#features')">
            {{ t.nav_features }}
          </a>
          <a class="nav-link cursor-pointer" @click="router.push('/#steps')">
            {{ t.nav_steps }}
          </a>
          <a class="nav-link cursor-pointer" @click="router.push('/#faq')">
            {{ t.nav_faq }}
          </a>
          <n-dropdown
            :options="langOptions"
            @select="switchLang"
            trigger="click"
          >
            <n-button size="small" class="ml-2 lang-dropdown-btn">
              <span v-if="currentLang === 'zh'" class="flex items-center gap-1">
                <img :src="cnFlag" alt="China Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                中文
              </span>
              <span v-else-if="currentLang === 'ru'" class="flex items-center gap-1">
                <img :src="ruFlag" alt="Russia Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                Русский
              </span>
              <span v-else class="flex items-center gap-1">
                <img :src="usFlag" alt="US Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                English
              </span>
            </n-button>
          </n-dropdown>
        </nav>
      </div>
    </header>

    <!-- Main Content -->
    <div class="min-h-screen bg-gray-50 py-10 px-4">
      <div class="max-w-2xl mx-auto">
        <!-- 页面标题 -->
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-gray-800">{{ t.topup_page_title }}</h1>
          <p class="text-gray-500 mt-2">{{ t.topup_page_subtitle }}</p>
        </div>

        <!-- 步骤条 -->
        <n-steps :current="currentStep" class="mb-8">
          <n-step :title="t.topup_step1_name" :description="t.topup_step1_subdesc" />
          <n-step :title="t.topup_step2_name" :description="t.topup_step2_subdesc" />
          <n-step :title="t.topup_step3_name" :description="t.topup_step3_subdesc" />
          <n-step :title="t.topup_step4_name" :description="t.topup_step4_subdesc" />
        </n-steps>

        <!-- 步骤内容卡片 -->
        <n-card class="shadow-md">
          <!-- Step 1: 验证会话令牌 -->
          <div v-if="currentStep === 1">
            <h2 class="text-lg font-semibold text-gray-700 mb-4">{{ t.topup_s1_heading }}</h2>
            <n-form
              ref="step1FormRef"
              :model="step1Form"
              :rules="step1Rules"
              label-placement="top"
            >
              <n-form-item :label="t.topup_email_label" path="userEmail">
                <n-input
                  v-model:value="step1Form.userEmail"
                  :placeholder="t.topup_email_ph"
                  clearable
                />
              </n-form-item>
              <n-form-item label="Session Token" path="userGptToken">
                <n-input
                  v-model:value="step1Form.userGptToken"
                  type="textarea"
                  :rows="4"
                  :placeholder="t.topup_token_ph"
                  clearable
                />
              </n-form-item>
              <n-form-item :label="t.topup_auth_label" path="fullAuthData">
                <n-input
                  v-model:value="step1Form.fullAuthData"
                  type="textarea"
                  :rows="3"
                  :placeholder="t.topup_auth_ph"
                  clearable
                />
              </n-form-item>
            </n-form>
            <div class="flex justify-end mt-4">
              <n-button type="primary" @click="handleStep1Next">
                {{ t.topup_next }}
              </n-button>
            </div>
          </div>

          <!-- Step 2: 验证CDK -->
          <div v-if="currentStep === 2">
            <h2 class="text-lg font-semibold text-gray-700 mb-4">{{ t.topup_s2_heading }}</h2>
            <n-form
              ref="step2FormRef"
              :model="step2Form"
              :rules="step2Rules"
              label-placement="top"
            >
              <n-form-item :label="t.topup_cdk_label" path="cdkKey">
                <n-input
                  v-model:value="step2Form.cdkKey"
                  :placeholder="t.topup_cdk_ph"
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
              {{ t.topup_cdk_ok }}
              <span v-if="cdkInfo.expire_time" class="ml-2 text-gray-500 text-sm">
                {{ t.topup_expire_prefix }}{{ formatTime(cdkInfo.expire_time) }}
              </span>
              <span v-else class="ml-2 text-gray-500 text-sm">{{ t.topup_permanent }}</span>
            </n-alert>

            <div class="flex justify-between mt-4">
              <n-button @click="currentStep = 1">{{ t.topup_prev }}</n-button>
              <div class="flex gap-2">
                <n-button
                  v-if="!cdkVerified"
                  type="primary"
                  :loading="verifyingCdk"
                  @click="handleVerifyCdk"
                >
                  {{ t.topup_verify_cdk }}
                </n-button>
                <n-button
                  v-if="cdkVerified"
                  type="primary"
                  @click="currentStep = 3"
                >
                  {{ t.topup_next }}
                </n-button>
              </div>
            </div>
          </div>

          <!-- Step 3: 确认充值 -->
          <div v-if="currentStep === 3">
            <h2 class="text-lg font-semibold text-gray-700 mb-6">{{ t.topup_s3_heading }}</h2>
            <div class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="flex justify-between">
                <span class="text-gray-500">{{ t.topup_account_label }}</span>
                <span class="font-medium">{{ step1Form.userEmail }}</span>
              </div>
              <n-divider class="my-2" />
              <div class="flex justify-between">
                <span class="text-gray-500">{{ t.topup_cdk_label }}</span>
                <span class="font-medium font-mono text-sm">{{ maskCdk(step2Form.cdkKey) }}</span>
              </div>
              <n-divider class="my-2" />
              <div class="flex justify-between">
                <span class="text-gray-500">{{ t.topup_cdk_expire_label }}</span>
                <span class="font-medium">
                  {{ cdkInfo.expire_time ? formatTime(cdkInfo.expire_time) : t.topup_permanent }}
                </span>
              </div>
            </div>
            <n-alert type="warning" class="mt-4">
              {{ t.topup_confirm_warning }}
            </n-alert>
            <div class="flex justify-between mt-6">
              <n-button @click="currentStep = 2">{{ t.topup_prev }}</n-button>
              <n-button type="primary" @click="handleConfirmTopup">
                {{ t.topup_confirm }}
              </n-button>
            </div>
          </div>

          <!-- Step 4: 充值进行中 -->
          <div v-if="currentStep === 4">
            <h2 class="text-lg font-semibold text-gray-700 mb-6">{{ t.topup_s4_heading }}</h2>

            <!-- 处理中 -->
            <div v-if="taskStatus === 0 || taskStatus === 1" class="text-center py-10">
              <n-spin size="large" />
              <p class="mt-4 text-gray-600">{{ t.topup_processing }}</p>
              <p class="text-gray-400 text-sm mt-1">{{ t.topup_processing_time }}</p>
            </div>

            <!-- 充值成功 -->
            <div v-else-if="taskStatus === 2" class="text-center py-10">
              <div class="text-6xl mb-4">🎉</div>
              <p class="text-2xl font-bold text-green-600">{{ t.topup_success }}</p>
              <p v-if="taskMessage" class="text-gray-500 mt-2 text-sm">{{ taskMessage }}</p>
              <n-button type="primary" class="mt-6" @click="handleReset">
                {{ t.topup_recharge_again }}
              </n-button>
            </div>

            <!-- 充值失败 -->
            <div v-else-if="taskStatus === 3" class="text-center py-10">
              <div class="text-6xl mb-4">❌</div>
              <p class="text-2xl font-bold text-red-600">{{ t.topup_failed }}</p>
              <p v-if="taskMessage" class="text-gray-500 mt-2 text-sm max-w-sm mx-auto">{{ taskMessage }}</p>
              <div class="flex gap-3 justify-center mt-6">
                <n-button @click="handleReset">{{ t.topup_restart }}</n-button>
                <n-button type="primary" @click="currentStep = 3">{{ t.topup_retry }}</n-button>
              </div>
            </div>
          </div>
        </n-card>
      </div>
    </div>

    <!-- Footer -->
    <footer class="bg-gray-100 text-center text-sm p-6 text-gray-600 flex flex-col sm:flex-row justify-center items-center gap-3">
      <span>{{ t.footer_copyright }}</span>
      <span class="hidden sm:inline-block mx-2">|</span>
      <n-button text @click="showTechModal = true" class="underline hover:text-black">
        {{ t.footer_tech }}
      </n-button>
      <span class="hidden sm:inline-block mx-2">|</span>
      <n-button text @click="showPrivacyModal = true" class="underline hover:text-black">
        {{ t.footer_privacy }}
      </n-button>
    </footer>

    <!-- Tech Modal -->
    <n-modal v-model:show="showTechModal">
      <n-card style="width: 600px; max-width: 90vw;" title="Tech Assurance" :bordered="false" size="huge" role="dialog" aria-modal="true">
        <div class="space-y-4">
          <p>Our platform provides comprehensive technical assurance:</p>
          <ul class="list-disc pl-6 space-y-2">
            <li>24/7 automated system monitoring</li>
            <li>99.9% uptime guarantee</li>
            <li>Secure token handling with encryption</li>
            <li>Real-time transaction processing</li>
            <li>Professional technical support</li>
          </ul>
        </div>
      </n-card>
    </n-modal>

    <!-- Privacy Modal -->
    <n-modal v-model:show="showPrivacyModal">
      <n-card style="width: 600px; max-width: 90vw;" title="Privacy Policy" :bordered="false" size="huge" role="dialog" aria-modal="true">
        <div class="space-y-4">
          <p>We are committed to protecting your privacy:</p>
          <ul class="list-disc pl-6 space-y-2">
            <li>We only collect necessary token information</li>
            <li>No personal account credentials are stored</li>
            <li>All data is encrypted in transit and at rest</li>
            <li>Tokens are securely deleted after processing</li>
            <li>We do not share data with third parties</li>
          </ul>
        </div>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, h } from 'vue'
import { useRouter } from 'vue-router'
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
  NDropdown,
  NModal,
  type FormInst,
  type FormRules,
  useMessage,
} from 'naive-ui'
import '@/styles/custom.css'
import { verifyCdk, startTopup, getTopupTaskStatus } from '@/api/topup'
import type { VerifyCdkResponse } from '@/api/topup'
import enLang from '@/lang/en'
import zhLang from '@/lang/zh'
import ruLang from '@/lang/ru'

const router = useRouter()

const message = useMessage()

// Language management
const cnFlag = '/flags/CN.svg'
const usFlag = '/flags/US.svg'
const ruFlag = '/flags/RU.svg'

const langDict: Record<string, any> = { en: enLang, zh: zhLang, ru: ruLang }
const currentLang = ref('zh')

const detectLang = () => {
  const lang = navigator.language || (navigator as any).userLanguage
  if (lang.startsWith('zh')) return 'zh'
  if (lang.startsWith('ru')) return 'ru'
  return 'en'
}

const switchLang = (lang: string) => {
  currentLang.value = lang
  localStorage.setItem('lang', lang)
  document.title = langDict[lang].page_title
}

const renderIcon = (flagSvg: string) => {
  return () => h('img', { src: flagSvg, style: 'width: 20px; height: 15px; border-radius: 2px;', alt: 'Flag' })
}

const langOptions = [
  { label: 'English', key: 'en', icon: renderIcon(usFlag) },
  { label: '中文', key: 'zh', icon: renderIcon(cnFlag) },
  { label: 'Русский', key: 'ru', icon: renderIcon(ruFlag) },
]

const t = computed(() => langDict[currentLang.value])

const showTechModal = ref(false)
const showPrivacyModal = ref(false)

onMounted(() => {
  const savedLang = localStorage.getItem('lang') || detectLang()
  currentLang.value = savedLang
  document.title = langDict[savedLang].page_title
})

const currentStep = ref(1)

// Step1 表单
const step1FormRef = ref<FormInst | null>(null)
const step1Form = ref({
  userEmail: '',
  userGptToken: '',
  fullAuthData: '',
})
const step1Rules = computed<FormRules>(() => ({
  userEmail: [
    { required: true, message: t.value.topup_email_required, trigger: 'blur' },
    { type: 'email', message: t.value.topup_email_invalid, trigger: 'blur' },
  ],
  userGptToken: [
    { required: true, message: t.value.topup_token_required, trigger: 'blur' },
    { min: 20, message: t.value.topup_token_invalid, trigger: 'blur' },
  ],
}))

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
const step2Rules = computed<FormRules>(() => ({
  cdkKey: [{ required: true, message: t.value.topup_cdk_required, trigger: 'blur' }],
}))
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
      message.error(e?.message || t.value.topup_cdk_error)
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
    taskMessage.value = e?.message || t.value.topup_start_error
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
  if (!ms) return t.value.topup_permanent
  return new Date(ms).toLocaleDateString(t.value.locale, {
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
