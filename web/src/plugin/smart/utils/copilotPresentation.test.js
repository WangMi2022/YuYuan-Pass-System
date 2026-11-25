import assert from 'node:assert/strict'
import test from 'node:test'
import {
  businessColumnLabel,
  businessColumnValue,
  clarificationOptions,
  isClarificationMessage,
  messageStatus,
  shouldHideBusinessColumn
} from './copilotPresentation.js'

test('clarification responses are not presented as completed facts', () => {
  const item = { data: { needsClarification: true } }
  assert.equal(isClarificationMessage(item), true)
  assert.deepEqual(messageStatus(item), { label: '待确认', className: 'message-status--clarification' })
})

test('clarification options are rendered from the assistant response', () => {
  const item = {
    intent: 'clarification',
    content: '请明确查询口径。',
    clarificationOptions: [
      { key: 'X', label: '按合同到期', value: '合同已到期', description: '按合同到期日筛选' },
      { key: 'Y', label: '按服务到期', value: '服务已到期', description: '按服务到期日筛选' }
    ]
  }
  assert.deepEqual(clarificationOptions(item).map(({ key, label, value }) => ({ key, label, value })), [
    { key: 'A', label: '按合同到期', value: '合同已到期' },
    { key: 'B', label: '按服务到期', value: '服务已到期' }
  ])
})

test('clarification messages without response options do not invent choices', () => {
  assert.deepEqual(clarificationOptions({ intent: 'clarification', content: '请说明高价值的金额门槛。' }), [])
})

test('structured facts do not expose clarification marker or internal fields', () => {
  assert.equal(shouldHideBusinessColumn('fingerprint'), true)
  assert.equal(shouldHideBusinessColumn('assetId'), true)
  assert.equal(shouldHideBusinessColumn('needsClarification'), true)
  assert.equal(businessColumnLabel('assetName'), '资产名称')
})

test('risk enum values are translated into user-facing Chinese', () => {
  assert.equal(businessColumnValue('severity', 'high'), '高风险')
  assert.equal(businessColumnValue('status', 'acknowledged'), '已确认')
  assert.equal(businessColumnValue('warrantyEndDate', '2026-09-11'), '2026/09/11')
  assert.equal(businessColumnValue('lastDetectedAt', '2026-09-11T03:30:00Z'), '2026/9/11 11:30:00')
})
