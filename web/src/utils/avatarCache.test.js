import assert from 'node:assert/strict'
import test from 'node:test'

import {
  getLocalCachedAvatar,
  setLocalCachedAvatar,
  fetchAndCacheAvatar
} from './avatarCache.js'

test('getLocalCachedAvatar returns null for empty or non-cached url', () => {
  assert.equal(getLocalCachedAvatar(''), null)
  assert.equal(getLocalCachedAvatar(null), null)
  assert.equal(getLocalCachedAvatar('http://example.com/not-cached.png'), null)
})

test('setLocalCachedAvatar caches data in memory and retrieves it', () => {
  const testUrl = 'http://example.com/user-avatar-1.png'
  const testData = 'data:image/png;base64,mockAvatarData'

  setLocalCachedAvatar(testUrl, testData)
  assert.equal(getLocalCachedAvatar(testUrl), testData)
})

test('fetchAndCacheAvatar caches data url directly', async () => {
  const dataUrl = 'data:image/png;base64,immediateData'
  const result = await fetchAndCacheAvatar(dataUrl)
  assert.equal(result, dataUrl)
  assert.equal(getLocalCachedAvatar(dataUrl), dataUrl)
})
