import assert from 'node:assert/strict'
import test from 'node:test'

import { assetPhotoUrl } from './photo.js'

test('asset photo URL uses the configured API prefix and encoded object key', () => {
  assert.equal(
    assetPhotoUrl({ key: 'assets/2026-08-14/设备 01.png' }, '/gateway/'),
    '/gateway/asset/photo?key=assets%2F2026-08-14%2F%E8%AE%BE%E5%A4%87%2001.png'
  )
})

test('asset photo URL falls back to the stored URL when no key exists', () => {
  assert.equal(assetPhotoUrl({ url: 'https://cdn.example.com/photo.png' }, '/api'), 'https://cdn.example.com/photo.png')
})
