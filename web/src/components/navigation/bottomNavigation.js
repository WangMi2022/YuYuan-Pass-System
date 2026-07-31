export const getVisibleMenuItems = (items = []) =>
  items.filter((item) => item && !item.hidden)

export const getMenuTitle = (item) => item?.meta?.title || item?.name || ''

export const containsActiveRoute = (item, activeName) => {
  if (!item || !activeName) return false
  if (item.name === activeName) return true
  return getVisibleMenuItems(item.children).some((child) =>
    containsActiveRoute(child, activeName)
  )
}

export const flattenLeafMenuItems = (items = [], parents = []) =>
  getVisibleMenuItems(items).flatMap((item) => {
    const children = getVisibleMenuItems(item.children)
    if (children.length) {
      return flattenLeafMenuItems(children, [...parents, item])
    }
    return [
      {
        item,
        trail: parents.map(getMenuTitle)
      }
    ]
  })

export const splitNavigationItems = (items = [], maxItems = 5) => {
  const visibleItems = getVisibleMenuItems(items)
  const safeMaxItems = Math.max(2, maxItems)
  if (visibleItems.length <= safeMaxItems) {
    return { primaryItems: visibleItems, overflowItems: [] }
  }
  return {
    primaryItems: visibleItems.slice(0, safeMaxItems - 1),
    overflowItems: visibleItems.slice(safeMaxItems - 1)
  }
}
