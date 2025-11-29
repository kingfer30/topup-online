import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { http } from '@/utils/http'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/initialize',
    name: 'Initialize',
    component: () => import('@/views/Initialize.vue'),
    meta: { title: '系统初始化', noAuth: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/admin/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '控制台', requiresAuth: true },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户列表', requiresAuth: true },
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '角色管理', requiresAuth: true },
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单列表', requiresAuth: true },
      },
      {
        path: 'refunds',
        name: 'Refunds',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '退款管理', requiresAuth: true },
      },
      {
        path: 'cards',
        name: 'Cards',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '卡密列表', requiresAuth: true },
      },
      {
        path: 'card-generate',
        name: 'CardGenerate',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '生成卡密', requiresAuth: true },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '基础设置', requiresAuth: true },
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '操作日志', requiresAuth: true },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 检查系统是否已初始化
  if (to.path !== '/initialize' && !to.meta.noAuth) {
    try {
      const res = await http.get('/system/init/status')
      if (res.code === 200 && res.data.initialized === false) {
        // 系统未初始化，跳转到初始化页面
        next('/initialize')
        return
      }
    } catch (error) {
      console.error('检查初始化状态失败:', error)
    }
  }

  const token = localStorage.getItem('admin_token')
  
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/admin/dashboard')
  } else if (to.path === '/initialize') {
    // 允许访问初始化页面
    next()
  } else {
    next()
  }
})

export default router

