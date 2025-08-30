<script setup lang="ts">

</script>

<template>
  <div class="min-h-screen bg-gray-50 p-8">
    <div class="max-w-4xl mx-auto">
      <h1 class="text-4xl font-bold text-center text-blue-600 mb-8">
        🎉 Nuxt 3 + TailwindCSS + Naive UI
      </h1>
      
      <div class="bg-white p-6 rounded-lg shadow-lg mb-6">
        <h2 class="text-2xl font-semibold mb-4">TailwindCSS 测试</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="bg-blue-500 text-white p-4 rounded-lg text-center hover:bg-blue-600 transition-colors">
            蓝色卡片
          </div>
          <div class="bg-green-500 text-white p-4 rounded-lg text-center hover:bg-green-600 transition-colors">
            绿色卡片
          </div>
          <div class="bg-purple-500 text-white p-4 rounded-lg text-center hover:bg-purple-600 transition-colors">
            紫色卡片
          </div>
        </div>
      </div>

      <div class="bg-white p-6 rounded-lg shadow-lg">
        <h2 class="text-2xl font-semibold mb-4">Naive UI 测试</h2>
        <div class="space-y-4">
          <div class="flex flex-wrap gap-4">
            <n-button type="primary">主要按钮</n-button>
            <n-button type="success">成功按钮</n-button>
            <n-button type="warning">警告按钮</n-button>
            <n-button type="error">错误按钮</n-button>
          </div>
          <n-input placeholder="输入测试文本" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>