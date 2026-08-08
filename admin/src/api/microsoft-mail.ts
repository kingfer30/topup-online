import http from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface MicrosoftMail {
  id: number
  account: string
  password?: string
  is_check?: number
  check_time?: number
  next_check_time?: number
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_by?: string
  sell_price?: number
  sell_date?: number
  sell_to?: string
  sell_order_no?: string
  sell_status: number
  status: number
  token?: string
  '2fa'?: string
  client_id?: string
  mail_url?: string
  remark?: string
  freeze_status?: number
  freeze_time?: number
  freeze_remark?: string
  account_card_id?: number
  account_card_table?: string
  created_at?: string
  updated_at?: string
}

export interface MicrosoftMailListParams {
  type?: string
  page?: number
  page_size?: number
  keyword?: string
  accounts?: string
  is_check?: number
  purchase_date?: string
  sell_to?: string
  purchase_by?: string
}

export interface MicrosoftMailListResponse {
  list: MicrosoftMail[]
  total: number
  page: number
  page_size: number
}

export interface MicrosoftMailRequest {
  account?: string
  password?: string
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_by?: string
  sell_price?: number
  sell_date?: number
  sell_to?: string
  sell_order_no?: string
  sell_status?: number
  status?: number
  token?: string
  '2fa'?: string
  client_id?: string
  mail_url?: string
  remark?: string
  account_card_id?: number | null
  account_card_table?: string
}

export const getMicrosoftMailList = (
  params: MicrosoftMailListParams
): Promise<ApiResponse<MicrosoftMailListResponse>> => {
  return http.get('/admin/microsoft-mails', { params }) as Promise<ApiResponse<MicrosoftMailListResponse>>
}

export const getMicrosoftMailById = (id: number): Promise<ApiResponse<MicrosoftMail>> => {
  return http.get(`/admin/microsoft-mails/${id}`) as Promise<ApiResponse<MicrosoftMail>>
}

export const createMicrosoftMail = (
  data: MicrosoftMailRequest
): Promise<ApiResponse<MicrosoftMail>> => {
  return http.post('/admin/microsoft-mails', data) as Promise<ApiResponse<MicrosoftMail>>
}

export const updateMicrosoftMail = (
  id: number,
  data: MicrosoftMailRequest
): Promise<ApiResponse<MicrosoftMail>> => {
  return http.put(`/admin/microsoft-mails/${id}`, data) as Promise<ApiResponse<MicrosoftMail>>
}

export const deleteMicrosoftMail = (id: number): Promise<ApiResponse> => {
  return http.delete(`/admin/microsoft-mails/${id}`) as Promise<ApiResponse>
}

export const batchImportMicrosoftMails = (
  mails: MicrosoftMailRequest[]
): Promise<ApiResponse> => {
  return http.post('/admin/microsoft-mails/batch-import', { mails }) as Promise<ApiResponse>
}

export const exportMicrosoftMails = (
  params: MicrosoftMailListParams
): Promise<ApiResponse<MicrosoftMail[]>> => {
  return http.get('/admin/microsoft-mails/export', { params }) as Promise<ApiResponse<MicrosoftMail[]>>
}

export const pickupMicrosoftMail = (format?: string): Promise<ApiResponse<MicrosoftMail>> => {
  return http.post('/admin/microsoft-mails/pickup', { format }) as Promise<ApiResponse<MicrosoftMail>>
}

export const completeMicrosoftMailPickup = (data: {
  id: number
  sell_price?: number
  sell_to?: string
}): Promise<ApiResponse> => {
  return http.post('/admin/microsoft-mails/complete-pickup', data) as Promise<ApiResponse>
}

export const rollbackMicrosoftMailPickup = (id: number): Promise<ApiResponse> => {
  return http.post('/admin/microsoft-mails/rollback-pickup', { id }) as Promise<ApiResponse>
}

export const rollbackMicrosoftMailSold = (id: number): Promise<ApiResponse> => {
  return http.post('/admin/microsoft-mails/rollback-sold', { id }) as Promise<ApiResponse>
}

export const batchPickupMicrosoftMails = (data: {
  ids: number[]
  sell_price?: number
  sell_to?: string
}): Promise<ApiResponse<number>> => {
  return http.post('/admin/microsoft-mails/batch-pickup', data) as Promise<ApiResponse<number>>
}

export const batchCheckMicrosoftMails = (ids: number[]): Promise<ApiResponse<number>> => {
  return http.post('/admin/microsoft-mails/batch-check', { ids }) as Promise<ApiResponse<number>>
}

export const batchDeleteMicrosoftMails = (ids: number[]): Promise<ApiResponse<number>> => {
  return http.post('/admin/microsoft-mails/batch-delete', { ids }) as Promise<ApiResponse<number>>
}

export const updateMicrosoftMailRemark = (id: number, remark: string): Promise<ApiResponse> => {
  return http.post('/admin/microsoft-mails/update-remark', { id, remark }) as Promise<ApiResponse>
}
