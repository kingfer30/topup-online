import { http } from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface GptCdk {
  id: number
  key: string
  gpt_mail?: string
  buyer?: string
  sell_status: number // -1作废 1待售 2已售
  expire_time?: number
  use_status: number // 1未使用 2占用中 3已使用
  use_time?: number
  card_id?: number
  sub_start_time?: number
  sub_end_time?: number
  sub_result?: string
  ip_addr?: string
  device_info?: string
  created_at?: string
  updated_at?: string
}

export interface GptCdkListParams {
  page?: number
  page_size?: number
  sell_status?: number
  use_status?: number
  keyword?: string
}

export interface GptCdkListResponse {
  list: GptCdk[]
  total: number
}

export interface BatchGenerateParams {
  count: number
  expire_time?: number
}

export interface UpdateGptCdkParams {
  gpt_mail?: string
  buyer?: string
  sell_status?: number
  use_status?: number
  card_id?: number
  sub_result?: string
  ip_addr?: string
  device_info?: string
  expire_time?: number
}

export const getGptCdkList = (params: GptCdkListParams): Promise<ApiResponse<GptCdkListResponse>> => {
  return http.get('/admin/gpt-cdk', { params }) as Promise<ApiResponse<GptCdkListResponse>>
}

export const batchGenerateGptCdk = (data: BatchGenerateParams): Promise<ApiResponse<{ generated: number }>> => {
  return http.post('/admin/gpt-cdk/batch-generate', data) as Promise<ApiResponse<{ generated: number }>>
}

export const updateGptCdk = (id: number, data: UpdateGptCdkParams): Promise<ApiResponse<null>> => {
  return http.put(`/admin/gpt-cdk/${id}`, data) as Promise<ApiResponse<null>>
}

export const deleteGptCdk = (id: number): Promise<ApiResponse<null>> => {
  return http.delete(`/admin/gpt-cdk/${id}`) as Promise<ApiResponse<null>>
}

export const batchDeleteGptCdks = (ids: number[]): Promise<ApiResponse<null>> => {
  return http.post('/admin/gpt-cdk/batch-delete', { ids }) as Promise<ApiResponse<null>>
}
