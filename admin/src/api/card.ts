import http from '@/utils/http'
import type { ApiResponse } from '@/types'

// 卡密数据类型
export interface Card {
  id: number
  account: string
  password?: string
  mail_password?: string
  subscription_status: number
  subscription_type?: string
  subscription_time?: number
  subscription_expired_time?: number
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_order_no?: string
  sell_price?: number
  sell_date?: number
  sell_to?: string
  sell_order_no?: string
  sell_status: number
  account_type: number
  status: number
  api_key?: string
  token?: string
  '2fa'?: string
  mail_url?: string
  remark?: string
  created_at?: string
  updated_at?: string
}

// 卡密列表查询参数
export interface CardListParams {
  category: string // 卡密类别
  type?: string // all/unsold/sold
  page?: number
  page_size?: number
  keyword?: string
}

// 卡密列表响应
export interface CardListResponse {
  list: Card[]
  total: number
  page: number
  page_size: number
}

// 创建/更新卡密请求
export interface CardRequest {
  account: string
  password?: string
  mail_password?: string
  subscription_status?: number
  subscription_type?: string
  subscription_time?: number
  subscription_expired_time?: number
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_order_no?: string
  sell_price?: number
  sell_date?: number
  sell_to?: string
  sell_order_no?: string
  sell_status?: number
  account_type?: number
  status?: number
  api_key?: string
  token?: string
  '2fa'?: string
  mail_url?: string
  remark?: string
}

// 获取卡密列表
export const getCardList = (params: CardListParams): Promise<ApiResponse<CardListResponse>> => {
  return http.get('/admin/cards', { params }) as Promise<ApiResponse<CardListResponse>>
}

// 获取卡密详情
export const getCardById = (
  category: string,
  id: number
): Promise<ApiResponse<Card>> => {
  return http.get(`/admin/cards/${id}`, {
    params: { category },
  }) as Promise<ApiResponse<Card>>
}

// 创建卡密
export const createCard = (
  category: string,
  data: CardRequest
): Promise<ApiResponse<Card>> => {
  return http.post('/admin/cards', { ...data, category }) as Promise<ApiResponse<Card>>
}

// 更新卡密
export const updateCard = (
  category: string,
  id: number,
  data: CardRequest
): Promise<ApiResponse<Card>> => {
  return http.put(`/admin/cards/${id}`, { ...data, category }) as Promise<ApiResponse<Card>>
}

// 删除卡密
export const deleteCard = (category: string, id: number): Promise<ApiResponse> => {
  return http.delete(`/admin/cards/${id}`, {
    params: { category },
  }) as Promise<ApiResponse>
}

// 批量导入卡密
export interface BatchImportRequest {
  category: string
  cards: CardRequest[]
}

export const batchImportCards = (data: BatchImportRequest): Promise<ApiResponse> => {
  return http.post('/admin/cards/batch-import', data) as Promise<ApiResponse>
}

// 获取未售出的订阅类型列表
export const getUnsoldSubscriptionTypes = (category: string): Promise<ApiResponse<string[]>> => {
  return http.get('/admin/cards/unsold-subscription-types', {
    params: { category },
  }) as Promise<ApiResponse<string[]>>
}

// 取货请求
export interface PickupRequest {
  category: string
  subscription_type: string
}

// 取货接口
export const pickupCard = (data: PickupRequest): Promise<ApiResponse<Card>> => {
  return http.post('/admin/cards/pickup', data) as Promise<ApiResponse<Card>>
}

// 完成取货请求
export interface CompletePickupRequest {
  category: string
  id: number
  sell_price?: number
  sell_to?: string
}

// 完成取货接口
export const completePickup = (data: CompletePickupRequest): Promise<ApiResponse> => {
  return http.post('/admin/cards/complete-pickup', data) as Promise<ApiResponse>
}

