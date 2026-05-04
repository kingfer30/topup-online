import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

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
    path: '/topup',
    name: 'Topup',
    component: () => import('@/views/Topup.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router

