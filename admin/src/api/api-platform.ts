import http from '@/utils/http'
import type { ApiResponse } from '@/types'

export interface ApiPlatform {
  id: number
  base_url: string
  api_key?: string
  api_key_mask?: string
  account: string
  password?: string
  password_mask?: string
  remark: string
  used_amount: number
  settled_amount: number
  unsettled_amount: number
  created_at?: string
  updated_at?: string
}

export interface ApiPlatformChannel {
  id: number
  platform_id: number
  channel_id: number
  name: string
  filter_name: string
  models: string
  status: number
  status_label: string
  used_quota: number
  used_amount: number
  settled_amount: number
  unsettled_amount: number
  cost_price: number | null
  sale_price: number | null
  last_sync_at?: string | null
  last_test_result?: string
  in_pending_bill?: boolean
  created_at?: string
  updated_at?: string
}

export interface PulledChannel {
  channel_id: number
  name: string
  models: string
  status: number
  status_label: string
  used_quota: number
  used_amount: number
  cost_price?: number | null
  sale_price?: number | null
}

export interface ApiPlatformBillItem {
  id: number
  bill_id: number
  channel_row_id: number
  channel_id: number
  name: string
  filter_name: string
  used_amount: number
  settled_amount: number
  pay_quota: number
  cost_price: number
  sale_price: number
  amount_cny: number
}

export interface ApiPlatformBill {
  id: number
  platform_id: number
  bill_no: string
  status: number
  status_label?: string
  pay_quota: number
  amount_cny: number
  paid_at?: string | null
  created_at?: string
  channel_count?: number
  items?: ApiPlatformBillItem[]
}

export interface ApiPlatformSettings {
  sync_enabled: boolean
  sync_minutes: number
  discover_enabled: boolean
  discover_minutes: number
  test_models: string[]
}

export const listApiPlatforms = () => http.get('/admin/api-platforms') as Promise<ApiResponse<ApiPlatform[]>>

export const getApiPlatform = (id: number) =>
  http.get(`/admin/api-platforms/${id}`) as Promise<ApiResponse<ApiPlatform>>

export const createApiPlatform = (data: Partial<ApiPlatform>) =>
  http.post('/admin/api-platforms', data) as Promise<ApiResponse<ApiPlatform>>

export const updateApiPlatform = (id: number, data: Partial<ApiPlatform>) =>
  http.put(`/admin/api-platforms/${id}`, data) as Promise<ApiResponse>

export const deleteApiPlatform = (id: number) =>
  http.delete(`/admin/api-platforms/${id}`) as Promise<ApiResponse>

export const listApiPlatformChannels = (id: number) =>
  http.get(`/admin/api-platforms/${id}/channels`) as Promise<ApiResponse<ApiPlatformChannel[]>>

export const pullApiPlatformChannels = (
  id: number,
  data: { min_id?: number; max_id?: number; keyword?: string; status?: string }
) => http.post(`/admin/api-platforms/${id}/channels/pull`, data) as Promise<ApiResponse<PulledChannel[]>>

export const saveApiPlatformChannels = (
  id: number,
  data: { filter_name: string; cost_price?: number | null; sale_price?: number | null; channels: PulledChannel[] }
) => http.post(`/admin/api-platforms/${id}/channels/save`, data) as Promise<ApiResponse>

export const syncApiPlatformChannels = (id: number) =>
  http.post(`/admin/api-platforms/${id}/channels/sync`) as Promise<ApiResponse<{ updated: number; failed: number }>>

export const updateApiPlatformChannelMoney = (
  id: number,
  channelRowId: number,
  data: { cost_price?: number; sale_price?: number; settled_amount?: number }
) => http.put(`/admin/api-platforms/${id}/channels/${channelRowId}`, data) as Promise<ApiResponse>

export const batchUpdateApiPlatformChannelPrices = (
  id: number,
  data: { channel_row_ids: number[]; cost_price?: number | null; sale_price?: number | null }
) => http.post(`/admin/api-platforms/${id}/channels/prices`, data) as Promise<ApiResponse<{ updated: number }>>

export const testApiPlatformChannel = (id: number, channelRowId: number, model: string) =>
  http.post(`/admin/api-platforms/${id}/channels/${channelRowId}/test`, { model }) as Promise<
    ApiResponse<{ status_code: number; message: string; model: string }>
  >

export const updateApiPlatformChannelModels = (id: number, channelRowId: number, models: string[]) =>
  http.put(`/admin/api-platforms/${id}/channels/${channelRowId}/models`, { models }) as Promise<
    ApiResponse<{ models: string }>
  >

export const updateApiPlatformChannelStatus = (id: number, channelRowId: number, status: number) =>
  http.post(`/admin/api-platforms/${id}/channels/${channelRowId}/status`, { status }) as Promise<ApiResponse>

export const listApiPlatformBills = (id: number) =>
  http.get(`/admin/api-platforms/${id}/bills`) as Promise<ApiResponse<ApiPlatformBill[]>>

export const getApiPlatformBill = (id: number, billId: number) =>
  http.get(`/admin/api-platforms/${id}/bills/${billId}`) as Promise<
    ApiResponse<{
      bill: ApiPlatformBill
      status_label: string
      platform: { id: number; base_url: string; account: string; remark: string }
    }>
  >

export const createApiPlatformBill = (id: number, channelRowIds: number[]) =>
  http.post(`/admin/api-platforms/${id}/bills`, { channel_row_ids: channelRowIds }) as Promise<ApiResponse<ApiPlatformBill>>

export const payApiPlatformBill = (id: number, billId: number) =>
  http.post(`/admin/api-platforms/${id}/bills/${billId}/pay`) as Promise<ApiResponse>

export const voidApiPlatformBill = (id: number, billId: number) =>
  http.post(`/admin/api-platforms/${id}/bills/${billId}/void`) as Promise<ApiResponse>

export const getAdminApiPlatformSettings = () =>
  http.get('/admin/settings/api-platform') as Promise<ApiResponse<ApiPlatformSettings>>

export const updateAdminApiPlatformSettings = (data: ApiPlatformSettings) =>
  http.put('/admin/settings/api-platform', data) as Promise<ApiResponse<ApiPlatformSettings>>
