import service from '@/utils/request'

export const createSite = (data) => service({ url: '/site/create', method: 'post', data })
export const updateSite = (data) => service({ url: '/site/update', method: 'put', data })
export const deleteSite = (params) => service({ url: '/site/delete', method: 'delete', params })
export const getSiteList = (params) => service({ url: '/site/list', method: 'get', params })
export const getSiteDetail = (params) => service({ url: '/site/detail', method: 'get', params })
export const getSiteCategories = () => service({ url: '/site/categories', method: 'get' })
export const visitSite = (params) => service({ url: '/site/visit', method: 'post', params, donNotShowLoading: true })
