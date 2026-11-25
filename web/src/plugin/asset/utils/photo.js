export const assetPhotoUrl = (photo = {}, baseURL = '', fallbackAssetId = 0) => {
  if (!photo.key) return photo.url || photo.previewUrl || ''

  const prefix = String(baseURL || '').replace(/\/+$/, '')
  const query = new URLSearchParams({ key: photo.key })
  const assetId = Number(photo.assetId || fallbackAssetId || 0)
  if (assetId > 0) query.set('assetId', String(assetId))
  if (photo.accessToken) query.set('token', photo.accessToken)
  return `${prefix}/asset/photo?${query.toString()}`
}
