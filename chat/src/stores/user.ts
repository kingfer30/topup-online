import { defineStore } from 'pinia'
import { ref } from 'vue'

interface UserInfo {
  id: number
  username: string
  email: string
  display_name?: string
  avatar?: string
  role?: number
  status?: number
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)
  const token = ref<string>('')

  // 设置用户信息
  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  // 设置token
  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
    
    // 同时保存到 Cookie，用于页面跳转时的认证
    // 设置 30 天过期
    const expires = new Date()
    expires.setDate(expires.getDate() + 30)
    document.cookie = `access_token=${newToken}; path=/; expires=${expires.toUTCString()}`
  }

  // 清除用户信息
  const clearUserInfo = () => {
    userInfo.value = null
    token.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    
    // 删除 Cookie
    document.cookie = 'access_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC'
  }

  // 初始化token和用户信息
  const initToken = () => {
    const savedToken = localStorage.getItem('token')
    const savedUserInfo = localStorage.getItem('userInfo')
    
    if (savedToken) {
      token.value = savedToken
      
      // 确保 Cookie 也有 token（防止刷新后 Cookie 丢失）
      const expires = new Date()
      expires.setDate(expires.getDate() + 30)
      document.cookie = `access_token=${savedToken}; path=/; expires=${expires.toUTCString()}`
    }
    
    if (savedUserInfo) {
      try {
        userInfo.value = JSON.parse(savedUserInfo)
      } catch (error) {
        console.error('解析用户信息失败:', error)
        localStorage.removeItem('userInfo')
      }
    }
  }

  return {
    userInfo,
    token,
    setUserInfo,
    setToken,
    clearUserInfo,
    initToken,
  }
})

