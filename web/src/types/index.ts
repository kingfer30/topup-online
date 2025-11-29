// 全局类型定义

// API 响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 分页数据类型
export interface PageData<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 用户类型
export interface User {
  id: string
  username: string
  email: string
  avatar?: string
  createdAt: string
  updatedAt: string
}

// 登录请求类型
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应类型
export interface LoginResponse {
  token: string
  user: User
}

