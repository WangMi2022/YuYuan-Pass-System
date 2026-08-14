export const assetPhotoUrl = (photo = {}, baseURL = '') => {
  if (!photo.key) return photo.url || ''

  const prefix = String(baseURL || '').replace(/\/+$/, '')
  return `${prefix}/asset/photo?key=${encodeURIComponent(photo.key)}`
}
