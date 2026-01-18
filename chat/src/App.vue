<template>
  <n-config-provider :theme="theme" :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-loading-bar-provider>
            <AppContent />
          </n-loading-bar-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  NLoadingBarProvider,
  zhCN,
  dateZhCN,
  darkTheme,
  lightTheme,
} from 'naive-ui'
import AppContent from './components/AppContent.vue'

// 主题切换
const isDark = ref(false)
const theme = ref(isDark.value ? darkTheme : lightTheme)

// 可以通过provide/inject在子组件中切换主题
// provide('toggleTheme', () => {
//   isDark.value = !isDark.value
//   theme.value = isDark.value ? darkTheme : lightTheme
// })
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
  width: 100%;
}
</style>

