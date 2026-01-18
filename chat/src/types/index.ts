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
  id: number
  username: string
  email: string
  display_name?: string
  avatar?: string
  role?: number
  status?: number
}

// 登录请求类型
export interface LoginRequest {
  username: string
  password: string
}

// 注册请求类型
export interface RegisterRequest {
  username: string
  email: string
  password: string
  source?: string
}

// 登录响应数据类型
export interface LoginResponseData {
  token: string
  user: User
}

// 登录响应类型（包含API响应包装）
export interface LoginResponse extends ApiResponse<LoginResponseData> {}

