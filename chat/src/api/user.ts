import { http } from '@/utils/http'
import type { LoginRequest, LoginResponse, RegisterRequest, User, ApiResponse, RoomListParams, RoomListData, JoinRoomData } from '@/types'

// 用户登录
export const login = (data: LoginRequest) => {
  return http.post<LoginResponse>('/user/login', data)
}

// 获取用户信息
export const getUserInfo = () => {
  return http.get<ApiResponse<User>>('/user/info')
}

// 用户登出
export const logout = () => {
  return http.post<ApiResponse>('/user/logout')
}

// 注册用户
export const register = (data: RegisterRequest) => {
  // 添加 source 标识为 chat
  return http.post<ApiResponse>('/user/register', { ...data, source: 'chat' })
}

// 获取房间列表
export const getRoomList = (params: RoomListParams) => {
  return http.get<ApiResponse<RoomListData>>('/user/rooms', { params })
}

// 加入房间
export const joinRoom = (room_id: string) => {
  return http.post<ApiResponse<JoinRoomData>>('/user/room/join', { room_id })
}

