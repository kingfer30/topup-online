import { http } from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface AdConfig {
  id: number
  title: string
  image: string
  link: string
  position: 'left' | 'right' | string
  buyer: string
  click_count: number
  sort: number
  status: number
  start_time: string
  end_time: string
  created_at?: string
  updated_at?: string
}

export interface AdConfigListParams {
  page?: number
  page_size?: number
  keyword?: string
  position?: string
}

export interface AdConfigListResponse {
  list: AdConfig[]
  total: number
  page: number
  page_size: number
}

export interface AdConfigPayload {
  title: string
  image: string
  link: string
  position: string
  buyer?: string
  sort?: number
  status?: number
  start_time: string
  end_time: string
}

export const getAdConfigList = (
  params: AdConfigListParams
): Promise<ApiResponse<AdConfigListResponse>> => {
  return http.get('/admin/ad-configs', { params }) as Promise<ApiResponse<AdConfigListResponse>>
}

export const createAdConfig = (data: AdConfigPayload): Promise<ApiResponse<AdConfig>> => {
  return http.post('/admin/ad-configs', data) as Promise<ApiResponse<AdConfig>>
}

export const updateAdConfig = (
  id: number,
  data: AdConfigPayload
): Promise<ApiResponse<AdConfig>> => {
  return http.put(`/admin/ad-configs/${id}`, data) as Promise<ApiResponse<AdConfig>>
}

export const deleteAdConfig = (id: number): Promise<ApiResponse<null>> => {
  return http.delete(`/admin/ad-configs/${id}`) as Promise<ApiResponse<null>>
}

export const uploadAdImage = (
  file: File
): Promise<ApiResponse<{ url: string; width: number; height: number }>> => {
  const form = new FormData()
  form.append('file', file)
  return http.post('/admin/upload/ad-image', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }) as Promise<ApiResponse<{ url: string; width: number; height: number }>>
}
