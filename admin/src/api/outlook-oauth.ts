import { http } from '@/utils/http'

export interface OutlookMailItem {
  id?: string
  seq_num: number
  folder: string
  subject: string
  from: string
  received_at: string
  preview: string
  body: string
  html_body: string
  code: string
  is_read: boolean
}

export interface OutlookMailDetail {
  body: string
  html_body: string
  code: string
}

export interface OutlookFetchData {
  email: string
  source?: string
  inbox: OutlookMailItem[]
  junk: OutlookMailItem[]
}

export interface ApiResp<T> {
  code: number
  message: string
  data: T
}

export const fetchOutlookMails = (account_line: string, account_format = '1') =>
  http.post<ApiResp<OutlookFetchData>>('/admin/outlook-oauth/fetch', { account_line, account_format })

export const fetchOutlookDetail = (
  account_line: string,
  folder: string,
  seq_num: number,
  account_format = '1',
  message_id = '',
) =>
  http.post<ApiResp<OutlookMailDetail>>('/admin/outlook-oauth/detail', {
    account_line,
    folder,
    seq_num,
    account_format,
    message_id,
  })
