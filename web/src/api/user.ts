import { http } from '@/utils/http'
import type { LoginRequest, LoginResponse, User } from '@/types'

// 用户登录
export const login = (data: LoginRequest) => {
  return http.post<LoginResponse>('/user/login', data)
}

// 获取用户信息
export const getUserInfo = () => {
  return http.get<User>('/user/info')
}

// 用户登出
export const logout = () => {
  return http.post('/user/logout')
}

// 注册用户
export const register = (data: { username: string; email: string; password: string }) => {
  return http.post('/user/register', data)
}

