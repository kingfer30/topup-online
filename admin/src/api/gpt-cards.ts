import { http } from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface GptCard {
  id: number
  supplier: string
  key: string
  gpt_mail?: string
  buyer?: string
  status: number // -1作废 1待使用 2已使用
  import_time?: number
  use_time?: number
  sub_start_time?: number
  sub_end_time?: number
  sub_result?: string
  created_at?: string
  updated_at?: string
}

export interface GptCardListParams {
  page?: number
  page_size?: number
  supplier?: string
  status?: number
  keyword?: string
}

export interface GptCardListResponse {
  list: GptCard[]
  total: number
}

export interface BatchImportParams {
  supplier: string
  keys: string[]
}

export interface UpdateGptCardParams {
  gpt_mail?: string
  buyer?: string
  status?: number
  sub_result?: string
}

export const getSuppliers = (): Promise<ApiResponse<string[]>> => {
  return http.get('/admin/gpt-cards/suppliers') as Promise<ApiResponse<string[]>>
}

export const getGptCardList = (params: GptCardListParams): Promise<ApiResponse<GptCardListResponse>> => {
  return http.get('/admin/gpt-cards', { params }) as Promise<ApiResponse<GptCardListResponse>>
}

export const batchImportGptCards = (data: BatchImportParams): Promise<ApiResponse<{ imported: number }>> => {
  return http.post('/admin/gpt-cards/batch-import', data) as Promise<ApiResponse<{ imported: number }>>
}

export const updateGptCard = (id: number, data: UpdateGptCardParams): Promise<ApiResponse<null>> => {
  return http.put(`/admin/gpt-cards/${id}`, data) as Promise<ApiResponse<null>>
}

export const deleteGptCard = (id: number): Promise<ApiResponse<null>> => {
  return http.delete(`/admin/gpt-cards/${id}`) as Promise<ApiResponse<null>>
}

export const batchDeleteGptCards = (ids: number[]): Promise<ApiResponse<null>> => {
  return http.post('/admin/gpt-cards/batch-delete', { ids }) as Promise<ApiResponse<null>>
}

export interface CheckCardResult {
  id: number
  key: string
  old_status: number
  new_status: number
  changed: boolean
  message: string
}

export interface BatchCheckResponse {
  checked: number
  updated: number
  results: CheckCardResult[]
}

export const batchCheckGptCards = (ids: number[]): Promise<ApiResponse<BatchCheckResponse>> => {
  return http.post('/admin/gpt-cards/batch-check', { ids }) as Promise<ApiResponse<BatchCheckResponse>>
}
