<template>
  <div class="login-container">
    <!-- 渐变背景 -->
    <div class="login-bg"></div>
    
    <!-- 登录卡片 -->
    <div class="login-card-wrapper">
      <n-card class="login-card" :bordered="false">
        <div class="text-center mb-8">
          <!-- Apple 风格 Logo -->
          <div class="mx-auto w-16 h-16 rounded-[20px] bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center mb-5 shadow-lg">
            <span class="text-white text-3xl font-bold">T</span>
          </div>
          <h1 class="text-[28px] font-bold text-gray-800 tracking-tight mb-1">后台管理系统</h1>
          <p class="text-[15px] text-gray-400 font-normal">请登录您的管理员账号</p>
        </div>

        <n-form ref="formRef" :model="formValue" :rules="rules" size="large">
          <n-form-item path="username" :show-label="false">
            <n-input
              v-model:value="formValue.username"
              placeholder="用户名"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <n-icon class="input-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                    <path fill-rule="evenodd" d="M7.5 6a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM3.751 20.105a8.25 8.25 0 0116.498 0 .75.75 0 01-.437.695A18.683 18.683 0 0112 22.5c-2.786 0-5.433-.608-7.812-1.7a.75.75 0 01-.437-.695z" clip-rule="evenodd" />
                  </svg>
                </n-icon>
              </template>
            </n-input>
          </n-form-item>

          <n-form-item path="password" :show-label="false">
            <n-input
              v-model:value="formValue.password"
              type="password"
              show-password-on="click"
              placeholder="密码"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <n-icon class="input-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                    <path fill-rule="evenodd" d="M12 1.5a5.25 5.25 0 00-5.25 5.25v3a3 3 0 00-3 3v6.75a3 3 0 003 3h10.5a3 3 0 003-3v-6.75a3 3 0 00-3-3v-3c0-2.9-2.35-5.25-5.25-5.25zm3.75 8.25v-3a3.75 3.75 0 10-7.5 0v3h7.5z" clip-rule="evenodd" />
                  </svg>
                </n-icon>
              </template>
            </n-input>
          </n-form-item>

          <n-form-item :show-label="false">
            <n-checkbox v-model:checked="formValue.remember">
              记住我
            </n-checkbox>
          </n-form-item>

          <n-form-item :show-label="false">
            <n-button
              type="primary"
              block
              size="large"
              :loading="loading"
              @click="handleLogin"
              class="login-btn"
            >
              登录
            </n-button>
          </n-form-item>
        </n-form>
      </n-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NCheckbox,
  NButton,
  NIcon,
  useMessage,
} from 'naive-ui'
import { adminLogin } from '@/api/admin'
import CryptoJS from 'crypto-js'

const router = useRouter()
const message = useMessage()

const formRef = ref()
const loading = ref(false)

const formValue = ref({
  username: '',
  password: '',
  remember: false,
})

// 页面加载时，读取保存的用户名
onMounted(() => {
  const savedUsername = localStorage.getItem('remember_username')
  if (savedUsername) {
    formValue.value.username = savedUsername
    formValue.value.remember = true
  }
})

const rules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur',
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur',
  },
}

// MD5加密函数
const md5 = (str: string) => {
  return CryptoJS.MD5(str).toString()
}

const handleLogin = () => {
  formRef.value?.validate(async (errors: any) => {
    if (!errors) {
      loading.value = true
      
      try {
        // 对密码进行MD5加密
        const hashedPassword = md5(formValue.value.password)
        
        // 发送登录请求到后端
        const res: any = await adminLogin({
          username: formValue.value.username,
          password: hashedPassword,
        })
        
        if (res.code === 200) {
          message.success('登录成功')
          
          // 保存token和管理员信息
          localStorage.setItem('admin_token', res.data.token)
          localStorage.setItem('admin_info', JSON.stringify(res.data.admin))
          
          // 处理"记住我"功能
          if (formValue.value.remember) {
            // 如果勾选了记住我，保存用户名
            localStorage.setItem('remember_username', formValue.value.username)
          } else {
            // 如果没有勾选，清除保存的用户名
            localStorage.removeItem('remember_username')
          }
          
          router.push('/admin/dashboard')
        } else {
          message.error(res.message || '登录失败')
        }
      } catch (error: any) {
        message.error(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
/* Apple 风格登录容器 */
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

/* 渐变背景 - Apple 风格 */
.login-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  z-index: 0;
}

.login-bg::before {
  content: '';
  position: absolute;
  inset: 0;
  background: 
    radial-gradient(ellipse at 20% 50%, rgba(255, 255, 255, 0.15) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 20%, rgba(255, 255, 255, 0.1) 0%, transparent 50%),
    radial-gradient(ellipse at 40% 80%, rgba(255, 255, 255, 0.08) 0%, transparent 40%);
}

/* 登录卡片包装 */
.login-card-wrapper {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 420px;
  padding: 20px;
}

/* Apple 毛玻璃卡片 */
.login-card {
  background: rgba(255, 255, 255, 0.85) !important;
  backdrop-filter: saturate(180%) blur(20px) !important;
  -webkit-backdrop-filter: saturate(180%) blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.15),
    0 0 0 1px rgba(255, 255, 255, 0.4) inset !important;
  padding: 12px !important;
}

/* 输入框样式 */
:deep(.n-input) {
  border-radius: 12px !important;
  background: rgba(0, 0, 0, 0.04) !important;
  border: 1px solid rgba(0, 0, 0, 0.06) !important;
  transition: all 0.25s ease !important;
  height: 48px !important;
}

:deep(.n-input:focus-within) {
  background: rgba(255, 255, 255, 1) !important;
  border-color: #007AFF !important;
  box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.15) !important;
}

:deep(.n-input__input-el) {
  font-size: 16px !important;
  height: 48px !important;
  caret-color: #007AFF !important;
}

:deep(.n-input__input-el::placeholder) {
  color: rgba(0, 0, 0, 0.3) !important;
}

.input-icon {
  color: rgba(0, 0, 0, 0.3) !important;
  font-size: 18px !important;
}

/* 登录按钮 */
.login-btn {
  height: 48px !important;
  border-radius: 12px !important;
  font-size: 16px !important;
  font-weight: 600 !important;
  letter-spacing: 0.5px;
  box-shadow: 0 4px 12px rgba(0, 122, 255, 0.3) !important;
  transition: all 0.25s ease !important;
}

.login-btn:hover {
  box-shadow: 0 6px 20px rgba(0, 122, 255, 0.4) !important;
  transform: translateY(-1px);
}

.login-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(0, 122, 255, 0.3) !important;
}

/* 复选框样式 */
:deep(.n-checkbox__label) {
  color: rgba(0, 0, 0, 0.6) !important;
  font-size: 14px;
}

/* 表单项间距 */
:deep(.n-form-item) {
  margin-bottom: 4px;
}

/* 响应式 */
@media (max-width: 768px) {
  .login-card-wrapper {
    max-width: 100%;
    padding: 16px;
  }
}
</style>
