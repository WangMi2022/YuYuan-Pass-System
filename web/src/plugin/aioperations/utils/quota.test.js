import assert from 'node:assert/strict'
import test from 'node:test'

import { defaultQuota, quotaFormValue, quotaPayloadValue } from './quota.js'

test('quota monthly budget round-trips between yuan and micros', () => {
  const form = quotaFormValue({
    ID: 7,
    scopeType: 'module',
    scopeId: 'asset',
    monthlyCostMicros: 12345678
  })

  assert.equal(form.monthlyBudgetYuan, 12.345678)
  assert.equal(form.monthlyCostMicros, 12345678)

  const payload = quotaPayloadValue(form)
  assert.equal(payload.monthlyCostMicros, 12345678)
  assert.equal('monthlyBudgetYuan' in payload, false)
})

test('quota defaults accept decimal yuan values', () => {
  const form = defaultQuota()
  form.monthlyBudgetYuan = 0.01

  assert.equal(quotaPayloadValue(form).monthlyCostMicros, 10000)
})
