import { http } from '@/utils/http'

// 系统初始化相关API

// 检查系统初始化状态
export const checkInitStatus = () => {
  return http.get('/system/init/status')
}

// 测试数据库连接
export const testDBConnection = (data: {
  db_host: string
  db_port: string
  db_name: string
  db_user: string
  db_password: string
}) => {
  return http.post('/system/init/test-db', data)
}

// 初始化系统
export const initializeSystem = (data: {
  db_host: string
  db_port: string
  db_name: string
  db_user: string
  db_password: string
  admin_user: string
  admin_pass: string
  admin_email: string
}) => {
  return http.post('/system/init', data)
}

