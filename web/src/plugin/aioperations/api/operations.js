import service from '@/utils/request'

export const getAIProviders = () => service({ url: '/ai/providers', method: 'get' })
export const updateAIProviders = (data) => service({ url: '/ai/providers', method: 'put', data })
export const getAIUsageSummary = () => service({ url: '/ai/usage/summary', method: 'get' })
export const getAIInvocations = (params) => service({ url: '/ai/invocations', method: 'get', params })
export const getAIQuotas = () => service({ url: '/ai/quotas', method: 'get' })
export const saveAIQuota = (data) => service({ url: '/ai/quotas', method: 'put', data })
export const getAIPrompts = () => service({ url: '/ai/prompts', method: 'get' })
export const createAIPrompt = (data) => service({ url: '/ai/prompts', method: 'post', data })
export const activateAIPrompt = (data) => service({ url: '/ai/prompts/activate', method: 'put', data })
