import assert from 'node:assert/strict'
import test from 'node:test'

import {
  providerFormValue,
  providerPayloadValue
} from './provider.js'

test('provider cost fields round-trip between yuan and integer micros', () => {
  const form = providerFormValue({
    model: 'MiniMax-M3',
    'api-key-configured': true,
    'input-cost-micros-per-million': 125000,
    'output-cost-micros-per-million': 4567890
  })

  assert.equal(form['input-cost-per-million'], 0.125)
  assert.equal(form['output-cost-per-million'], 4.56789)
  assert.equal(form['api-key-configured'], true)

  const payload = providerPayloadValue(form)
  assert.equal(payload['input-cost-micros-per-million'], 125000)
  assert.equal(payload['output-cost-micros-per-million'], 4567890)
  assert.equal('input-cost-per-million' in payload, false)
  assert.equal('output-cost-per-million' in payload, false)
  assert.equal('api-key-configured' in payload, false)
})

test('provider decimal costs are rounded to the nearest micro unit', () => {
  const payload = providerPayloadValue({
    'input-cost-per-million': 0.0000014,
    'output-cost-per-million': 0.0000016
  })

  assert.equal(payload['input-cost-micros-per-million'], 1)
  assert.equal(payload['output-cost-micros-per-million'], 2)
})
