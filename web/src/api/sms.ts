import { http } from '@/utils/http'

export interface CursorSmsQueryResult {
  status: 'received' | 'waiting' | 'error'
  code: string
  message: string
  expires_at: string
}

// queryCursorSms 查询 Cursor 账号短信验证码
export function queryCursorSms(account: string, pass: string): Promise<{ data: CursorSmsQueryResult }> {
  return http.get('/sms/cursor/query', { params: { account, pass } })
}
