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
  id: number
  username: string
  password?: string
  display_name: string
  status: number // 1-启用 2-禁用 3-已删除
  email: string
  github_id?: string
  wechat_id?: string
  lark_id?: string
  oidc_id?: string
  access_token?: string
  aff_code: string
  inviter_id: number
  last_login?: string
  mirror_card_id?: number
  source?: string
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

// 用户列表查询参数
export interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
}

// 用户列表响应
export interface UserListResponse {
  list: User[]
  total: number
  page: number
  page_size: number
}

// 创建用户请求
export interface CreateUserRequest {
  username: string
  password: string
  display_name?: string
  email?: string
  status?: number
  source?: string
}

// 更新用户请求
export interface UpdateUserRequest {
  username?: string
  password?: string
  display_name?: string
  email?: string
  status?: number
  source?: string
  mirror_card_id?: number
}

