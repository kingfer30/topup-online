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

      <n-menu v-model:value="activeKey" v-model:expanded-keys="expandedKeys" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="22"
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
import { ref, computed, h, onMounted, onUnmounted, watch, nextTick } from 'vue'
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
import { getMenuTree, type Menu } from '@/api/menu'

const router = useRouter()
const route = useRoute()
const message = useMessage()

// 侧边栏折叠状态
const collapsed = ref(false)

// 菜单数据
const menus = ref<Menu[]>([])
const menuPathMap = ref<Record<string, string>>({})
const menuTitleMap = ref<Record<string, string>>({})
const menuParentMap = ref<Record<string, string>>({})

// 渲染图标
const renderIcon = (icon?: string) => {
  if (!icon) return undefined
  return () => h('span', { class: 'text-xl' }, icon)
}

// 将 Menu 转换为 MenuOption
const convertMenuToOption = (menu: Menu, parentKey?: string): MenuOption => {
  const option: MenuOption = {
    label: menu.path
      ? () =>
          h(
            RouterLink,
            {
              to: menu.path as string,
            },
            { default: () => menu.title }
          )
      : menu.title,
    key: menu.key,
    icon: renderIcon(menu.icon),
  }

  // 记录路径和标题映射
  if (menu.path) {
    menuPathMap.value[menu.path] = menu.key
    menuTitleMap.value[menu.path] = menu.title
  }

  // 记录父子关系
  if (parentKey) {
    menuParentMap.value[menu.key] = parentKey
  }

  // 处理子菜单
  if (menu.children && menu.children.length > 0) {
    option.children = menu.children.map(child => convertMenuToOption(child, menu.key))
  }

  return option
}

// 菜单选项
const menuOptions = computed<MenuOption[]>(() => {
  return menus.value.map(menu => convertMenuToOption(menu))
})

// 路由路径到菜单key的映射（包含查询参数）
const routeToMenuKey = (path: string, query: any = {}): string => {
  // 先尝试精确匹配（包含查询参数）
  const queryString = new URLSearchParams(query).toString()
  const fullPath = queryString ? `${path}?${queryString}` : path
  
  if (menuPathMap.value[fullPath]) {
    return menuPathMap.value[fullPath]
  }
  
  // 如果没有精确匹配，尝试只匹配路径
  if (menuPathMap.value[path]) {
    return menuPathMap.value[path]
  }
  
  // 默认返回 dashboard
  return 'dashboard'
}

// 获取菜单项的所有父级菜单key
const getParentKeys = (menuKey: string): string[] => {
  const parents: string[] = []
  let currentKey = menuKey
  
  while (menuParentMap.value[currentKey]) {
    const parentKey = menuParentMap.value[currentKey]
    parents.push(parentKey)
    currentKey = parentKey
  }
  
  return parents
}

// 当前激活的菜单项 - 初始为空，等菜单加载后设置
const activeKey = ref<string>('')

// 展开的菜单项（父菜单）
const expandedKeys = ref<string[]>([])

// 监听路由变化，更新菜单激活状态和展开状态
watch(
  () => [route.path, route.query] as const,
  ([newPath, newQuery]) => {
    const key = routeToMenuKey(newPath, newQuery)
    if (key) {
      activeKey.value = key
      // 展开当前菜单项的所有父级菜单
      expandedKeys.value = getParentKeys(key)
    }
  },
  { immediate: false }
)

// 管理员信息
const adminInfo = ref<any>(null)

// 主题切换
const isDark = ref(false)

const toggleTheme = () => {
  isDark.value = !isDark.value
}

// 加载菜单数据
const loadMenus = async () => {
  try {
    const response = await getMenuTree()
    if (response.code === 200) {
      menus.value = response.data || []
      // 等待 computed 更新完成，确保 menuPathMap 和 menuParentMap 已经构建
      await nextTick()
      // 根据当前路由设置激活的菜单项
      const key = routeToMenuKey(route.path, route.query)
      if (key) {
        activeKey.value = key
        // 展开当前菜单项的所有父级菜单
        expandedKeys.value = getParentKeys(key)
      }
    } else {
      console.error('加载菜单失败:', response.message)
      message.error(response.message || '加载菜单失败')
    }
  } catch (error) {
    console.error('加载菜单失败', error)
    message.error('加载菜单失败')
  }
}

// 加载管理员信息
onMounted(async () => {
  const info = localStorage.getItem('admin_info')
  if (info) {
    try {
      adminInfo.value = JSON.parse(info)
    } catch (e) {
      console.error('解析管理员信息失败', e)
    }
  }

  // 加载菜单
  await loadMenus()

  // 监听菜单刷新事件
  window.addEventListener('refreshMenus', loadMenus)
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  window.removeEventListener('refreshMenus', loadMenus)
})

// 面包屑
const breadcrumbs = computed(() => {
  const path = route.path
  const title = menuTitleMap.value[path] || '详情'

  return [
    { name: 'admin', label: '后台管理' },
    { name: path, label: title },
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
