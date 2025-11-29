<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-purple-600 p-4">
    <n-card class="w-full max-w-3xl shadow-2xl">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800 mb-2">系统初始化</h1>
        <p class="text-gray-500">首次使用请完成系统初始化配置</p>
      </div>

      <n-steps :current="currentStep" :status="currentStatus">
        <n-step title="数据库配置" description="配置数据库连接信息" />
        <n-step title="管理员设置" description="创建管理员账号" />
        <n-step title="完成" description="初始化完成" />
      </n-steps>

      <div class="mt-8">
        <!-- 步骤1：数据库配置 -->
        <div v-if="currentStep === 1">
          <n-form
            ref="dbFormRef"
            :model="dbForm"
            :rules="dbRules"
            label-placement="left"
            label-width="120px"
          >
            <n-form-item label="数据库地址" path="db_host">
              <n-input
                v-model:value="dbForm.db_host"
                placeholder="localhost"
              />
            </n-form-item>

            <n-form-item label="数据库端口" path="db_port">
              <n-input
                v-model:value="dbForm.db_port"
                placeholder="3306"
              />
            </n-form-item>

            <n-form-item label="数据库名称" path="db_name">
              <n-input
                v-model:value="dbForm.db_name"
                placeholder="topup_online"
              />
            </n-form-item>

            <n-form-item label="数据库用户名" path="db_user">
              <n-input
                v-model:value="dbForm.db_user"
                placeholder="root"
              />
            </n-form-item>

            <n-form-item label="数据库密码" path="db_password">
              <n-input
                v-model:value="dbForm.db_password"
                type="password"
                show-password-on="click"
                placeholder="请输入数据库密码"
              />
            </n-form-item>

            <n-form-item>
              <n-space>
                <n-button
                  type="info"
                  :loading="testing"
                  @click="testConnection"
                >
                  测试连接
                </n-button>
                <n-button
                  type="primary"
                  :disabled="!dbConnected"
                  @click="nextStep"
                >
                  下一步
                </n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤2：管理员设置 -->
        <div v-if="currentStep === 2">
          <n-form
            ref="adminFormRef"
            :model="adminForm"
            :rules="adminRules"
            label-placement="left"
            label-width="120px"
          >
            <n-form-item label="管理员用户名" path="admin_user">
              <n-input
                v-model:value="adminForm.admin_user"
                placeholder="admin"
              />
            </n-form-item>

            <n-form-item label="管理员密码" path="admin_pass">
              <n-input
                v-model:value="adminForm.admin_pass"
                type="password"
                show-password-on="click"
                placeholder="请输入管理员密码"
              />
            </n-form-item>

            <n-form-item label="确认密码" path="admin_pass_confirm">
              <n-input
                v-model:value="adminForm.admin_pass_confirm"
                type="password"
                show-password-on="click"
                placeholder="请再次输入密码"
              />
            </n-form-item>

            <n-form-item label="管理员邮箱" path="admin_email">
              <n-input
                v-model:value="adminForm.admin_email"
                placeholder="admin@example.com"
              />
            </n-form-item>

            <n-form-item>
              <n-space>
                <n-button @click="prevStep">上一步</n-button>
                <n-button
                  type="primary"
                  :loading="initializing"
                  @click="initialize"
                >
                  开始初始化
                </n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤3：完成 -->
        <div v-if="currentStep === 3" class="text-center py-8">
          <n-result
            status="success"
            title="初始化完成！"
            description="系统已成功初始化，您现在可以登录后台管理系统"
          >
            <template #footer>
              <n-button type="primary" @click="goToLogin">
                前往登录
              </n-button>
            </template>
          </n-result>
        </div>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard,
  NSteps,
  NStep,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NResult,
  useMessage,
} from 'naive-ui'
import { http } from '@/utils/http'

const router = useRouter()
const message = useMessage()

// 当前步骤
const currentStep = ref(1)
const currentStatus = ref<'process' | 'finish' | 'error' | 'wait'>('process')

// 数据库表单
const dbFormRef = ref()
const dbForm = ref({
  db_host: 'localhost',
  db_port: '3306',
  db_name: 'topup_online',
  db_user: 'root',
  db_password: '',
})

const dbRules = {
  db_host: {
    required: true,
    message: '请输入数据库地址',
    trigger: 'blur',
  },
  db_port: {
    required: true,
    message: '请输入数据库端口',
    trigger: 'blur',
  },
  db_name: {
    required: true,
    message: '请输入数据库名称',
    trigger: 'blur',
  },
  db_user: {
    required: true,
    message: '请输入数据库用户名',
    trigger: 'blur',
  },
}

// 管理员表单
const adminFormRef = ref()
const adminForm = ref({
  admin_user: 'admin',
  admin_pass: '',
  admin_pass_confirm: '',
  admin_email: '',
})

const adminRules = {
  admin_user: {
    required: true,
    message: '请输入管理员用户名',
    trigger: 'blur',
  },
  admin_pass: [
    {
      required: true,
      message: '请输入管理员密码',
      trigger: 'blur',
    },
    {
      min: 6,
      message: '密码长度不能少于6位',
      trigger: 'blur',
    },
  ],
  admin_pass_confirm: [
    {
      required: true,
      message: '请确认密码',
      trigger: 'blur',
    },
    {
      validator: (_rule: any, value: string) => {
        return value === adminForm.value.admin_pass
      },
      message: '两次输入的密码不一致',
      trigger: 'blur',
    },
  ],
  admin_email: [
    {
      required: true,
      message: '请输入管理员邮箱',
      trigger: 'blur',
    },
    {
      type: 'email',
      message: '请输入正确的邮箱格式',
      trigger: 'blur',
    },
  ],
}

// 状态
const testing = ref(false)
const dbConnected = ref(false)
const initializing = ref(false)

// 测试数据库连接
const testConnection = () => {
  dbFormRef.value?.validate((errors: any) => {
    if (!errors) {
      testing.value = true
      
      http
        .post('/system/init/test-db', {
          db_host: dbForm.value.db_host,
          db_port: dbForm.value.db_port,
          db_name: dbForm.value.db_name,
          db_user: dbForm.value.db_user,
          db_password: dbForm.value.db_password,
        })
        .then((res: any) => {
          if (res.code === 200) {
            message.success('数据库连接成功！')
            dbConnected.value = true
          } else {
            message.error(res.message || '数据库连接失败')
            dbConnected.value = false
          }
        })
        .catch((error: any) => {
          message.error('测试连接失败: ' + error.message)
          dbConnected.value = false
        })
        .finally(() => {
          testing.value = false
        })
    }
  })
}

// 下一步
const nextStep = () => {
  if (!dbConnected.value) {
    message.warning('请先测试数据库连接')
    return
  }
  currentStep.value = 2
}

// 上一步
const prevStep = () => {
  currentStep.value = 1
}

// MD5加密函数
const md5 = async (str: string) => {
  const encoder = new TextEncoder()
  const data = encoder.encode(str)
  try {
    const hashBuffer = await crypto.subtle.digest('SHA-256', data)
    return Array.from(new Uint8Array(hashBuffer))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('')
  } catch {
    // 简单的备用方案
    return str
  }
}

// 初始化系统
const initialize = () => {
  adminFormRef.value?.validate(async (errors: any) => {
    if (!errors) {
      initializing.value = true
      
      try {
        // 对管理员密码进行MD5加密
        const hashedPassword = await md5(adminForm.value.admin_pass)
        
        const initData = {
          ...dbForm.value,
          admin_user: adminForm.value.admin_user,
          admin_pass: hashedPassword,
          admin_email: adminForm.value.admin_email,
        }
        
        const res: any = await http.post('/system/init', initData)
        
        if (res.code === 200) {
          message.success('系统初始化成功！')
          currentStep.value = 3
          currentStatus.value = 'finish'
        } else {
          message.error(res.message || '初始化失败')
          currentStatus.value = 'error'
        }
      } catch (error: any) {
        message.error('初始化失败: ' + error.message)
        currentStatus.value = 'error'
      } finally {
        initializing.value = false
      }
    }
  })
}

// 前往登录
const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
:deep(.n-card) {
  border-radius: 16px;
}

:deep(.n-steps) {
  margin-bottom: 32px;
}
</style>

