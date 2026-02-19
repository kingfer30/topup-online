<template>
  <div class="login-container min-h-screen flex items-center justify-center">
    <!-- 星空背景 -->
    <canvas ref="starsRef" class="stars-canvas"></canvas>
    
    <!-- Particles.js 容器 -->
    <div id="particles-js"></div>
    
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
import CryptoJS from 'crypto-js'

const router = useRouter()
const message = useMessage()

const formRef = ref()
const loading = ref(false)
const starsRef = ref<HTMLCanvasElement>()

const formValue = ref({
  username: '',
  password: '',
  remember: false,
})

// 星空动画
let starsAnimationId: number
interface Star {
  x: number
  y: number
  radius: number
  vx: number
  vy: number
  opacity: number
}

const initStars = () => {
  if (!starsRef.value) return
  
  const canvas = starsRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight
  
  // 创建星星
  const stars: Star[] = []
  const starCount = 200
  
  for (let i = 0; i < starCount; i++) {
    stars.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      radius: Math.random() * 1.5,
      vx: (Math.random() - 0.5) * 0.3,
      vy: (Math.random() - 0.5) * 0.3,
      opacity: Math.random() * 0.5 + 0.5
    })
  }
  
  const animate = () => {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.05)'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    
    stars.forEach(star => {
      star.x += star.vx
      star.y += star.vy
      
      if (star.x < 0 || star.x > canvas.width) star.vx *= -1
      if (star.y < 0 || star.y > canvas.height) star.vy *= -1
      
      // 闪烁效果
      star.opacity += (Math.random() - 0.5) * 0.02
      star.opacity = Math.max(0.3, Math.min(1, star.opacity))
      
      ctx.beginPath()
      ctx.arc(star.x, star.y, star.radius, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(255, 255, 255, ${star.opacity})`
      ctx.fill()
      
      // 添加星星光晕
      if (star.radius > 1) {
        ctx.beginPath()
        ctx.arc(star.x, star.y, star.radius * 2, 0, Math.PI * 2)
        ctx.fillStyle = `rgba(255, 87, 88, ${star.opacity * 0.2})`
        ctx.fill()
      }
    })
    
    starsAnimationId = requestAnimationFrame(animate)
  }
  
  animate()
}

// 加载 particles.js 库
const loadParticlesJS = (): Promise<void> => {
  return new Promise((resolve) => {
    if (typeof (window as any).particlesJS !== 'undefined') {
      resolve()
      return
    }
    
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/particles.js/2.0.0/particles.min.js'
    script.onload = () => resolve()
    script.onerror = () => {
      console.error('Failed to load particles.js')
      resolve()
    }
    document.head.appendChild(script)
  })
}

// 初始化 particles.js
const initParticles = () => {
  if (typeof (window as any).particlesJS === 'undefined') {
    console.error('particles.js is not loaded')
    return
  }
  
  // 参考 particles.js 官网配置 - 红色主题
  ;(window as any).particlesJS('particles-js', {
    particles: {
      number: {
        value: 80,
        density: {
          enable: true,
          value_area: 800
        }
      },
      color: {
        value: '#ff5758'
      },
      shape: {
        type: 'circle',
        stroke: {
          width: 0,
          color: '#000000'
        },
        polygon: {
          nb_sides: 5
        }
      },
      opacity: {
        value: 0.5,
        random: false,
        anim: {
          enable: false,
          speed: 1,
          opacity_min: 0.1,
          sync: false
        }
      },
      size: {
        value: 3,
        random: true,
        anim: {
          enable: false,
          speed: 40,
          size_min: 0.1,
          sync: false
        }
      },
      line_linked: {
        enable: true,
        distance: 150,
        color: '#ff5758',
        opacity: 0.4,
        width: 1
      },
      move: {
        enable: true,
        speed: 6,
        direction: 'none',
        random: false,
        straight: false,
        out_mode: 'out',
        bounce: false,
        attract: {
          enable: false,
          rotateX: 600,
          rotateY: 1200
        }
      }
    },
    interactivity: {
      detect_on: 'canvas',
      events: {
        onhover: {
          enable: true,
          mode: 'repulse'
        },
        onclick: {
          enable: true,
          mode: 'push'
        },
        resize: true
      },
      modes: {
        grab: {
          distance: 400,
          line_linked: {
            opacity: 1
          }
        },
        bubble: {
          distance: 400,
          size: 40,
          duration: 2,
          opacity: 8,
          speed: 3
        },
        repulse: {
          distance: 200,
          duration: 0.4
        },
        push: {
          particles_nb: 4
        },
        remove: {
          particles_nb: 2
        }
      }
    },
    retina_detect: true
  })
}

// 窗口大小改变时重新初始化
const handleResize = () => {
  if (starsRef.value) {
    starsRef.value.width = window.innerWidth
    starsRef.value.height = window.innerHeight
  }
}

// 页面加载时，读取保存的用户名并初始化动画
onMounted(async () => {
  const savedUsername = localStorage.getItem('remember_username')
  if (savedUsername) {
    formValue.value.username = savedUsername
    formValue.value.remember = true
  }
  
  // 初始化星空背景
  setTimeout(() => {
    initStars()
  }, 50)
  
  // 加载并初始化 particles.js
  await loadParticlesJS()
  setTimeout(() => {
    initParticles()
  }, 100)
  
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  
  // 清理星空动画
  if (starsAnimationId) {
    cancelAnimationFrame(starsAnimationId)
  }
  
  // 清理 particles.js
  if ((window as any).pJSDom && (window as any).pJSDom.length > 0) {
    (window as any).pJSDom[0].pJS.fn.vendors.destroypJS()
    ;(window as any).pJSDom = []
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
.login-container {
  position: relative;
  overflow: hidden;
  background: #000000;
}

/* 星空背景 */
.stars-canvas {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  z-index: 1;
}

/* Particles.js 容器 */
#particles-js {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  z-index: 2;
}

/* 登录卡片 */
.login-card {
  position: relative;
  z-index: 10;
  backdrop-filter: blur(16px) saturate(180%);
  background: rgba(20, 20, 20, 0.85) !important;
  border: 1px solid rgba(255, 87, 88, 0.3);
  box-shadow: 
    0 8px 32px 0 rgba(255, 87, 88, 0.3),
    0 0 40px 0 rgba(255, 87, 88, 0.1),
    0 0 0 1px rgba(255, 87, 88, 0.2) inset;
  transition: all 0.3s ease;
}

.login-card:hover {
  box-shadow: 
    0 12px 40px 0 rgba(255, 87, 88, 0.4),
    0 0 50px 0 rgba(255, 87, 88, 0.2),
    0 0 0 1px rgba(255, 87, 88, 0.3) inset;
}

/* 标题样式 */
:deep(.n-card__content) h1 {
  color: #ff5758 !important;
  text-shadow: 0 0 20px rgba(255, 87, 88, 0.3);
  font-weight: 800;
  letter-spacing: 1px;
}

/* 副标题 */
:deep(.n-card__content) p {
  color: rgba(255, 255, 255, 0.7);
  font-weight: 500;
}

/* 输入框样式 */
:deep(.n-input) {
  background: rgba(40, 40, 40, 0.6) !important;
  border: 1px solid rgba(255, 87, 88, 0.2) !important;
  color: #fff !important;
  transition: all 0.3s ease;
}

:deep(.n-input:focus-within) {
  border-color: rgba(255, 87, 88, 0.6) !important;
  box-shadow: 0 0 15px rgba(255, 87, 88, 0.3);
  transform: translateY(-2px);
}

:deep(.n-input input) {
  color: #fff !important;
}

:deep(.n-input input::placeholder) {
  color: rgba(255, 255, 255, 0.4) !important;
}

:deep(.n-icon) {
  color: rgba(255, 87, 88, 0.8) !important;
}

/* 按钮样式 - 主题色已在全局配置 */
:deep(.n-button--primary) {
  font-weight: 700;
  letter-spacing: 1px;
  box-shadow: 0 4px 20px rgba(255, 87, 88, 0.4);
  transition: all 0.3s ease;
}

:deep(.n-button--primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 25px rgba(255, 87, 88, 0.6);
}

:deep(.n-button--primary:active) {
  transform: translateY(0);
}

/* 复选框样式 - 主题色已在全局配置 */
:deep(.n-checkbox) {
  transition: all 0.3s ease;
}

:deep(.n-checkbox__label) {
  color: rgba(255, 255, 255, 0.9) !important;
}

:deep(.n-checkbox:hover) {
  transform: scale(1.05);
}

/* 响应式 */
@media (max-width: 768px) {
  .login-card {
    margin: 1rem;
  }
}
</style>
