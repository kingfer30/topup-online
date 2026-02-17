<script setup lang="ts">
import { ref, onMounted, h, Component, computed } from 'vue'
import { 
  NCard, 
  NMenu, 
  NIcon, 
  NAvatar,
  NInput,
  NEmpty,
  NSpin,
  useMessage
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import Header from '@/components/Layout/Header.vue'
import '@/styles/custom.css'
import { 
  ChatbubbleEllipsesOutline,
  SettingsOutline,
  StarOutline,
  TimeOutline,
  SearchOutline,
  OpenOutline,
  PeopleOutline,
  ArrowUpOutline
} from '@vicons/ionicons5'
import { getRoomList, joinRoom } from '@/api/user'
import type { RoomItem } from '@/types'

const message = useMessage()

// 菜单状态
const activeKey = ref('chatgpt-1')

// 渲染菜单图标（Ionicons）
const renderMenuIcon = (icon: Component) => {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 渲染 logo 图标
const renderLogoIcon = (logoPath: string) => {
  return () => h('img', {
    src: logoPath,
    style: {
      width: '18px',
      height: '18px',
      objectFit: 'contain'
    }
  })
}

// 菜单项配置
const menuOptions: MenuOption[] = [
  {
    label: 'ChatGPT一区',
    key: 'chatgpt-1',
    icon: renderLogoIcon('/logo/icon-chatgpt.svg')
  },
  {
    label: () => h('div', { style: 'display: flex; align-items: center;' }, [
      h('span', 'ChatGPT二区'),
      h(NIcon, { size: 16, style: { marginLeft: '4px', color: '#18a058' } }, { default: () => h(OpenOutline) })
    ]),
    key: 'chatgpt-2',
    icon: renderLogoIcon('/logo/icon-chatgpt.svg')
  },
  {
    label: () => h('div', { style: 'display: flex; align-items: center;' }, [
      h('span', 'ChatGPT三区'),
      h(NIcon, { size: 16, style: { marginLeft: '4px', color: '#18a058' } }, { default: () => h(OpenOutline) })
    ]),
    key: 'chatgpt-3',
    icon: renderLogoIcon('/logo/icon-chatgpt.svg')
  },
  {
    label: 'Claude一区',
    key: 'claude-1',
    icon: renderLogoIcon('/logo/icon-claude.png')
  },
  {
    label: 'Gemini一区',
    key: 'gemini-1',
    icon: renderLogoIcon('/logo/icon-gemini.png')
  },
  {
    label: 'Midjourney一区',
    key: 'midjourney-1',
    icon: renderLogoIcon('/logo/icon-mj.jpg')
  },
  {
    label: '我的收藏',
    key: 'favorite',
    icon: renderMenuIcon(StarOutline)
  },
  {
    label: '最近访问',
    key: 'recent',
    icon: renderMenuIcon(TimeOutline)
  },
  {
    label: '设置',
    key: 'settings',
    icon: renderMenuIcon(SettingsOutline)
  }
]

// 房间列表数据
interface Room {
  id: string
  name: string
  description: string
  members: number
  occupancyRate: number // 拥挤度 0-100
  avatar?: string
  lastActivity: string
  isFavorite: boolean
  isFault: boolean // 是否故障
}

const searchKeyword = ref('')
const rooms = ref<Room[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const currentPage = ref(1)
const pageSize = ref(50)
const totalCount = ref(0)
const hasMore = ref(true)
const contentRef = ref<HTMLElement | null>(null)
const showBackToTop = ref(false)

// 平台类型映射
const platformTypeMap: Record<string, string> = {
  'chatgpt-1': 'gpt-1',
  'chatgpt-2': 'gpt-2',
  'chatgpt-3': 'gpt-3',
  'claude-1': 'claude-1',
  'gemini-1': 'gemini-1',
  'midjourney-1': 'mj-1'
}

// 加载房间列表
const loadRoomList = async (isLoadMore = false) => {
  const platformType = platformTypeMap[activeKey.value]
  
  // 如果不是支持的平台类型，不请求
  if (!platformType) {
    return
  }
  
  // 如果是加载更多，使用 loadingMore，否则使用 loading
  if (isLoadMore) {
    loadingMore.value = true
  } else {
    loading.value = true
  }
  
  try {
    const response = await getRoomList({
      page: currentPage.value,
      pageSize: pageSize.value,
      platform_type: platformType
    })
    
    if (response.code === 200 && response.data) {
      // 转换后端数据为前端格式
      const newRooms = response.data.list.map((item: RoomItem) => {
        // 计算占比：count/100，超过100按100%
        const occupancyRate = Math.min(Math.round((item.count / 100) * 100), 100)
        
        return {
          id: item.carid,
          name: item.carname,
          description: `房间ID: ${item.carid}`,
          members: item.count,
          occupancyRate: occupancyRate,
          lastActivity: '刚刚',
          isFavorite: false,
          isFault: !item.available // available 为 false 时表示故障
        }
      })
      
      // 如果是加载更多，追加数据；否则替换数据
      if (isLoadMore) {
        rooms.value = [...rooms.value, ...newRooms]
      } else {
        rooms.value = newRooms
      }
      
      totalCount.value = response.data.totalCount
      currentPage.value = response.data.page
      
      // 判断是否还有更多数据
      if (newRooms.length === 0 || rooms.value.length >= totalCount.value) {
        hasMore.value = false
      } else {
        hasMore.value = true
      }
    }
  } catch (error: any) {
    message.error(error.message || '获取房间列表失败')
    if (!isLoadMore) {
      rooms.value = []
    }
  } finally {
    if (isLoadMore) {
      loadingMore.value = false
    } else {
      loading.value = false
    }
  }
}

// 搜索过滤后的房间列表
const filteredRooms = computed(() => {
  if (!searchKeyword.value.trim()) {
    return rooms.value
  }
  
  const keyword = searchKeyword.value.toLowerCase()
  return rooms.value.filter(room => 
    room.name.toLowerCase().includes(keyword) || 
    room.description.toLowerCase().includes(keyword)
  )
})

// 处理菜单点击
const handleMenuSelect = (key: string) => {
  activeKey.value = key
  
  // 处理外链跳转
  if (key === 'chatgpt-2') {
    window.open('https://chatgpt.com', '_blank')
    return
  } else if (key === 'chatgpt-3') {
    window.open('https://chat.openai.com', '_blank')
    return
  }
  
  // 如果是房间列表相关的菜单，重新加载数据
  if (platformTypeMap[key]) {
    currentPage.value = 1 // 重置页码
    hasMore.value = true // 重置加载更多状态
    loadRoomList()
  }
  
  const menuItem = menuOptions.find(m => m.key === key)
  if (menuItem) {
    const label = typeof menuItem.label === 'function' ? key : menuItem.label
    message.info(`切换到: ${label}`)
  }
}

// 加载更多数据
const loadMore = async () => {
  // 如果正在加载或没有更多数据，不执行
  if (loadingMore.value || !hasMore.value || loading.value) {
    return
  }
  
  // 页码+1
  currentPage.value += 1
  await loadRoomList(true)
}

// 处理滚动事件
const handleScroll = (e: Event) => {
  const target = e.target as HTMLElement
  const scrollTop = target.scrollTop
  const scrollHeight = target.scrollHeight
  const clientHeight = target.clientHeight
  
  // 判断是否显示返回顶部按钮（滚动超过300px时显示）
  showBackToTop.value = scrollTop > 300
  
  // 判断是否滚动到底部（距离底部小于100px时触发）
  if (scrollHeight - scrollTop - clientHeight < 100) {
    loadMore()
  }
}

// 返回顶部
const scrollToTop = () => {
  if (contentRef.value) {
    contentRef.value.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }
}

// 进入房间
const enterRoom = async (room: Room) => {
  if (room.isFault) {
    message.error('该房间正在维护中，暂时无法进入')
    return
  }
  
  // 显示加载提示
  const loadingMsg = message.loading('正在进入房间...', { duration: 0 })
  
  try {
    // 调用后台接口
    const response = await joinRoom(room.id)
    
    // 关闭加载提示
    loadingMsg.destroy()
    
    if (response.code === 200 && response.data) {
      message.success(`进入房间: ${room.name}`)
      
      // 跳转到反向代理路由
      // 后端会自动处理token和代理转发
      window.location.href = `/api/rooms/${room.id}`
    } else {
      message.error(response.message || '进入房间失败')
    }
  } catch (error: any) {
    // 关闭加载提示
    loadingMsg.destroy()
    message.error(error.message || '进入房间失败，请稍后重试')
  }
}

// 切换收藏状态
const toggleFavorite = (room: Room, event: Event) => {
  event.stopPropagation()
  room.isFavorite = !room.isFavorite
  message.success(room.isFavorite ? '已添加到收藏' : '已取消收藏')
}

// 获取拥挤状态信息
const getOccupancyInfo = (rate: number) => {
  if (rate < 50) {
    return { text: '空闲', color: '#18a058', bgColor: 'rgba(24, 160, 88, 0.1)' }
  } else if (rate < 100) {
    return { text: '拥挤', color: '#f0a020', bgColor: 'rgba(240, 160, 32, 0.1)' }
  } else {
    return { text: '爆满', color: '#d03050', bgColor: 'rgba(208, 48, 80, 0.1)' }
  }
}

onMounted(() => {
  // 页面初始化时加载房间列表（chatgpt-1）
  loadRoomList()
})
</script>

<template>
  <div class="rooms-page">
    <!-- Header 组件 -->
    <Header />

    <!-- 主内容区域 -->
    <div class="main-container">
      <!-- 左侧菜单 -->
      <div class="sidebar">
        <div class="sidebar-header">
          <h2 class="sidebar-title">
            <NIcon size="24">
              <ChatbubbleEllipsesOutline />
            </NIcon>
            聊天室
          </h2>
        </div>
        
        <NMenu
          v-model:value="activeKey"
          :options="menuOptions"
          @update:value="handleMenuSelect"
        />
      </div>

      <!-- 右侧房间列表 -->
      <div class="content" ref="contentRef" @scroll="handleScroll">
      <div class="content-header">
        <h1 class="page-title">房间列表</h1>
        
        <div class="search-bar">
          <NInput
            v-model:value="searchKeyword"
            placeholder="搜索房间名称或描述..."
            size="large"
            clearable
          >
            <template #prefix>
              <NIcon>
                <SearchOutline />
              </NIcon>
            </template>
          </NInput>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <NSpin size="large" />
      </div>

      <!-- 房间网格 -->
      <div v-else class="rooms-grid">
        <NCard
          v-for="room in filteredRooms"
          :key="room.id"
          :class="['room-card', { 'room-card-fault': room.isFault }]"
          hoverable
          @click="enterRoom(room)"
        >
          <div class="room-card-content">
            <!-- 故障标签 -->
            <div v-if="room.isFault" class="fault-badge">
              维护中
            </div>
            
            <!-- 房间头部 -->
            <div class="room-header">
              <div class="room-info">
                <NAvatar
                  :size="48"
                  :style="{ backgroundColor: room.isFault ? '#909399' : '#18a058' }"
                >
                  {{ room.name.substring(0, 2) }}
                </NAvatar>
                <div class="room-title-wrapper">
                  <div class="room-title-row">
                    <h3 class="room-name">{{ room.name }}</h3>
                    <NIcon
                      :size="20"
                      :color="room.isFavorite ? '#f0a020' : '#d0d0d0'"
                      class="favorite-icon"
                      @click="toggleFavorite(room, $event)"
                    >
                      <StarOutline />
                    </NIcon>
                  </div>
                  
                  <!-- 拥挤状态进度条 -->
                  <div class="occupancy-status">
                    <div class="occupancy-header">
                      <span class="occupancy-label">{{ getOccupancyInfo(room.occupancyRate).text }}</span>
                      <span class="occupancy-rate">{{ room.occupancyRate }}%</span>
                    </div>
                    <div class="occupancy-bar">
                      <div 
                        class="occupancy-progress"
                        :style="{ 
                          width: room.occupancyRate + '%',
                          backgroundColor: getOccupancyInfo(room.occupancyRate).color
                        }"
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 房间描述 -->
            <p class="room-description">{{ room.description }}</p>

            <!-- 房间状态 -->
            <div class="room-footer">
              <div class="room-stats">
                <span class="stat-item">
                  <NIcon size="16">
                    <PeopleOutline />
                  </NIcon>
                  {{ room.members }} 成员
                </span>
              </div>
              <span class="last-activity">{{ room.lastActivity }}</span>
            </div>
          </div>
        </NCard>
      </div>

      <!-- 加载更多状态 -->
      <div v-if="!loading && filteredRooms.length > 0" class="load-more-container">
        <div v-if="loadingMore" class="loading-more">
          <NSpin size="small" />
          <span class="loading-text">加载中...</span>
        </div>
        <div v-else-if="!hasMore" class="no-more">
          已经到底啦！
        </div>
      </div>

      <!-- 空状态 -->
      <NEmpty
        v-if="!loading && filteredRooms.length === 0"
        description="暂无房间"
        class="empty-state"
      />
      </div>
    </div>
    
    <!-- 返回顶部按钮 -->
    <transition name="fade">
      <div v-if="showBackToTop" class="back-to-top" @click="scrollToTop">
        <NIcon size="24" color="#fff">
          <ArrowUpOutline />
        </NIcon>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.rooms-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f5f7fa;
}

/* 主内容区域 */
.main-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 左侧菜单样式 */
.sidebar {
  width: 240px;
  background-color: #ffffff;
  border-right: 1px solid #e8eaed;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.sidebar-header {
  padding: 24px 20px;
  border-bottom: 1px solid #e8eaed;
}

.sidebar-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 右侧内容区样式 */
.content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.content-header {
  margin-bottom: 24px;
}

.page-title {
  margin: 0 0 16px 0;
  font-size: 28px;
  font-weight: 600;
  color: #333;
}

.search-bar {
  max-width: 500px;
}

/* 房间网格布局 */
.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.room-card {
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.room-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

/* 故障状态样式 */
.room-card-fault {
  opacity: 0.6;
  cursor: not-allowed;
  filter: grayscale(80%);
}

.room-card-fault:hover {
  transform: none;
  box-shadow: none;
}

.fault-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  background-color: #d03050;
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  z-index: 10;
}

.room-card-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 房间头部 */
.room-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.room-info {
  display: flex;
  gap: 12px;
  flex: 1;
}

.room-title-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.room-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.room-name {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
  flex: 1;
}

.favorite-icon {
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.favorite-icon:hover {
  transform: scale(1.2);
}

/* 拥挤状态进度条 */
.occupancy-status {
  width: 100%;
  margin-top: 4px;
}

.occupancy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.occupancy-label {
  font-size: 12px;
  font-weight: 600;
  color: #666;
}

.occupancy-rate {
  font-size: 12px;
  font-weight: 600;
  color: #999;
}

.occupancy-bar {
  width: 100%;
  height: 6px;
  background-color: #f0f0f0;
  border-radius: 3px;
  overflow: hidden;
}

.occupancy-progress {
  height: 100%;
  border-radius: 3px;
  transition: all 0.3s ease;
}

.room-description {
  margin: 0;
  font-size: 14px;
  color: #666;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 房间底部 */
.room-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.room-stats {
  display: flex;
  gap: 16px;
  align-items: center;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #666;
}

.last-activity {
  font-size: 12px;
  color: #999;
}

/* 空状态 */
.empty-state {
  margin-top: 100px;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

/* 加载更多和到底提示 */
.load-more-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 24px 0;
  margin-top: 20px;
}

.loading-more {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 14px;
}

.loading-text {
  color: #666;
}

.no-more {
  color: #999;
  font-size: 14px;
  padding: 12px 24px;
  background-color: #f5f7fa;
  border-radius: 20px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .main-container {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid #e8eaed;
  }

  .rooms-grid {
    grid-template-columns: 1fr;
  }

  .search-bar {
    max-width: 100%;
    width: 100%;
  }
}

/* 滚动条样式 */
.sidebar::-webkit-scrollbar,
.content::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-track,
.content::-webkit-scrollbar-track {
  background: #f5f7fa;
}

.sidebar::-webkit-scrollbar-thumb,
.content::-webkit-scrollbar-thumb {
  background: #d0d0d0;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb:hover,
.content::-webkit-scrollbar-thumb:hover {
  background: #b0b0b0;
}

/* 返回顶部按钮 */
.back-to-top {
  position: fixed;
  right: 40px;
  bottom: 40px;
  width: 50px;
  height: 50px;
  background-color: #18a058;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(24, 160, 88, 0.3);
  transition: all 0.3s ease;
  z-index: 1000;
}

.back-to-top:hover {
  background-color: #16976e;
  transform: translateY(-4px);
  box-shadow: 0 6px 16px rgba(24, 160, 88, 0.4);
}

.back-to-top:active {
  transform: translateY(-2px);
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

.fade-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

/* 响应式：移动端调整按钮位置 */
@media (max-width: 768px) {
  .back-to-top {
    right: 20px;
    bottom: 20px;
    width: 45px;
    height: 45px;
  }
}
</style>

