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
