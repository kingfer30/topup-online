import http from '@/utils/http'

// 管理员相关API

// 管理员登录
export const adminLogin = (data: {
  username: string
  password: string // MD5加密后的密码
}) => {
  return http.post('/admin/login', data)
}

// 获取管理员信息
export const getAdminInfo = () => {
  return http.get('/admin/info')
}

// 退出登录
export const adminLogout = () => {
  return http.post('/admin/logout')
}

// 修改密码
export const changePassword = (data: {
  old_password: string
  new_password: string
}) => {
  return http.post('/admin/change-password', data)
}

