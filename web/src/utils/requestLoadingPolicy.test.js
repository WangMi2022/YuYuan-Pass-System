import assert from 'node:assert/strict'
import test from 'node:test'

import { shouldUseRequestLoadingOverlay } from './requestLoadingPolicy.js'

test('ordinary API requests do not cover the page with a loading overlay', () => {
  assert.equal(shouldUseRequestLoadingOverlay({}), false)
})

test('blocking operations can explicitly opt in to the loading overlay', () => {
  assert.equal(
    shouldUseRequestLoadingOverlay({ loadingOption: { text: '处理中' } }),
    true
  )
})

test('legacy opt-out still wins when a loading option is supplied', () => {
  assert.equal(
    shouldUseRequestLoadingOverlay({
      donNotShowLoading: true,
      loadingOption: { text: '处理中' }
    }),
    false
  )
})
