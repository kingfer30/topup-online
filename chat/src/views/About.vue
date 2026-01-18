<template>
  <div class="min-h-screen bg-gradient-to-br from-purple-50 to-pink-100">
    <!-- 导航栏 -->
    <nav class="bg-white shadow-lg">
      <div class="container-custom">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center">
            <h1 class="text-2xl font-bold text-primary-600">充值在线</h1>
          </div>
          <div class="flex gap-4">
            <router-link to="/" class="text-gray-700 hover:text-primary-600 px-3 py-2 rounded-md text-sm font-medium">
              首页
            </router-link>
            <router-link to="/about" class="text-gray-700 hover:text-primary-600 px-3 py-2 rounded-md text-sm font-medium">
              关于
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- 主内容 -->
    <div class="container-custom py-12">
      <div class="card max-w-3xl mx-auto">
        <h2 class="text-3xl font-bold text-gray-900 mb-6">关于项目</h2>
        
        <n-space vertical :size="24">
          <div>
            <h3 class="text-xl font-semibold mb-3">技术栈</h3>
            <n-list bordered>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Vue 3:</span>
                  <span class="text-gray-600">渐进式JavaScript框架</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Vite:</span>
                  <span class="text-gray-600">下一代前端构建工具</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Naive UI:</span>
                  <span class="text-gray-600">Vue 3 组件库</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Tailwind CSS:</span>
                  <span class="text-gray-600">实用优先的CSS框架</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">TypeScript:</span>
                  <span class="text-gray-600">JavaScript的超集</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Pinia:</span>
                  <span class="text-gray-600">Vue状态管理库</span>
                </div>
              </n-list-item>
              <n-list-item>
                <div class="flex items-center">
                  <span class="font-semibold mr-2">Vue Router:</span>
                  <span class="text-gray-600">Vue官方路由</span>
                </div>
              </n-list-item>
            </n-list>
          </div>

          <div>
            <h3 class="text-xl font-semibold mb-3">项目结构</h3>
            <n-code language="text" :code="projectStructure" />
          </div>

          <div>
            <h3 class="text-xl font-semibold mb-3">快速开始</h3>
            <n-steps vertical :current="3">
              <n-step title="安装依赖">
                <n-code language="bash" code="npm install" />
              </n-step>
              <n-step title="启动开发服务器">
                <n-code language="bash" code="npm run dev" />
              </n-step>
              <n-step title="构建生产版本">
                <n-code language="bash" code="npm run build" />
              </n-step>
            </n-steps>
          </div>
        </n-space>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NSpace, NList, NListItem, NCode, NSteps, NStep } from 'naive-ui'

const projectStructure = `web/
├── src/
│   ├── assets/          # 静态资源
│   ├── components/      # 通用组件
│   ├── views/           # 页面组件
│   ├── router/          # 路由配置
│   ├── stores/          # 状态管理
│   ├── utils/           # 工具函数
│   ├── styles/          # 全局样式
│   ├── App.vue          # 根组件
│   └── main.ts          # 入口文件
├── public/              # 公共资源
├── index.html           # HTML模板
├── vite.config.ts       # Vite配置
├── tailwind.config.js   # Tailwind配置
└── package.json         # 项目配置`
</script>

