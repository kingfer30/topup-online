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

function cursorLookupParams(raw: string, token?: string): Record<string, string> {
  if (token) return { t: token }
  return { q: raw }
}

// queryCursorSms 查询 Cursor 账号短信验证码
export function queryCursorSms(raw: string, token?: string): Promise<{ data: CursorSmsQueryResult }> {
  return http.get('/sms/cursor/query', { params: cursorLookupParams(raw, token) })
}

export function queryCursorEmail(raw: string, token?: string): Promise<{ data: CursorMailQueryResult }> {
  return http.get('/email/cursor/query', { params: cursorLookupParams(raw, token), timeout: MAIL_TIMEOUT })
}

export function queryCursorEmailDetail(
  raw: string,
  mail: Pick<CursorMailItem, 'id' | 'folder' | 'seq_num'>,
  token?: string
): Promise<{ data: CursorMailDetail }> {
  return http.get('/email/cursor/detail', {
    params: {
      ...cursorLookupParams(raw, token),
      message_id: mail.id || undefined,
      folder: mail.folder || undefined,
      seq_num: mail.seq_num || undefined,
    },
    timeout: MAIL_TIMEOUT,
  })
}

export function queryCursorRecoveryMail(raw: string, token?: string): Promise<{ data: CursorMailQueryResult }> {
  return http.get('/recovery-mail/cursor/query', { params: cursorLookupParams(raw, token), timeout: MAIL_TIMEOUT })
}

export function queryCursorRecoveryMailDetail(
  raw: string,
  mailId: string,
  token?: string
): Promise<{ data: CursorMailDetail }> {
  return http.get('/recovery-mail/cursor/detail', {
    params: { ...cursorLookupParams(raw, token), mail_id: mailId },
    timeout: MAIL_TIMEOUT,
  })
}
