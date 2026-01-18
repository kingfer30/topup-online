<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NCard, 
  NTabs, 
  NTabPane, 
  NForm, 
  NFormItem, 
  NInput, 
  NButton, 
  NCheckbox,
  NDropdown,
  useMessage 
} from 'naive-ui'
import { login, register } from '@/api/user'
import { useUserStore } from '@/stores/user'
import type { LoginRequest } from '@/types'
import { encryptPassword } from '@/utils/crypto'

// Import language files
import enLang from '@/lang/en'
import zhLang from '@/lang/zh'
import ruLang from '@/lang/ru'

// Language management
const currentLang = ref('zh')

// Use flag SVGs from public directory
const cnFlag = '/flags/CN.svg'
const usFlag = '/flags/US.svg'
const ruFlag = '/flags/RU.svg'

// Language dictionary
const langDict: Record<string, any> = {
  en: enLang,
  zh: zhLang,
  ru: ruLang
}

// Helper function to render flag icons
const renderIcon = (flagSvg: string) => {
  return () => h('img', { 
    src: flagSvg, 
    style: 'width: 20px; height: 15px; border-radius: 2px;',
    alt: 'Flag'
  })
}

// Language options for dropdown
const langOptions = [
  {
    label: 'English',
    key: 'en',
    icon: renderIcon(usFlag)
  },
  {
    label: '中文',
    key: 'zh',
    icon: renderIcon(cnFlag)
  },
  {
    label: 'Русский',
    key: 'ru',
    icon: renderIcon(ruFlag)
  }
]

// Language detection and management
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

// Computed translation object
const t = computed(() => langDict[currentLang.value])

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

// 当前激活的标签页
const activeTab = ref<'login' | 'register'>('login')

// 登录表单
const loginFormRef = ref()
const loginForm = reactive({
  username: '',
  password: '',
  remember: false
})

// 注册表单
const registerFormRef = ref()
const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// 加载状态
const loginLoading = ref(false)
const registerLoading = ref(false)

// 表单验证规则
const loginRules = computed(() => ({
  username: {
    required: true,
    message: t.value.validate_username_required,
    trigger: 'blur'
  },
  password: {
    required: true,
    message: t.value.validate_password_required,
    trigger: 'blur'
  }
}))

const registerRules = computed(() => ({
  username: {
    required: true,
    message: t.value.validate_username_required,
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!value) {
        return new Error(t.value.validate_username_required)
      }
      if (value.length < 3) {
        return new Error(t.value.validate_username_min)
      }
      return true
    }
  },
  email: {
    required: true,
    message: t.value.validate_email_required,
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!value) {
        return new Error(t.value.validate_email_required)
      }
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(value)) {
        return new Error(t.value.validate_email_invalid)
      }
      return true
    }
  },
  password: {
    required: true,
    message: t.value.validate_password_required,
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!value) {
        return new Error(t.value.validate_password_required)
      }
      if (value.length < 6) {
        return new Error(t.value.validate_password_min)
      }
      return true
    }
  },
  confirmPassword: {
    required: true,
    message: t.value.validate_confirm_required,
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!value) {
        return new Error(t.value.validate_confirm_required)
      }
      if (value !== registerForm.password) {
        return new Error(t.value.validate_confirm_mismatch)
      }
      return true
    }
  }
}))

// 处理登录
const handleLogin = async () => {
  try {
    await loginFormRef.value?.validate()
  } catch (error: any) {
    // 表单验证失败，不处理（表单会自动显示错误）
    console.log('表单验证失败:', error)
    return
  }
  
  try {
    loginLoading.value = true
    
    const loginData: LoginRequest = {
      username: loginForm.username,
      password: encryptPassword(loginForm.password) // SHA256加密密码
    }
    
    const response = await login(loginData)
    
    // 保存token和用户信息
    userStore.setToken(response.data.token)
    userStore.setUserInfo(response.data.user)
    
    // 记住我功能 - 只记住用户名
    if (loginForm.remember) {
      localStorage.setItem('remembered_username', loginForm.username)
    } else {
      localStorage.removeItem('remembered_username')
    }
    
    message.success(t.value.login_success)
    
    // 跳转到首页
    router.push('/')
  } catch (error: any) {
    console.error('登录失败:', error)
    message.error(error.message || t.value.login_error)
  } finally {
    loginLoading.value = false
  }
}

// 处理注册
const handleRegister = async () => {
  try {
    await registerFormRef.value?.validate()
  } catch (error: any) {
    // 表单验证失败，不处理（表单会自动显示错误）
    console.log('表单验证失败:', error)
    return
  }
  
  try {
    registerLoading.value = true
    
    const registerData = {
      username: registerForm.username,
      email: registerForm.email,
      password: encryptPassword(registerForm.password) // SHA256加密密码
    }
    
    await register(registerData)
    
    message.success(t.value.register_success)
    
    // 切换到登录标签页，并填充用户名
    activeTab.value = 'login'
    loginForm.username = registerForm.username
    loginForm.password = ''
    
    // 清空注册表单
    registerForm.username = ''
    registerForm.email = ''
    registerForm.password = ''
    registerForm.confirmPassword = ''
  } catch (error: any) {
    console.error('注册失败:', error)
    message.error(error.message || t.value.register_error)
  } finally {
    registerLoading.value = false
  }
}

// 回到首页
const goHome = () => {
  router.push('/')
}

// 初始化记住的用户名
const initRememberedUsername = () => {
  const remembered = localStorage.getItem('remembered_username')
  if (remembered) {
    loginForm.username = remembered
    loginForm.remember = true
  }
}

// 组件挂载时初始化
onMounted(() => {
  // 初始化语言
  const savedLang = localStorage.getItem('lang') || detectLang()
  currentLang.value = savedLang
  document.title = langDict[savedLang].page_title
  
  // 初始化记住的用户名
  initRememberedUsername()
})
</script>

<template>
  <div class="login-container">
    <!-- 粒子背景 -->
    <div id="particles-js" class="particles-bg"></div>
    
    <!-- 返回首页按钮 -->
    <div class="back-home">
      <n-button text @click="goHome" class="back-btn">
        ← {{ t.back_home }}
      </n-button>
    </div>
    
    <!-- 登录/注册卡片 -->
    <div class="login-card-wrapper">
      <n-card class="login-card" :bordered="false">
        <!-- 语言切换 - 右上角 -->
        <div class="lang-switch">
          <n-dropdown 
            :options="langOptions"
            @select="switchLang"
            trigger="click"
          >
            <n-button size="small" class="lang-dropdown-btn" circle>
              <img 
                v-if="currentLang === 'zh'" 
                :src="cnFlag" 
                alt="中文" 
                class="flag-icon"
              />
              <img 
                v-else-if="currentLang === 'ru'" 
                :src="ruFlag" 
                alt="Русский" 
                class="flag-icon"
              />
              <img 
                v-else 
                :src="usFlag" 
                alt="English" 
                class="flag-icon"
              />
            </n-button>
          </n-dropdown>
        </div>
        
        <!-- Logo -->
        <div class="logo-section">
          <img 
            alt="Logo" 
            class="logo" 
            src="@/assets/images/ChatGPT-Logo.svg"
          />
          <h1 class="title">{{ t.login_title }}</h1>
          <p class="subtitle">{{ t.login_subtitle }}</p>
        </div>
        
        <!-- 标签页 -->
        <n-tabs
          v-model:value="activeTab"
          type="segment"
          animated
          size="large"
          class="tabs"
        >
          <!-- 登录标签页 -->
          <n-tab-pane name="login" :tab="t.login_tab">
            <n-form
              ref="loginFormRef"
              :model="loginForm"
              :rules="loginRules"
              size="large"
              class="form"
            >
              <n-form-item path="username" :show-label="false">
                <n-input
                  v-model:value="loginForm.username"
                  :placeholder="t.login_username_placeholder"
                  clearable
                >
                  <template #prefix>
                    <span class="input-icon">👤</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <n-form-item path="password" :show-label="false">
                <n-input
                  v-model:value="loginForm.password"
                  type="password"
                  :placeholder="t.login_password_placeholder"
                  show-password-on="click"
                  clearable
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <span class="input-icon">🔒</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <div class="form-footer">
                <n-checkbox v-model:checked="loginForm.remember">
                  {{ t.login_remember }}
                </n-checkbox>
                <n-button text type="primary" size="small">
                  {{ t.login_forgot }}
                </n-button>
              </div>
              
              <n-button
                type="primary"
                size="large"
                block
                :loading="loginLoading"
                @click="handleLogin"
                class="submit-btn"
              >
                {{ t.login_btn }}
              </n-button>
            </n-form>
          </n-tab-pane>
          
          <!-- 注册标签页 -->
          <n-tab-pane name="register" :tab="t.register_tab">
            <n-form
              ref="registerFormRef"
              :model="registerForm"
              :rules="registerRules"
              size="large"
              class="form"
            >
              <n-form-item path="username" :show-label="false">
                <n-input
                  v-model:value="registerForm.username"
                  :placeholder="t.register_username_placeholder"
                  clearable
                >
                  <template #prefix>
                    <span class="input-icon">👤</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <n-form-item path="email" :show-label="false">
                <n-input
                  v-model:value="registerForm.email"
                  :placeholder="t.register_email_placeholder"
                  clearable
                >
                  <template #prefix>
                    <span class="input-icon">✉️</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <n-form-item path="password" :show-label="false">
                <n-input
                  v-model:value="registerForm.password"
                  type="password"
                  :placeholder="t.register_password_placeholder"
                  show-password-on="click"
                  clearable
                >
                  <template #prefix>
                    <span class="input-icon">🔒</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <n-form-item path="confirmPassword" :show-label="false">
                <n-input
                  v-model:value="registerForm.confirmPassword"
                  type="password"
                  :placeholder="t.register_confirm_placeholder"
                  show-password-on="click"
                  clearable
                  @keyup.enter="handleRegister"
                >
                  <template #prefix>
                    <span class="input-icon">🔐</span>
                  </template>
                </n-input>
              </n-form-item>
              
              <n-button
                type="primary"
                size="large"
                block
                :loading="registerLoading"
                @click="handleRegister"
                class="submit-btn"
              >
                {{ t.register_btn }}
              </n-button>
            </n-form>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  overflow: hidden;
}

.particles-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.back-home {
  position: absolute;
  top: 20px;
  left: 20px;
  z-index: 10;
}

.back-btn {
  color: white;
  font-size: 16px;
  font-weight: 500;
  padding: 8px 16px;
  border-radius: 8px;
  transition: all 0.3s;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateX(-4px);
}

.login-card-wrapper {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 440px;
  animation: slideInUp 0.6s ease-out;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-card {
  position: relative;
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.98);
  padding: 20px;
  border: 1px solid rgba(230, 230, 230, 0.5);
}

.logo-section {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  width: 60px;
  height: 60px;
  margin-bottom: 16px;
  animation: rotate 20s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #666;
  margin: 0;
  font-weight: 400;
}

.lang-switch {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 10;
}

.lang-dropdown-btn {
  width: 40px !important;
  height: 40px !important;
  padding: 0 !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  transition: all 0.3s;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid #e0e0e0;
  border-radius: 50% !important;
  overflow: hidden;
}

.lang-dropdown-btn :deep(.n-button__content) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.lang-dropdown-btn:hover {
  background: rgba(102, 126, 234, 0.1);
  transform: scale(1.05);
  border-color: rgba(102, 126, 234, 0.3);
}

.flag-icon {
  width: 24px;
  height: 18px;
  border-radius: 2px;
  object-fit: cover;
  display: block;
}

.tabs {
  margin-top: 24px;
}

.tabs :deep(.n-tabs-nav) {
  background: #f5f5f5;
  padding: 4px;
  border-radius: 10px;
}

.tabs :deep(.n-tabs-tab) {
  font-weight: 600;
  font-size: 15px;
  padding: 10px 20px;
}

.tabs :deep(.n-tabs-tab--active) {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form {
  margin-top: 24px;
}

.form :deep(.n-input) {
  border: 1.5px solid #e0e0e0;
  transition: all 0.3s;
}

.form :deep(.n-input:hover) {
  border-color: #b0b0b0;
}

.form :deep(.n-input.n-input--focus) {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.1);
}

.form :deep(.n-input__input) {
  font-size: 15px;
}

.input-icon {
  font-size: 18px;
  margin-right: 4px;
}

.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.submit-btn {
  height: 48px;
  font-size: 16px;
  font-weight: 700;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: 2px solid transparent;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
  transition: all 0.3s;
  letter-spacing: 0.5px;
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.5);
  border-color: rgba(255, 255, 255, 0.3);
}

.submit-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.3);
}

/* 响应式设计 */
@media (max-width: 640px) {
  .login-card {
    padding: 16px;
  }
  
  .title {
    font-size: 20px;
  }
  
  .logo {
    width: 50px;
    height: 50px;
  }
}
</style>

