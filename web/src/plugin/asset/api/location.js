import service from '@/utils/request'

export const createLocation = (data) => service({ url: '/assetLocation/create', method: 'post', data })
export const updateLocation = (data) => service({ url: '/assetLocation/update', method: 'put', data })
export const deleteLocation = (params) => service({ url: '/assetLocation/delete', method: 'delete', params })
export const getLocationList = (params) => service({ url: '/assetLocation/list', method: 'get', params })
export const getLocationOptions = (params) => service({ url: '/assetLocation/options', method: 'get', params })
