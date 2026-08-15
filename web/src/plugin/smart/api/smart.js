import service from '@/utils/request'

export const queryCopilot = (data) => service({ url: '/smart/copilot/query', method: 'post', data })
export const getCopilotSessions = () => service({ url: '/smart/copilot/sessions', method: 'get' })
export const getCopilotSession = (params) => service({ url: '/smart/copilot/session', method: 'get', params })
export const deleteCopilotSession = (params) => service({ url: '/smart/copilot/session', method: 'delete', params })
export const getCopilotTools = () => service({ url: '/smart/copilot/tools', method: 'get' })

export const getTodaySmartReport = () => service({ url: '/smartReport/today', method: 'get' })
export const getSmartReports = (params) => service({ url: '/smartReport/list', method: 'get', params })
export const getSmartReport = (params) => service({ url: '/smartReport/detail', method: 'get', params })
export const generateSmartReport = () => service({ url: '/smartReport/generate', method: 'post' })
export const getSmartReportSubscription = () => service({ url: '/smartReport/subscription', method: 'get' })
export const saveSmartReportSubscription = (data) => service({ url: '/smartReport/subscription', method: 'put', data })
export const getSmartReportDeliveries = () => service({ url: '/smartReport/deliveries', method: 'get' })

export const extractAnnouncementSchedule = (data) => service({ url: '/smart/announcement/extract', method: 'post', data })
export const createOperationDraft = (data) => service({ url: '/smart/operation/draft', method: 'post', data })
export const getOperationAssetCandidates = (params) => service({ url: '/smart/operation/assets', method: 'get', params })
export const getSmartDrafts = (params) => service({ url: '/smart/drafts', method: 'get', params })
export const acceptSmartDraft = (data) => service({ url: '/smart/draft/accept', method: 'post', data })
