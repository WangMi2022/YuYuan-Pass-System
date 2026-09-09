export function finiteNumber(value) {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export function share(value, total) {
  return finiteNumber(total) > 0 ? Math.min(100, Math.max(0, finiteNumber(value) / finiteNumber(total) * 100)) : 0
}

export function assetSegments(rows = [], total = 0) {
  const result = rows.map((item) => ({ ...item, quantity: Math.max(0, finiteNumber(item.quantity)) }))
  const sum = result.reduce((value, item) => value + item.quantity, 0)
  if (finiteNumber(total) > sum) result.push({ key: 'other', label: '其他状态', tone: 'muted', quantity: finiteNumber(total) - sum })
  const denominator = Math.max(sum, finiteNumber(total))
  return result.map((item) => ({ ...item, ratio: share(item.quantity, denominator) }))
}

export function rankedLocations(rows = [], total = 0) {
  return rows.map((item) => ({
    ...item,
    location: String(item.location || '未标注位置'),
    quantity: Math.max(0, finiteNumber(item.quantity)),
    ratio: share(item.quantity, total)
  })).sort((a, b) => b.quantity - a.quantity).slice(0, 5)
}

// Every message is guarded by both permission and availability. Failed refreshes
// can retain the last good snapshot, but must never look like fresh observations.
export function attentionItems(snapshot) {
  const items = []
  const allowed = (key) => snapshot.access?.[key] && snapshot.moduleLoaded?.[key]
  const stale = (key) => snapshot.moduleFailed?.[key] ? ' · 上次成功数据' : ''
  if (allowed('risk')) {
    for (const event of (snapshot.risk?.recentEvents || []).slice(0, 8)) {
      items.push({ key: `risk-${event.ID ?? event.id ?? items.length}`, label: '风险动态', text: `${event.title || '资产风险待关注'}${stale('risk')}`, tone: ['high', 'critical'].includes(event.severity) ? 'danger' : 'warning' })
    }
  }
  if (allowed('assets')) {
    items.push({ key: 'assets', label: '资产概况', text: `当前纳管 ${finiteNumber(snapshot.asset?.totalQuantity)} 件实物资产，覆盖 ${finiteNumber(snapshot.asset?.categoryCount)} 个资产分类${stale('assets')}`, tone: 'primary' })
  }
  if (allowed('invoices')) {
    items.push({ key: 'invoices', label: '票据关注', text: `待核对 ${finiteNumber(snapshot.invoice?.pendingCount)} 张 · 识别失败 ${finiteNumber(snapshot.invoice?.failedCount)} 张${stale('invoices')}`, tone: 'info' })
  }
  if (!items.length) items.push({ key: 'empty', label: '数据提示', text: '当前暂无可展示动态，可刷新获取最新数据', tone: 'muted' })
  return items
}
