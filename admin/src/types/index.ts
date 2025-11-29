// 全局类型定义

// API 响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 系统初始化状态
export interface InitStatus {
  initialized: boolean
}

// 数据库配置
export interface DBConfig {
  db_host: string
  db_port: string
  db_name: string
  db_user: string
  db_password: string
}

// 管理员配置
export interface AdminConfig {
  admin_user: string
  admin_pass: string
  admin_email: string
}

// 初始化请求
export interface InitRequest extends DBConfig, AdminConfig {}

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

