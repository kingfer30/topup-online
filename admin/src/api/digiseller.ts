import http from '@/utils/http'
import type { ApiResponse } from '@/types'

// Digiseller订阅类型售价配置
export interface DigisellerPrice {
  id?: number
  subscription_type: string
  price: number
  created_at?: string
  updated_at?: string
}

// 获取所有订阅类型售价配置
export const getDigisellerPrices = (): Promise<ApiResponse<DigisellerPrice[]>> => {
  return http.get('/admin/digiseller/prices') as Promise<ApiResponse<DigisellerPrice[]>>
}

// 新增或更新某订阅类型的今日售价
export const upsertDigisellerPrice = (data: { subscription_type: string; price: number }): Promise<ApiResponse> => {
  return http.post('/admin/digiseller/prices', data) as Promise<ApiResponse>
}

export interface DigisellerOrder {
  id: number
  inv: number
  unique_code: string
  id_goods: number
  amount: number
  type_curr: string
  amount_usd: number
  profit: string
  date_pay?: string | null
  email: string
  agent_id: number
  agent_percent: number
  cnt_goods: number
  promo_code: string
  bonus_code: string
  cart_uid: string
  uc_state: number
  uc_date_check?: string | null
  uc_date_delivery?: string | null
  uc_date_confirmed?: string | null
  uc_date_refuted?: string | null
  options_json: string
  created_at?: string
  updated_at?: string
}

export interface DigisellerOrderListParams {
  page?: number
  page_size?: number
  keyword?: string
  uc_state?: number | null
}

export interface DigisellerOrderListResponse {
  list: DigisellerOrder[]
  total: number
  page: number
  page_size: number
}

export interface DigisellerOrderSyncResult {
  saved: number
  failed: number
  total_rows: number
  pages: number
  date_start: string
  date_finish: string
}

export const syncDigisellerOrders = (): Promise<ApiResponse<DigisellerOrderSyncResult>> => {
  return http.post('/admin/digiseller/orders/sync') as Promise<ApiResponse<DigisellerOrderSyncResult>>
}

export const getDigisellerOrders = (
  params: DigisellerOrderListParams
): Promise<ApiResponse<DigisellerOrderListResponse>> => {
  const query: Record<string, string | number> = {}
  if (params.page) query.page = params.page
  if (params.page_size) query.page_size = params.page_size
  if (params.keyword) query.keyword = params.keyword
  if (params.uc_state != null) query.uc_state = params.uc_state
  return http.get('/admin/digiseller/orders', { params: query }) as Promise<ApiResponse<DigisellerOrderListResponse>>
}
