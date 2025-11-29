<template>
  <div class="login-container min-h-screen flex items-center justify-center">
    <!-- AI粒子动画背景 -->
    <canvas ref="canvasRef" class="particle-canvas"></canvas>
    
    <!-- 渐变背景层 -->
    <div class="gradient-overlay"></div>
    
    <n-card class="w-full max-w-md shadow-2xl login-card">
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
import { ref, onMounted, onUnmounted } from 'vue'
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
const canvasRef = ref<HTMLCanvasElement>()

const formValue = ref({
  username: '',
  password: '',
  remember: false,
})

// AI粒子动画
let animationId: number
let particles: Array<{
  x: number
  y: number
  vx: number
  vy: number
  radius: number
}> = []

const initParticles = () => {
  if (!canvasRef.value) return
  
  const canvas = canvasRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  // 设置canvas尺寸
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight
  
  // 创建粒子
  const particleCount = 80
  particles = []
  
  for (let i = 0; i < particleCount; i++) {
    particles.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      vx: (Math.random() - 0.5) * 0.5,
      vy: (Math.random() - 0.5) * 0.5,
      radius: Math.random() * 2 + 1,
    })
  }
  
  // 动画循环
  const animate = () => {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    // 更新和绘制粒子
    particles.forEach((particle, i) => {
      // 更新位置
      particle.x += particle.vx
      particle.y += particle.vy
      
      // 边界检测
      if (particle.x < 0 || particle.x > canvas.width) particle.vx *= -1
      if (particle.y < 0 || particle.y > canvas.height) particle.vy *= -1
      
      // 绘制粒子
      ctx.beginPath()
      ctx.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2)
      ctx.fillStyle = 'rgba(255, 255, 255, 0.8)'
      ctx.fill()
      
      // 绘制连线
      particles.slice(i + 1).forEach(otherParticle => {
        const dx = particle.x - otherParticle.x
        const dy = particle.y - otherParticle.y
        const distance = Math.sqrt(dx * dx + dy * dy)
        
        if (distance < 150) {
          ctx.beginPath()
          ctx.moveTo(particle.x, particle.y)
          ctx.lineTo(otherParticle.x, otherParticle.y)
          ctx.strokeStyle = `rgba(255, 255, 255, ${0.2 * (1 - distance / 150)})`
          ctx.lineWidth = 0.5
          ctx.stroke()
        }
      })
    })
    
    animationId = requestAnimationFrame(animate)
  }
  
  animate()
}

// 窗口大小改变时重新初始化
const handleResize = () => {
  if (canvasRef.value) {
    canvasRef.value.width = window.innerWidth
    canvasRef.value.height = window.innerHeight
  }
}

// 页面加载时，读取保存的用户名并初始化动画
onMounted(() => {
  const savedUsername = localStorage.getItem('remember_username')
  if (savedUsername) {
    formValue.value.username = savedUsername
    formValue.value.remember = true
  }
  
  // 初始化粒子动画
  setTimeout(() => {
    initParticles()
  }, 100)
  
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
  window.removeEventListener('resize', handleResize)
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

<style scoped>
.login-container {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.particle-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.gradient-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.8) 0%,
    rgba(118, 75, 162, 0.8) 100%
  );
  z-index: 2;
  animation: gradientShift 15s ease infinite;
}

@keyframes gradientShift {
  0%, 100% {
    opacity: 0.8;
  }
  50% {
    opacity: 0.6;
  }
}

.login-card {
  position: relative;
  z-index: 3;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95) !important;
}

/* 为卡片添加悬浮动画 */
.login-card {
  animation: cardFloat 6s ease-in-out infinite;
}

@keyframes cardFloat {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

/* 标题添加AI光效 */
:deep(.n-card__content) h1 {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: titleGlow 3s ease-in-out infinite;
}

@keyframes titleGlow {
  0%, 100% {
    filter: brightness(1);
  }
  50% {
    filter: brightness(1.2);
  }
}
</style>
