import { http } from '@/utils/http'

export interface VerifyCdkResponse {
  valid: boolean
  cdk_id: number
  expire_time: number | null
}

export interface StartTopupRequest {
  cdk_key: string
  user_email: string
  user_gpt_token: string
  full_auth_data?: string
  supplier?: string
}

export interface StartTopupResponse {
  task_id: number
}

export interface TopupTaskStatus {
  task_id: number
  status: number // 0待处理 1处理中 2成功 3失败
  message: string
  created_at: string
}

// verifyCdk 验证CDK有效性
export function verifyCdk(cdkKey: string): Promise<{ data: VerifyCdkResponse }> {
  return http.post('/topup/verify-cdk', { cdk_key: cdkKey })
}

// startTopup 发起充值
export function startTopup(data: StartTopupRequest): Promise<{ data: StartTopupResponse }> {
  return http.post('/topup/start', data)
}

// getTopupTaskStatus 查询充值任务状态
export function getTopupTaskStatus(taskId: number): Promise<{ data: TopupTaskStatus }> {
  return http.get(`/topup/task/${taskId}`)
}
