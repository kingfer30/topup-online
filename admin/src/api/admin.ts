import { http } from '@/utils/http'
import type {
  ApiResponse,
  UserListParams,
  UserListResponse,
  User,
  CreateUserRequest,
  UpdateUserRequest,
} from '@/types'

// 管理员相关API

// 管理员登录
export const adminLogin = (data: {
  username: string
  password: string // MD5加密后的密码
}) => {
  return http.post<ApiResponse<{ token: string; admin: any }>>('/admin/login', data)
}

// 获取管理员信息
export const getAdminInfo = () => {
  return http.get<ApiResponse<any>>('/admin/info')
}

// 退出登录
export const adminLogout = () => {
  return http.post<ApiResponse>('/admin/logout')
}

// 修改密码
export const changePassword = (data: {
  old_password: string
  new_password: string
}) => {
  return http.post<ApiResponse>('/admin/change-password', data)
}

// ==================== 登录设备管理 ====================

export interface AdminSession {
  id: number
  session_uuid: string
  ip_address: string
  user_agent: string
  created_at: number
  is_current: boolean
}

// 获取已登录设备列表
export const getAdminSessions = () => {
  return http.get<ApiResponse<AdminSession[]>>('/admin/sessions')
}

// 踢出指定设备
export const kickSession = (uuid: string) => {
  return http.delete<ApiResponse>(`/admin/sessions/${uuid}`)
}

// 踢出所有设备（含自己）
export const kickAllSessions = () => {
  return http.delete<ApiResponse>('/admin/sessions')
}

// ==================== 用户管理相关API ====================

// 获取用户列表
export const getUserList = (params: UserListParams) => {
  return http.get<ApiResponse<UserListResponse>>('/admin/users', { params })
}

// 获取用户详情
export const getUserDetail = (id: number) => {
  return http.get<ApiResponse<User>>(`/admin/users/${id}`)
}

// 创建用户
export const createUser = (data: CreateUserRequest) => {
  return http.post<ApiResponse<User>>('/admin/users', data)
}

// 更新用户信息
export const updateUser = (id: number, data: UpdateUserRequest) => {
  return http.put<ApiResponse<User>>(`/admin/users/${id}`, data)
}

// 删除用户
export const deleteUser = (id: number) => {
  return http.delete<ApiResponse>(`/admin/users/${id}`)
}

