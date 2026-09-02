import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    // TODO: 主站暂不对外开放，先临时指向占位页；后续开放时改回 Home.vue 即可
    path: '/',
    name: 'Home',
    component: () => import('@/views/ComingSoon.vue'),
  },
  {
    // TODO: 主站暂不对外开放，先临时指向占位页；后续开放时改回 About.vue 即可
    path: '/about',
    name: 'About',
    component: () => import('@/views/ComingSoon.vue'),
  },
  {
    // TODO: 主站暂不对外开放，先临时指向占位页；后续开放时改回 Topup.vue 即可
    path: '/topup',
    name: 'Topup',
    component: () => import('@/views/ComingSoon.vue'),
  },
  {
    // 完全独立页：Cursor 短信验证码查询，例如 /sms/cursor?account----pass
    path: '/sms/cursor',
    name: 'SmsCursor',
    component: () => import('@/views/SmsCursor.vue'),
  },
  {
    // 其余未定义路径统一显示占位页
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/ComingSoon.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
