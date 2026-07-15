import service from '@/utils/request'

export const getDocumentList = (params) => service({ url: '/document/list', method: 'get', params })
export const getDocumentDetail = (params) => service({ url: '/document/detail', method: 'get', params })
export const updateDocumentContent = (data) => service({ url: '/document/updateContent', method: 'put', data })
export const deleteDocument = (params) => service({ url: '/document/delete', method: 'delete', params })
export const downloadDocumentFile = (params) => service({ url: '/document/file', method: 'get', params, responseType: 'arraybuffer', donNotShowLoading: true })

export const uploadDocument = (file) => {
  const data = new FormData()
  data.append('file', file)
  return service({
    url: '/document/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
