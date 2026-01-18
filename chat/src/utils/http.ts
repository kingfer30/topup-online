import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

// 获取API基础路径
const getBaseURL = () => {
  // 开发环境使用代理，生产环境使用配置的URL
  if (import.meta.env.DEV) {
    return '/api'
  }
  return import.meta.env.VITE_API_BASE_URL || '/api'
}

// 创建axios实例
const service: AxiosInstance = axios.create({
  baseURL: getBaseURL(),
  timeout: 15000, // 请求超时时间
  headers: {
    'Content-Type': 'application/json;charset=UTF-8',
  },
})

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    // 在发送请求之前做些什么
    // 例如：添加token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    // 对请求错误做些什么
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    // 对响应数据做点什么
    const res = response.data
    
    // 根据后端返回的状态码进行处理
    if (res.code !== 200) {
      console.error('响应错误:', res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    
    return res
  },
  (error) => {
    // 对响应错误做点什么
    console.error('响应错误:', error)
    
    let errorMessage = '请求失败'
    
    if (error.response) {
      // 如果后端返回了错误信息，优先使用后端的message
      if (error.response.data && error.response.data.message) {
        errorMessage = error.response.data.message
      } else {
        // 否则根据状态码显示默认消息
        switch (error.response.status) {
          case 401:
            errorMessage = '未授权，请重新登录'
            break
          case 403:
            errorMessage = '拒绝访问'
            break
          case 404:
            errorMessage = '请求地址不存在'
            break
          case 500:
            errorMessage = '服务器内部错误'
            break
          default:
            errorMessage = '请求失败'
        }
      }
    } else if (error.message) {
      errorMessage = error.message
    }
    
    return Promise.reject(new Error(errorMessage))
  }
)

// 导出封装的请求方法
export const http = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.get(url, config)
  },

  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.post(url, data, config)
  },

  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.put(url, data, config)
  },

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.delete(url, config)
  },
}

export default service

