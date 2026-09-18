import { http } from '@/utils/http'

export interface CursorSmsQueryResult {
  account: string
  status: 'received' | 'waiting' | 'error'
  code: string
  message: string
  expires_at: string
}

export interface CursorMailItem {
  id: string
  subject: string
  from: string
  received_at: string
  preview: string
  mailbox: string
  code: string
  body: string
  html_body: string
  seq_num: number
  folder: string
  view_href?: string
}

export interface CursorMailQueryResult {
  account: string
  email: string
  source: string
  inbox: CursorMailItem[]
  junk: CursorMailItem[]
  message: string
}

export interface CursorMailDetail {
  subject?: string
  from?: string
  received_at?: string
  body: string
  html_body: string
  code: string
}

const MAIL_TIMEOUT = 60000

// queryCursorSms 查询 Cursor 账号短信验证码
// raw 为地址栏 ? 后面的原始整串内容（account----pass），不在前端拆分，交由后端统一处理
export function queryCursorSms(raw: string): Promise<{ data: CursorSmsQueryResult }> {
  return http.get('/sms/cursor/query', { params: { q: raw } })
}

export function queryCursorEmail(raw: string): Promise<{ data: CursorMailQueryResult }> {
  return http.get('/email/cursor/query', { params: { q: raw }, timeout: MAIL_TIMEOUT })
}

export function queryCursorEmailDetail(
  raw: string,
  mail: Pick<CursorMailItem, 'id' | 'folder' | 'seq_num'>
): Promise<{ data: CursorMailDetail }> {
  return http.get('/email/cursor/detail', {
    params: {
      q: raw,
      message_id: mail.id || undefined,
      folder: mail.folder || undefined,
      seq_num: mail.seq_num || undefined,
    },
    timeout: MAIL_TIMEOUT,
  })
}

export function queryCursorRecoveryMail(raw: string): Promise<{ data: CursorMailQueryResult }> {
  return http.get('/recovery-mail/cursor/query', { params: { q: raw }, timeout: MAIL_TIMEOUT })
}

export function queryCursorRecoveryMailDetail(
  raw: string,
  mailId: string
): Promise<{ data: CursorMailDetail }> {
  return http.get('/recovery-mail/cursor/detail', {
    params: { q: raw, mail_id: mailId },
    timeout: MAIL_TIMEOUT,
  })
}
