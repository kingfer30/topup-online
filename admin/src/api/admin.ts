import http from '@/utils/http'
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

