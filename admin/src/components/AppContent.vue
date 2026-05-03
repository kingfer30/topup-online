<template>
  <div class="h-full w-full">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useLoadingBar, useDialog } from 'naive-ui'
import { useRouter } from 'vue-router'

const loadingBar = useLoadingBar()
const dialog = useDialog()
const router = useRouter()

// 路由切换时显示加载条
router.beforeEach(() => {
  loadingBar.start()
})

router.afterEach(() => {
  loadingBar.finish()
})

router.onError(() => {
  loadingBar.error()
})

// 防止重复弹出的标志位
let isDialogShown = false

const handleUnauthorized = (event: Event) => {
  // 当前在登录页不弹窗
  if (router.currentRoute.value.path === '/login' || router.currentRoute.value.path === '/initialize') return
  if (isDialogShown) return
  isDialogShown = true

  const msg = (event as CustomEvent).detail?.message || '登录已失效，请重新登录'

  dialog.warning({
    title: '登录状态失效',
    content: msg,
    positiveText: '重新登录',
    closable: false,
    maskClosable: false,
    closeOnEsc: false,
    onPositiveClick: () => {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      isDialogShown = false
      router.push('/login')
    },
  })
}

onMounted(() => {
  window.addEventListener('admin-unauthorized', handleUnauthorized)
})

onUnmounted(() => {
  window.removeEventListener('admin-unauthorized', handleUnauthorized)
  isDialogShown = false
})
</script>

