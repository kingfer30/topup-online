<template>
  <header class="bg-white p-4 sticky top-0 z-40 shadow-xl border-b border-black/20">
    <div class="max-w-7xl mx-auto flex justify-between items-center flex-col sm:flex-row gap-3 sm:gap-0">
      <div>
        <n-button 
          @click="$router.push('/')" 
          class="nav-link-active"
          :bordered="false"
        >
          {{ isHomePage ? t.nav_home : '首页' }}
        </n-button>
      </div>
      <nav class="flex gap-2 items-center justify-center">
        <a href="/rooms" class="nav-link" @click.prevent="$router.push('/rooms')">
          立即开始
        </a>
        <a :href="isHomePage ? '#features' : '/#features'" class="nav-link">
          {{ t.nav_features }}
        </a>
        <a :href="isHomePage ? '#steps' : '/#steps'" class="nav-link">
          {{ t.nav_steps }}
        </a>
        <a :href="isHomePage ? '#faq' : '/#faq'" class="nav-link">
          {{ t.nav_faq }}
        </a>
        <n-dropdown 
          :options="langOptions"
          @select="switchLang"
          trigger="click"
        >
          <n-button size="small" class="ml-2 lang-dropdown-btn" circle>
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
        
        <!-- 黑暗模式切换 -->
        <n-button size="small" circle class="ml-2 lang-dropdown-btn" @click="toggleDarkMode">
          <template #icon>
            <n-icon>
              <SunnyOutline v-if="!isDarkMode" />
              <MoonOutline v-else />
            </n-icon>
          </template>
        </n-button>
        
        <!-- 用户登录状态 -->
        <div v-if="isLoggedIn" class="ml-2">
          <n-dropdown 
            :options="userMenuOptions"
            @select="handleUserMenuSelect"
            trigger="click"
          >
            <n-button size="small" class="user-btn">
              <span class="flex items-center gap-1">
                <span class="user-icon">👤</span>
                {{ displayName }}
              </span>
            </n-button>
          </n-dropdown>
        </div>
        <router-link v-else to="/login" class="ml-2">
          <n-button size="small" type="primary" class="login-btn">
            登录/注册
          </n-button>
        </router-link>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NButton, NDropdown, NIcon, useMessage } from 'naive-ui'
import { SunnyOutline, MoonOutline } from '@vicons/ionicons5'
import { useUserStore } from '@/stores/user'
import { logout } from '@/api/user'

// 导入语言文件
import enLang from '@/lang/en'
import zhLang from '@/lang/zh'
import ruLang from '@/lang/ru'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const userStore = useUserStore()

// 语言管理
const currentLang = ref('zh')
const cnFlag = '/flags/CN.svg'
const usFlag = '/flags/US.svg'
const ruFlag = '/flags/RU.svg'

// 黑暗模式
const isDarkMode = ref(false)

// 语言字典
const langDict: Record<string, any> = {
  en: enLang,
  zh: zhLang,
  ru: ruLang
}

// 语言选项
const langOptions = [
  { label: '中文', key: 'zh', icon: () => h('img', { src: cnFlag, class: 'flag-icon-small' }) },
  { label: 'English', key: 'en', icon: () => h('img', { src: usFlag, class: 'flag-icon-small' }) },
  { label: 'Русский', key: 'ru', icon: () => h('img', { src: ruFlag, class: 'flag-icon-small' }) }
]

// 用户菜单选项
const userMenuOptions = [
  { label: '个人中心', key: 'profile' },
  { label: '退出登录', key: 'logout' }
]

// 计算属性
const t = computed(() => langDict[currentLang.value])
const isLoggedIn = computed(() => !!userStore.token && !!userStore.userInfo)
const displayName = computed(() => userStore.userInfo?.username || userStore.userInfo?.display_name || '用户')
const isHomePage = computed(() => route.path === '/')

// 切换语言
const switchLang = (lang: string) => {
  currentLang.value = lang
  localStorage.setItem('lang', lang)
  if (isHomePage.value) {
    document.title = langDict[lang].page_title
  }
}

// 切换黑暗模式
const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value
  message.info(isDarkMode.value ? '已切换到黑暗模式' : '已切换到明亮模式')
}

// 处理用户菜单点击
const handleUserMenuSelect = async (key: string) => {
  if (key === 'logout') {
    try {
      await logout()
      userStore.clearUserInfo()
      router.push('/login')
    } catch (error) {
      console.error('退出登录失败:', error)
      userStore.clearUserInfo()
      router.push('/login')
    }
  } else if (key === 'profile') {
    message.info('个人中心功能开发中...')
  }
}

// 初始化
onMounted(() => {
  const savedLang = localStorage.getItem('lang') || 'zh'
  currentLang.value = savedLang
  userStore.initToken()
})
</script>

<style scoped>
/* Header 组件特定样式 - 全局导航样式在 custom.css 中定义 */
</style>
