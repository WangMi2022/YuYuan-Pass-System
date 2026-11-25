// 外部在线功能已禁用，保留空实现避免旧页面引用时报错。
export const getShopPluginList = async () => ({
  code: 0,
  msg: '外部在线功能已禁用',
  data: {
    list: [],
    total: 0
  }
})
