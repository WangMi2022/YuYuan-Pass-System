import service from '@/utils/request'

export const createAsset = (data) => service({ url: '/asset/create', method: 'post', data })
export const updateAsset = (data) => service({ url: '/asset/update', method: 'put', data })
export const deleteAsset = (params) => service({ url: '/asset/delete', method: 'delete', params })
export const getAssetDetail = (params) => service({ url: '/asset/detail', method: 'get', params })
export const getAssetList = (params) => service({ url: '/asset/list', method: 'get', params })
export const getAssetDashboard = () => service({ url: '/asset/dashboard', method: 'get' })
export const getCategoryOptions = () => service({ url: '/asset/categoryOptions', method: 'get' })

export const uploadAssetPhoto = (file) => {
  const data = new FormData()
  data.append('file', file)
  return service({
    url: '/asset/uploadPhoto',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export const deleteAssetPhoto = (params) => service({ url: '/asset/deletePhoto', method: 'delete', params })
