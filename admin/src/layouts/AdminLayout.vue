<template>
  <n-layout has-sider class="h-screen">
    <!-- 左侧菜单栏 -->
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="240" :collapsed="collapsed"
      show-trigger @collapse="collapsed = true" @expand="collapsed = false" :native-scrollbar="false">
      <div class="p-4 flex items-center justify-center border-b border-gray-200">
        <h1 v-if="!collapsed" class="text-xl font-bold text-primary-600">
          后台管理系统
        </h1>
        <h1 v-else class="text-xl font-bold text-primary-600">
          管理
        </h1>
      </div>

      <n-menu v-model:value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="22"
        :options="menuOptions" @update:value="handleMenuSelect" />
    </n-layout-sider>

    <!-- 右侧内容区域 -->
    <n-layout>
      <!-- 顶部导航栏 -->
      <n-layout-header bordered class="h-16 flex items-center justify-between px-6">
        <n-breadcrumb>
          <n-breadcrumb-item v-for="item in breadcrumbs" :key="item.name">
            {{ item.label }}
          </n-breadcrumb-item>
        </n-breadcrumb>

        <n-space>
          <!-- 主题切换 -->
          <n-button circle @click="toggleTheme">
            <template #icon>
              <n-icon v-if="isDark">☀️</n-icon>
              <n-icon v-else>🌙</n-icon>
            </template>
          </n-button>

          <!-- 用户信息 -->
          <n-dropdown :options="userOptions" @select="handleUserAction">
            <n-button>
              <template #icon>
                <n-avatar size="small">
                  {{ adminInfo?.username || '管理员' }}
                </n-avatar>
              </template>
            </n-button>
          </n-dropdown>
        </n-space>
      </n-layout-header>

      <!-- 主内容区域 -->
      <n-layout-content content-style="padding: 24px;" :native-scrollbar="false">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NBreadcrumb,
  NBreadcrumbItem,
  NSpace,
  NButton,
  NIcon,
  NAvatar,
  NDropdown,
  useMessage,
  type MenuOption,
} from 'naive-ui'
import { RouterLink } from 'vue-router'
import { adminLogout } from '@/api/admin'

const router = useRouter()
const route = useRoute()
const message = useMessage()

// 侧边栏折叠状态
const collapsed = ref(false)

// 路由路径到菜单key的映射
const routeToMenuKey = (path: string): string => {
  const pathMap: Record<string, string> = {
    '/admin/dashboard': 'dashboard',
    '/admin/users': 'users',
    '/admin/roles': 'roles',
    '/admin/orders': 'orders',
    '/admin/refunds': 'refunds',
    '/admin/cards': 'cards',
    '/admin/card-generate': 'card-generate',
    '/admin/mirror-cards': 'mirror-cards',
    '/admin/settings': 'settings',
    '/admin/logs': 'logs',
  }
  return pathMap[path] || 'dashboard'
}

// 当前激活的菜单项 - 根据当前路由初始化
const activeKey = ref<string>(routeToMenuKey(route.path))

// 监听路由变化，更新菜单激活状态
watch(
  () => route.path,
  (newPath) => {
    activeKey.value = routeToMenuKey(newPath)
  }
)

// 管理员信息
const adminInfo = ref<any>(null)

// 主题切换
const isDark = ref(false)

const toggleTheme = () => {
  isDark.value = !isDark.value
}

// 加载管理员信息
onMounted(() => {
  const info = localStorage.getItem('admin_info')
  if (info) {
    try {
      adminInfo.value = JSON.parse(info)
    } catch (e) {
      console.error('解析管理员信息失败', e)
    }
  }
})

// 渲染图标
const renderIcon = (icon: string) => {
  return () => h('span', { class: 'text-xl' }, icon)
}

// 菜单选项
const menuOptions = computed<MenuOption[]>(() => [
  {
    label: () =>
      h(
        RouterLink,
        {
          to: '/admin/dashboard',
        },
        { default: () => '控制台' }
      ),
    key: 'dashboard',
    icon: renderIcon('📊'),
  },
  {
    label: '用户管理',
    key: 'user',
    icon: renderIcon('👥'),
    children: [
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/users',
            },
            { default: () => '用户列表' }
          ),
        key: 'users',
      },
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/roles',
            },
            { default: () => '角色管理' }
          ),
        key: 'roles',
      },
    ],
  },
  {
    label: '订单管理',
    key: 'order',
    icon: renderIcon('📦'),
    children: [
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/orders',
            },
            { default: () => '订单列表' }
          ),
        key: 'orders',
      },
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/refunds',
            },
            { default: () => '退款管理' }
          ),
        key: 'refunds',
      },
    ],
  },
  {
    label: '卡密管理',
    key: 'card',
    icon: renderIcon('🎫'),
    children: [
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/cards',
            },
            { default: () => '卡密列表' }
          ),
        key: 'cards',
      },
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/card-generate',
            },
            { default: () => '生成卡密' }
          ),
        key: 'card-generate',
      },
    ],
  },
  {
    label: '镜像管理',
    key: 'mirror',
    icon: renderIcon('🔐'),
    children: [
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/mirror-cards',
            },
            { default: () => '卡密管理' }
          ),
        key: 'mirror-cards',
      },
    ],
  },
  {
    label: '系统设置',
    key: 'system',
    icon: renderIcon('⚙️'),
    children: [
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/settings',
            },
            { default: () => '基础设置' }
          ),
        key: 'settings',
      },
      {
        label: () =>
          h(
            RouterLink,
            {
              to: '/admin/logs',
            },
            { default: () => '操作日志' }
          ),
        key: 'logs',
      },
    ],
  },
])

// 面包屑
const breadcrumbs = computed(() => {
  const path = route.path
  const breadcrumbMap: Record<string, string> = {
    '/admin': '首页',
    '/admin/dashboard': '控制台',
    '/admin/users': '用户列表',
    '/admin/roles': '角色管理',
    '/admin/orders': '订单列表',
    '/admin/refunds': '退款管理',
    '/admin/cards': '卡密列表',
    '/admin/card-generate': '生成卡密',
    '/admin/mirror-cards': '镜像卡密管理',
    '/admin/settings': '基础设置',
    '/admin/logs': '操作日志',
  }

  return [
    { name: 'admin', label: '后台管理' },
    { name: path, label: breadcrumbMap[path] || '详情' },
  ]
})

// 用户下拉菜单
const userOptions = [
  {
    label: '个人设置',
    key: 'profile',
  },
  {
    label: '修改密码',
    key: 'password',
  },
  {
    type: 'divider',
    key: 'd1',
  },
  {
    label: '退出登录',
    key: 'logout',
  },
]

// 菜单选择处理
const handleMenuSelect = (key: string) => {
  activeKey.value = key
}

// 用户操作处理
const handleUserAction = async (key: string) => {
  if (key === 'logout') {
    try {
      await adminLogout()
      message.success('退出成功')
    } catch (error) {
      console.error('退出登录失败', error)
    } finally {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      router.push('/login')
    }
  } else if (key === 'profile') {
    router.push('/admin/profile')
  } else if (key === 'password') {
    router.push('/admin/change-password')
  }
}
</script>

<style scoped>
:deep(.n-layout-sider) {
  background: #fff;
}

:deep(.n-menu) {
  background: transparent;
}
</style>
