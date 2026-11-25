import test from 'node:test'
import assert from 'node:assert/strict'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)

test('cookie adapter dependency is available to the user store', () => {
  assert.doesNotThrow(() => require.resolve('universal-cookie'))
})
