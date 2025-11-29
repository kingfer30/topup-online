import { defineStore } from 'pinia'
import { ref } from 'vue'

interface UserInfo {
  id: string
  username: string
  email: string
  avatar?: string
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)
  const token = ref<string>('')

  // 设置用户信息
  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  // 设置token
  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  // 清除用户信息
  const clearUserInfo = () => {
    userInfo.value = null
    token.value = ''
    localStorage.removeItem('token')
  }

  // 初始化token
  const initToken = () => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      token.value = savedToken
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

