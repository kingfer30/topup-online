import { http } from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface MirrorCard {
  id: number
  username: string
  password: string
  bind_status: number // 0-未绑定 1-已绑定
  bind_user_id: number
  bind_user_name?: string // 绑定用户的用户名
  bind_time: string | null
  status: number // 1-启用 2-禁用
  created_at: string
  updated_at: string
}

export interface MirrorCardListParams {
  page: number
  page_size: number
  keyword?: string
}

export interface MirrorCardListData {
  list: MirrorCard[]
  total: number
  page: number
  page_size: number
}


export interface CreateMirrorCardParams {
  username: string
  password: string
}

export interface UpdateMirrorCardParams {
  username?: string
  password?: string
  status?: number
  bind_user_id?: number
}

export interface BatchImportParams {
  data: string // 格式: 用户名----密码（每行一条）
}

// 获取镜像卡密列表
export const getMirrorCardList = (params: MirrorCardListParams) => {
  return http.get<ApiResponse<MirrorCardListData>>('/admin/mirror-cards', { params })
}

// 获取镜像卡密详情
export const getMirrorCardDetail = (id: number) => {
  return http.get<ApiResponse<MirrorCard>>(`/admin/mirror-cards/${id}`)
}

// 创建镜像卡密
export const createMirrorCard = (data: CreateMirrorCardParams) => {
  return http.post<ApiResponse<MirrorCard>>('/admin/mirror-cards', data)
}

// 更新镜像卡密
export const updateMirrorCard = (id: number, data: UpdateMirrorCardParams) => {
  return http.put<ApiResponse<MirrorCard>>(`/admin/mirror-cards/${id}`, data)
}

// 删除镜像卡密
export const deleteMirrorCard = (id: number) => {
  return http.delete<ApiResponse<null>>(`/admin/mirror-cards/${id}`)
}

// 批量导入镜像卡密
export const batchImportMirrorCards = (data: BatchImportParams) => {
  return http.post<ApiResponse<any>>('/admin/mirror-cards/batch-import', data)
}

