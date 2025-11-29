<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-purple-600">
    <n-card class="w-full max-w-md shadow-2xl">
      <div class="text-center mb-6">
        <h1 class="text-3xl font-bold text-gray-800 mb-2">后台管理系统</h1>
        <p class="text-gray-500">请登录您的账号</p>
      </div>

      <n-form ref="formRef" :model="formValue" :rules="rules" size="large">
        <n-form-item path="username">
          <n-input
            v-model:value="formValue.username"
            placeholder="用户名"
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <n-icon>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                  <path fill-rule="evenodd" d="M7.5 6a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM3.751 20.105a8.25 8.25 0 0116.498 0 .75.75 0 01-.437.695A18.683 18.683 0 0112 22.5c-2.786 0-5.433-.608-7.812-1.7a.75.75 0 01-.437-.695z" clip-rule="evenodd" />
                </svg>
              </n-icon>
            </template>
          </n-input>
        </n-form-item>

        <n-form-item path="password">
          <n-input
            v-model:value="formValue.password"
            type="password"
            show-password-on="click"
            placeholder="密码"
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <n-icon>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                  <path fill-rule="evenodd" d="M12 1.5a5.25 5.25 0 00-5.25 5.25v3a3 3 0 00-3 3v6.75a3 3 0 003 3h10.5a3 3 0 003-3v-6.75a3 3 0 00-3-3v-3c0-2.9-2.35-5.25-5.25-5.25zm3.75 8.25v-3a3.75 3.75 0 10-7.5 0v3h7.5z" clip-rule="evenodd" />
                </svg>
              </n-icon>
            </template>
          </n-input>
        </n-form-item>

        <n-form-item>
          <n-checkbox v-model:checked="formValue.remember">
            记住我
          </n-checkbox>
        </n-form-item>

        <n-form-item>
          <n-button
            type="primary"
            block
            size="large"
            :loading="loading"
            @click="handleLogin"
          >
            登录
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>
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

// MD5加密函数（简化版）
const md5 = (str: string) => {
  // 这里使用浏览器内置的crypto API进行简单加密
  // 在生产环境中建议使用专门的MD5库如crypto-js
  const encoder = new TextEncoder()
  const data = encoder.encode(str)
  return crypto.subtle.digest('MD5', data).then(hash => {
    return Array.from(new Uint8Array(hash))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('')
  }).catch(() => {
    // 如果不支持MD5，使用SHA-256作为替代
    return crypto.subtle.digest('SHA-256', data).then(hash => {
      return Array.from(new Uint8Array(hash))
        .map(b => b.toString(16).padStart(2, '0'))
        .join('')
    })
  })
}

const handleLogin = () => {
  formRef.value?.validate(async (errors: any) => {
    if (!errors) {
      loading.value = true
      
      try {
        // 对密码进行MD5加密
        const hashedPassword = await md5(formValue.value.password)
        
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

