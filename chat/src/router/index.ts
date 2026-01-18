import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
  },
  {
    path: '/about',
    name: 'About',
    component: () => import('@/views/About.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresGuest: true } // 只允许未登录用户访问
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 全局前置守卫
router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  
  // 初始化 token 和用户信息（从 localStorage 读取）
  if (!userStore.token) {
    userStore.initToken()
  }
  
  // 如果路由需要游客身份（如登录页），但用户已登录
  if (to.meta.requiresGuest && userStore.token && userStore.userInfo) {
    // 跳转到首页
    next({ name: 'Home' })
  } else {
    // 允许访问
    next()
  }
})

export default router

