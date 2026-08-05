import service from '@/utils/request'

export const getWorkSchedules = () => service({
  url: '/workSchedule/list',
  method: 'get',
  donNotShowLoading: true
})

export const createWorkSchedule = (data) => service({
  url: '/workSchedule/create',
  method: 'post',
  data
})

export const updateWorkSchedule = (data) => service({
  url: '/workSchedule/update',
  method: 'put',
  data
})

export const deleteWorkSchedule = (params) => service({
  url: '/workSchedule/delete',
  method: 'delete',
  params
})

export const importLegacyWorkSchedules = (data) => service({
  url: '/workSchedule/import',
  method: 'post',
  data,
  donNotShowLoading: true
})

export const getWorkScheduleNotifications = (params = {}) => service({
  url: '/workSchedule/notifications',
  method: 'get',
  params,
  donNotShowLoading: true
})

export const markWorkScheduleNotificationRead = (data) => service({
  url: '/workSchedule/notifications/read',
  method: 'post',
  data,
  donNotShowLoading: true
})

export const markAllWorkScheduleNotificationsRead = () => service({
  url: '/workSchedule/notifications/readAll',
  method: 'post',
  donNotShowLoading: true
})
