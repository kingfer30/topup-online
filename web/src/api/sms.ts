import { http } from '@/utils/http'

export interface CursorSmsQueryResult {
  account: string
  status: 'received' | 'waiting' | 'error'
  code: string
  message: string
  expires_at: string
}

// queryCursorSms 查询 Cursor 账号短信验证码
// raw 为地址栏 ? 后面的原始整串内容（account----pass），不在前端拆分，交由后端统一处理
export function queryCursorSms(raw: string): Promise<{ data: CursorSmsQueryResult }> {
  return http.get('/sms/cursor/query', { params: { q: raw } })
}
