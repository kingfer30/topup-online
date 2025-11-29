<template>
  <header class="bg-white shadow-md">
    <div class="container-custom">
      <div class="flex justify-between items-center h-16">
        <!-- Logo -->
        <div class="flex items-center">
          <router-link to="/" class="text-2xl font-bold text-primary-600 hover:text-primary-700">
            充值在线
          </router-link>
        </div>

        <!-- 导航菜单 -->
        <nav class="hidden md:flex items-center gap-6">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="text-gray-700 hover:text-primary-600 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            active-class="text-primary-600 bg-primary-50"
          >
            {{ item.label }}
          </router-link>
        </nav>

        <!-- 用户操作 -->
        <div class="flex items-center gap-4">
          <n-button type="primary" @click="handleAction">
            开始充值
          </n-button>
        </div>

        <!-- 移动端菜单按钮 -->
        <div class="md:hidden">
          <n-button text @click="showMobileMenu = true">
            <template #icon>
              <span class="text-2xl">☰</span>
            </template>
          </n-button>
        </div>
      </div>
    </div>

    <!-- 移动端菜单抽屉 -->
    <n-drawer v-model:show="showMobileMenu" :width="250" placement="right">
      <n-drawer-content title="菜单">
        <n-menu :options="navItems" @update:value="handleMenuSelect" />
      </n-drawer-content>
    </n-drawer>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDrawer, NDrawerContent, NMenu } from 'naive-ui'

const router = useRouter()
const showMobileMenu = ref(false)

const navItems = [
  { label: '首页', key: 'home', path: '/' },
  { label: '关于', key: 'about', path: '/about' },
]

const handleAction = () => {
  console.log('开始充值')
}

const handleMenuSelect = (key: string) => {
  const item = navItems.find(i => i.key === key)
  if (item) {
    router.push(item.path)
    showMobileMenu.value = false
  }
}
</script>

