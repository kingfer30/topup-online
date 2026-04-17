import { http } from '@/utils/http'

export interface AdminAISettings {
  model_name: string
  base_url: string
  api_key_configured: boolean
}

export const getAdminAISettings = () => {
  return http.get<{ code: number; message: string; data: AdminAISettings }>('/admin/settings/ai')
}

export const updateAdminAISettings = (data: {
  model_name: string
  base_url: string
  api_key: string
}) => {
  return http.put<{ code: number; message: string }>('/admin/settings/ai', data)
}

export const adminAITranslate = (data: {
  text: string
  source_lang: string
  target_lang: string
}) => {
  return http.post<{ code: number; message: string; data?: { translation: string } }>(
    '/admin/ai/translate',
    data
  )
}
