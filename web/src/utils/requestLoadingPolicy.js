/**
 * 页面内容遮罩只允许由明确的 loadingOption 触发。
 * 普通 API 请求仍然正常发送，但反馈交给调用按钮或局部数据区域。
 */
export const shouldUseRequestLoadingOverlay = (config = {}) => (
  config.donNotShowLoading !== true && Boolean(config.loadingOption)
)
