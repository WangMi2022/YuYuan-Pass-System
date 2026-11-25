import service from '@/utils/request'

export const createCategory = (data) => service({ url: '/assetCategory/create', method: 'post', data })
export const updateCategory = (data) => service({ url: '/assetCategory/update', method: 'put', data })
export const deleteCategory = (params) => service({ url: '/assetCategory/delete', method: 'delete', params })
export const getCategoryList = (params) => service({ url: '/assetCategory/list', method: 'get', params })
