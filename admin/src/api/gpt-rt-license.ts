import { http } from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface GptRtLicense {
  id: number
  license_key: string
  app_id: string
  customer?: string
  status: number
  expires_at: string
  activated_at?: string
  last_verified_at?: string
  last_using_ip?: string
  used_count: number
  available_count: number
  max_devices: number
  bound_device_count?: number
  created_at?: string
  updated_at?: string
}

export interface GptRtLicenseDevice {
  id: number
  license_id: number
  machine_id: string
  login_ip?: string
  user_agent?: string
  bound_at: string
  created_at?: string
  updated_at?: string
}

export interface GptRtLicenseListParams {
  page?: number
  page_size?: number
  status?: number
  keyword?: string
}

export interface GptRtLicenseListResponse {
  list: GptRtLicense[]
  total: number
}

export interface CreateGptRtLicenseParams {
  customer?: string
  app_id?: string
  months: number
  available_count: number
  max_devices?: number
}

export interface UpdateGptRtLicenseParams {
  customer?: string
  status?: number
  expires_at?: string
  months?: number
  available_count?: number
  add_available_count?: number
  max_devices?: number
}

export const getGptRtLicenseList = (
  params: GptRtLicenseListParams
): Promise<ApiResponse<GptRtLicenseListResponse>> => {
  return http.get('/admin/gpt-rt-licenses', { params }) as Promise<ApiResponse<GptRtLicenseListResponse>>
}

export const getGptRtLicenseDevices = (
  id: number
): Promise<ApiResponse<{ list: GptRtLicenseDevice[] }>> => {
  return http.get(`/admin/gpt-rt-licenses/${id}/devices`) as Promise<ApiResponse<{ list: GptRtLicenseDevice[] }>>
}

export const createGptRtLicense = (
  data: CreateGptRtLicenseParams
): Promise<ApiResponse<GptRtLicense>> => {
  return http.post('/admin/gpt-rt-licenses', data) as Promise<ApiResponse<GptRtLicense>>
}

export const updateGptRtLicense = (
  id: number,
  data: UpdateGptRtLicenseParams
): Promise<ApiResponse<null>> => {
  return http.put(`/admin/gpt-rt-licenses/${id}`, data) as Promise<ApiResponse<null>>
}

export const deleteGptRtLicense = (id: number): Promise<ApiResponse<null>> => {
  return http.delete(`/admin/gpt-rt-licenses/${id}`) as Promise<ApiResponse<null>>
}
