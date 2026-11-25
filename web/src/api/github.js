// 外部源码仓库动态请求已移除，保留兼容接口供旧组件引用。
export function Commits() {
  return Promise.resolve({ data: [] })
}

export function Members() {
  return Promise.resolve({ data: [] })
}
