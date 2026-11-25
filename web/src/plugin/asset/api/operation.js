import service from '@/utils/request'

export const createAssetOperation = (data) => service({ url: '/assetOperation/create', method: 'post', data })
export const updateAssetOperation = (data) => service({ url: '/assetOperation/update', method: 'put', data })
export const submitAssetOperation = (params) => service({ url: '/assetOperation/submit', method: 'put', params })
export const deleteAssetOperation = (params) => service({ url: '/assetOperation/delete', method: 'delete', params })
export const getAssetOperationDetail = (params) => service({ url: '/assetOperation/detail', method: 'get', params })
export const getAssetOperationList = (params) => service({ url: '/assetOperation/list', method: 'get', params })
export const getOperationAssetOptions = (params) => service({ url: '/assetOperation/assetOptions', method: 'get', params })
