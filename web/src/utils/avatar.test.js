import assert from 'node:assert/strict'
import test from 'node:test'

import { resolveAvatarUrl } from './avatar.js'

test('avatar prefers the signed preview URL returned for private media', () => {
  assert.equal(
    resolveAvatarUrl({
      headerImg: 'http://storage.local/assets/avatar.png',
      headerImgPreviewUrl: 'http://storage.local/assets/avatar.png?X-Amz-Signature=test',
      fileBaseUrl: '/files'
    }),
    'http://storage.local/assets/avatar.png?X-Amz-Signature=test'
  )
})

test('avatar resolves relative canonical paths against the file API', () => {
  assert.equal(
    resolveAvatarUrl({ headerImg: '/uploads/avatar.png', fileBaseUrl: '/files/' }),
    '/files/uploads/avatar.png'
  )
})

test('same-origin media preview paths are not prefixed with the file API twice', () => {
  assert.equal(
    resolveAvatarUrl({
      headerImgPreviewUrl: '/api/fileUploadAndDownload/preview?key=avatar.png&token=signed',
      fileBaseUrl: '/api'
    }),
    '/api/fileUploadAndDownload/preview?key=avatar.png&token=signed'
  )
})

test('explicit picture source wins and safe browser URLs stay intact', () => {
  assert.equal(
    resolveAvatarUrl({
      picSrc: ' data:image/png;base64,avatar ',
      headerImg: '/uploads/old.png',
      fileBaseUrl: '/files'
    }),
    'data:image/png;base64,avatar'
  )
})
