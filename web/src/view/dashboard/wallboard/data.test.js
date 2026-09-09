import test from 'node:test'
import assert from 'node:assert/strict'
import { assetSegments, attentionItems, finiteNumber, rankedLocations, share } from './data.js'

test('asset segments account for unclassified inventory and retain true proportions', () => {
  const rows = [{ key: 'in_use', quantity: 1 }, { key: 'idle', quantity: 99 }]
  const segments = assetSegments(rows, 125)
  assert.deepEqual(segments.map((item) => item.ratio), [0.8, 79.2, 20])
  assert.equal(segments[2].quantity, 25)
  assert.equal(rows.length, 2)
  assert.deepEqual(assetSegments([{ quantity: -1 }, { quantity: 'invalid' }], 0).map((item) => item.ratio), [0, 0])
})

test('invalid counts and zero totals cannot produce NaN or overflow', () => {
  assert.equal(finiteNumber(Infinity), 0)
  assert.equal(share(4, 0), 0)
  assert.equal(share(-10, 20), 0)
  assert.equal(share(50, 20), 100)
  assert.equal(finiteNumber(-2500), -2500)
})

test('location ranking uses total inventory as denominator without mutating source', () => {
  const input = [{ location: '', quantity: 5 }, { location: '仓库', quantity: 20 }]
  const result = rankedLocations(input, 100)
  assert.equal(result[0].location, '仓库')
  assert.equal(result[0].ratio, 20)
  assert.equal(result[1].location, '未标注位置')
  assert.equal(input[0].quantity, 5)
})

test('attention feed cannot disclose retained data after permission removal', () => {
  const snapshot = { access: { assets: true }, moduleLoaded: { assets: true, risk: true, invoices: true }, asset: { totalQuantity: 12 }, risk: { recentEvents: [{ title: 'restricted' }] }, invoice: { pendingCount: 99 } }
  assert.deepEqual(attentionItems(snapshot).map((item) => item.key), ['assets'])
  snapshot.moduleFailed = { assets: true }
  assert.match(attentionItems(snapshot)[0].text, /上次成功数据/)
  snapshot.moduleLoaded.assets = false
  assert.equal(attentionItems(snapshot)[0].key, 'empty')
})
