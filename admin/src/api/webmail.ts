import { http } from '@/utils/http'

export interface ApiResp<T> {
  code: number
  message: string
  data: T
}

export interface LqqqMailItem {
  subject: string
  date: string
  mailbox: string
  code: string
  view_href: string
  html_body?: string
  body?: string
}

export interface LqqqMailDetail {
  body: string
  html_body: string
  code: string
}

export interface WebmailMailDetail {
  body: string
  html_body: string
  code: string
}

export interface LqqqFetchData {
  email: string
  inbox: LqqqMailItem[]
  junk: LqqqMailItem[]
}

export interface ToolsvipMailItem {
  subject: string
  date: string
  mailbox: string
  code: string
  html_body: string
  body: string
}

export interface ToolsvipFetchData {
  email: string
  inbox: ToolsvipMailItem[]
  junk: ToolsvipMailItem[]
}

export const fetchLqqqMails = (account_line: string, account_format = '1') =>
  http.post<ApiResp<LqqqFetchData>>('/admin/webmail/lqqq/fetch', { account_line, account_format })

export const fetchLqqqDetail = (
  account_line: string,
  view_href: string,
  account_format = '1',
) =>
  http.post<ApiResp<LqqqMailDetail>>('/admin/webmail/lqqq/detail', {
    account_line,
    view_href,
    account_format,
  })

export const fetchToolsvipMails = (account_line: string, account_format = '1') =>
  http.post<ApiResp<ToolsvipFetchData>>('/admin/webmail/toolsvip/fetch', { account_line, account_format })

export interface LurentoolMailItem {
  subject: string
  from: string
  received_at: string
  preview: string
  body: string
  is_read: boolean
  code: string
  source: string
}

export interface LurentoolFetchData {
  email: string
  ok: boolean
  source: string
  message: string
  mails: LurentoolMailItem[]
}

export const fetchLurentoolMails = (account_line: string, account_format = '1', top = 20) =>
  http.post<ApiResp<LurentoolFetchData>>('/admin/webmail/lurentool/fetch', {
    account_line,
    account_format,
    top,
  })

export interface OuleyxMailItem {
  id: string
  subject: string
  from: string
  received_at: string
  preview: string
  mailbox: string
  code: string
  body?: string
  html_body?: string
}

export interface OuleyxMailDetail {
  subject: string
  from: string
  received_at: string
  body: string
  html_body: string
  code: string
}

export interface OuleyxFetchData {
  email: string
  token: string
  inbox: OuleyxMailItem[]
  junk: OuleyxMailItem[]
}

export const fetchOuleyxMails = (account_line: string, account_format = '1') =>
  http.post<ApiResp<OuleyxFetchData>>('/admin/webmail/ouleyx/fetch', { account_line, account_format })

export const fetchOuleyxDetail = (
  account_line: string,
  mail_id: string,
  token = '',
  account_format = '1',
) =>
  http.post<ApiResp<OuleyxMailDetail>>('/admin/webmail/ouleyx/detail', {
    account_line,
    mail_id,
    token,
    account_format,
  })
