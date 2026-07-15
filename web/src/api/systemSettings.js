import service from '@/utils/request'

export const getCurrentLoginBackground = () => {
  return service({
    url: '/appearance/login-background',
    method: 'get',
    donNotShowLoading: true
  })
}

export const getLoginBackgrounds = () => {
  return service({
    url: '/appearance/login-backgrounds',
    method: 'get'
  })
}

export const createLoginBackground = (data) => {
  return service({
    url: '/appearance/login-background',
    method: 'post',
    data
  })
}

export const activateLoginBackground = (data) => {
  return service({
    url: '/appearance/login-background/activate',
    method: 'put',
    data
  })
}

export const deleteLoginBackground = (params) => {
  return service({
    url: '/appearance/login-background',
    method: 'delete',
    params
  })
}
