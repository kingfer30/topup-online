import http from '@/utils/http'
import type { ApiResponse } from '@/types'

// 菜单数据类型
export interface Menu {
  id: number
  parent_id: number
  title: string
  key: string
  path?: string
  icon?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
  children?: Menu[]
}

// 创建/更新菜单请求
export interface MenuRequest {
  parent_id: number
  title: string
  key: string
  path?: string
  icon?: string
  sort: number
  status: number
}

// 创建卡密菜单请求
export interface CardMenuRequest {
  category: string
  menuName: string
  icon?: string
  sort: number
}

// 获取菜单树
export const getMenuTree = (): Promise<ApiResponse<Menu[]>> => {
  return http.get('/admin/menus/tree') as Promise<ApiResponse<Menu[]>>
}

// 获取所有菜单（扁平列表）
export const getAllMenus = (): Promise<ApiResponse<Menu[]>> => {
  return http.get('/admin/menus') as Promise<ApiResponse<Menu[]>>
}

// 获取菜单详情
export const getMenuById = (id: number): Promise<ApiResponse<Menu>> => {
  return http.get(`/admin/menus/${id}`) as Promise<ApiResponse<Menu>>
}

// 创建菜单
export const createMenu = (data: MenuRequest): Promise<ApiResponse<Menu>> => {
  return http.post('/admin/menus', data) as Promise<ApiResponse<Menu>>
}

// 更新菜单
export const updateMenu = (id: number, data: MenuRequest): Promise<ApiResponse<Menu>> => {
  return http.put(`/admin/menus/${id}`, data) as Promise<ApiResponse<Menu>>
}

// 删除菜单
export const deleteMenu = (id: number): Promise<ApiResponse> => {
  return http.delete(`/admin/menus/${id}`) as Promise<ApiResponse>
}

// 获取子菜单
export const getMenusByParentId = (parentId: number): Promise<ApiResponse<Menu[]>> => {
  return http.get('/admin/menus/children', {
    params: { parent_id: parentId },
  }) as Promise<ApiResponse<Menu[]>>
}

// 创建卡密菜单（父菜单+3个子菜单）
export const createCardMenu = (data: CardMenuRequest): Promise<ApiResponse> => {
  return http.post('/admin/menus/card-menu', data) as Promise<ApiResponse>
}

