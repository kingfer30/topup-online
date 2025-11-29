<template>
  <footer class="bg-gray-800 text-white py-8">
    <div class="container-custom">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
        <!-- 公司信息 -->
        <div>
          <h3 class="text-xl font-bold mb-4">充值在线</h3>
          <p class="text-gray-400 text-sm">
            提供安全、快速、便捷的在线充值服务
          </p>
        </div>

        <!-- 快速链接 -->
        <div>
          <h3 class="text-lg font-semibold mb-4">快速链接</h3>
          <ul class="space-y-2">
            <li>
              <router-link to="/" class="text-gray-400 hover:text-white text-sm transition-colors">
                首页
              </router-link>
            </li>
            <li>
              <router-link to="/about" class="text-gray-400 hover:text-white text-sm transition-colors">
                关于我们
              </router-link>
            </li>
          </ul>
        </div>

        <!-- 联系方式 -->
        <div>
          <h3 class="text-lg font-semibold mb-4">联系我们</h3>
          <ul class="space-y-2 text-sm text-gray-400">
            <li>邮箱: support@topup-online.com</li>
            <li>电话: +86 123-4567-8900</li>
          </ul>
        </div>
      </div>

      <div class="border-t border-gray-700 mt-8 pt-8 text-center text-sm text-gray-400">
        <p>&copy; {{ currentYear }} 充值在线. All rights reserved.</p>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const currentYear = computed(() => new Date().getFullYear())
</script>

