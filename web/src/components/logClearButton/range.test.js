import test from 'node:test'
import assert from 'node:assert/strict'

import { buildLogClearScope, buildLogCountParams } from './range.js'

test('retention presets clear records older than the selected number of days', () => {
  const now = new Date('2026-07-22T12:00:00.000Z')
  assert.deepEqual(buildLogClearScope('older30', [], now), {
    endTime: '2026-06-22T12:00:00.000Z'
  })
})

test('custom scope requires an ordered pair and keeps both time boundaries', () => {
  assert.equal(buildLogClearScope('custom', []), null)
  assert.equal(
    buildLogClearScope('custom', ['2026-07-22T12:00:00.000Z', '2026-07-22T11:00:00.000Z']),
    null
  )
  assert.deepEqual(
    buildLogClearScope('custom', ['2026-07-01T00:00:00.000Z', '2026-07-15T23:59:59.000Z']),
    {
      startTime: '2026-07-01T00:00:00.000Z',
      endTime: '2026-07-15T23:59:59.000Z'
    }
  )
})

test('all-time cleanup remains explicit and count requests omit destructive flags', () => {
  const scope = buildLogClearScope('all')
  assert.deepEqual(scope, { clearAll: true })
  assert.deepEqual(buildLogCountParams(scope), { page: 1, pageSize: 1 })
})
