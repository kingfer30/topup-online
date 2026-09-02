import { http } from '@/utils/http'

export interface AdminCursorPaySettings {
  billing_name: string
  billing_postal: string
  billing_state: string
  billing_city: string
  billing_line1: string
  billing_country: string
  proxy_scheme: string
  proxy_host: string
  proxy_username: string
  proxy_password_configured: boolean
}

export const getAdminCursorPaySettings = () => {
  return http.get<{ code: number; message: string; data: AdminCursorPaySettings }>('/admin/settings/cursor-pay')
}

export const updateAdminCursorPaySettings = (data: {
  billing_name: string
  billing_postal: string
  billing_state: string
  billing_city: string
  billing_line1: string
  billing_country: string
  proxy_scheme: string
  proxy_host: string
  proxy_username: string
  proxy_password: string
}) => {
  return http.put<{ code: number; message: string }>('/admin/settings/cursor-pay', data)
}
