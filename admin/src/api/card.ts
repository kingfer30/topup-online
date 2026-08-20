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
  subscription_credits?: number
  is_check?: number
  check_time?: number
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_by?: string
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
  client_id?: string
  mail_url?: string
  remark?: string
  code_link?: string
  freeze_status?: number
  freeze_time?: number
  freeze_remark?: string
  created_at?: string
  updated_at?: string
}

// 卡密列表查询参数
export interface CardListParams {
  category: string // 卡密类别
  type?: string // all/unsold/sold
  page?: number
  page_size?: number
  keyword?: string // 兼容旧接口
  accounts?: string // 多行/逗号分隔的账号列表
  subscription_type?: string
  subscription_status?: number // 已售列表订阅状态过滤
  sell_to?: string              // 已售列表：出售对方（模糊）
  purchase_by?: string           // 已售列表：卖家名称（模糊）
  is_check?: number            // 检查状态过滤 -1未检查 1检查成功 2检查失败
  // 购买时间筛选：
  // - 传 YYYY-MM-DD：按当天(UTC+8)范围查询
  // - 传 Unix 秒/毫秒、或 "YYYY-MM-DD HH:mm:ss"：后端也会兼容解析
  purchase_date?: string
  subscription_time?: string   // 订阅时间精确查询（保留兼容）
  freeze_status?: number       // 冻结状态过滤 -1未冻结 1已冻结（仅普号列表）
  freeze_time?: string         // 冻结时间筛选 YYYY-MM-DD（仅普号列表，按当天范围）
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
  account?: string
  password?: string
  mail_password?: string
  subscription_status?: number
  subscription_type?: string
  subscription_time?: number
  subscription_expired_time?: number
  purchase_date?: number
  purchase_price?: number
  purchase_from?: string
  purchase_by?: string
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
  client_id?: string
  mail_url?: string
  remark?: string
  code_link?: string
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
  format?: string
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

// 回滚取货请求
export interface RollbackPickupRequest {
  category: string
  id: number
}

// 回滚取货接口（将发货中重置为未出售）
export const rollbackPickup = (data: RollbackPickupRequest): Promise<ApiResponse> => {
  return http.post('/admin/cards/rollback-pickup', data) as Promise<ApiResponse>
}

// 回滚已售接口（将已出售重置为未出售）
export const rollbackSoldCard = (data: RollbackPickupRequest): Promise<ApiResponse> => {
  return http.post('/admin/cards/rollback-sold', data) as Promise<ApiResponse>
}

// 提链结果为 cursor dashboard 且掉订阅时：批量回滚已售并标记订阅 -2、检查成功
export const batchDashboardGotoResolve = (data: {
  category: string
  ids: number[]
}): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-dashboard-goto-resolve', data) as Promise<ApiResponse<number>>
}

// 按订阅类型细分的库存
export interface SubscriptionTypeStat {
  subscription_type: string
  count: number
}

// 控制台统计 - 按卡密类型
export interface CardTypeStat {
  category: string                               // 卡密类型名称
  sold_count: number                             // 今日售出数量
  product_stock_count: number                    // 成品未售库存合计（account_type=2）
  product_stock_by_type: SubscriptionTypeStat[]  // 成品按订阅类型细分
  regular_stock_count: number                    // 普号库存合计（account_type=1，未冻结）
  regular_stock_by_type: SubscriptionTypeStat[]  // 普号按订阅类型细分
  revenue_usd: number                            // 今日收入（美元）
  revenue_cny: number                            // 今日收入（人民币，汇率7）
}

// 获取控制台统计数据，date 格式 YYYY-MM-DD，不传则默认今天
export const getDashboardStats = (date?: string): Promise<ApiResponse<CardTypeStat[]>> => {
  return http.get('/admin/dashboard/stats', { params: date ? { date } : {} }) as Promise<ApiResponse<CardTypeStat[]>>
}

// 批量升级为成品请求
export interface BatchUpgradeRequest {
  category: string
  ids: number[]
  subscription_type?: string
  subscription_time?: number   // Unix 秒
  subscription_remaining_days?: number // 订阅剩余天数，用于计算过期时间
  purchase_price?: number      // 追加金额
  purchase_from?: string
  purchase_date?: number       // Unix 秒
}

// 批量升级为成品接口
export const batchUpgradeToProduct = (data: BatchUpgradeRequest): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-upgrade', data) as Promise<ApiResponse<number>>
}

// 导出卡密（返回全部符合条件的数据）
export const exportCards = (params: CardListParams): Promise<ApiResponse<Card[]>> => {
  return http.get('/admin/cards/export', { params }) as Promise<ApiResponse<Card[]>>
}

// 批量取货请求
export interface BatchPickupRequest {
  category: string
  ids: number[]
  sell_price?: number
  sell_to?: string
}

// 批量取货接口
export const batchPickup = (data: BatchPickupRequest): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-pickup', data) as Promise<ApiResponse<number>>
}

// 批量检查订阅状态请求
export interface BatchCheckRequest {
  category: string
  ids: number[]
}

// 批量检查订阅状态接口
export const batchCheckCards = (data: BatchCheckRequest): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-check', data) as Promise<ApiResponse<number>>
}

// 开启按需付费接口
export const enableOnDemandSpend = (category: string, id: number): Promise<ApiResponse> => {
  return http.post('/admin/cards/enable-on-demand', { category, id }) as Promise<ApiResponse>
}

// 批量开启按需付费接口
export const batchEnableOnDemandSpend = (category: string, ids: number[]): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-enable-on-demand', { category, ids }) as Promise<ApiResponse<number>>
}

// 提链：获取 Cursor Pro 付款链接
export const gotoProUpgrade = (token: string, subscriptionType: string): Promise<ApiResponse<string>> => {
  return http.post('/admin/cards/goto-pro', { token, subscription_type: subscriptionType }) as Promise<ApiResponse<string>>
}

export const halfPriceCheckout = (data: {
  uid: string
  token: string
  tier: string
}): Promise<ApiResponse<string>> => {
  return http.post('/admin/cards/half-price-checkout', data) as Promise<ApiResponse<string>>
}

// 单独更新卡密备注
export const updateCardRemark = (category: string, id: number, remark: string): Promise<ApiResponse> => {
  return http.post('/admin/cards/update-remark', { category, id, remark }) as Promise<ApiResponse>
}

// 批量冻结/解冻请求
export interface BatchFreezeRequest {
  category: string
  ids: number[]
  freeze: number  // 1=冻结 -1=解冻
  remark?: string
}

// 批量冻结/解冻接口
export const batchFreezeCards = (data: BatchFreezeRequest): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-freeze', data) as Promise<ApiResponse<number>>
}

// 批量删除卡密（status=-1 软删）
export const batchDeleteCards = (data: { category: string; ids: number[] }): Promise<ApiResponse<number>> => {
  return http.post('/admin/cards/batch-delete', data) as Promise<ApiResponse<number>>
}

