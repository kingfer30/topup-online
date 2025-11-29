<template>
  <div class="h-full w-full">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useLoadingBar } from 'naive-ui'
import { useRouter } from 'vue-router'

const loadingBar = useLoadingBar()
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
</script>

