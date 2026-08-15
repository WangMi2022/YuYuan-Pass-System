import assert from 'node:assert/strict'
import test from 'node:test'
import { defaultInvoiceRecognition, invoiceRecognitionFormValue, invoiceRecognitionPayload } from './invoiceRecognition.js'

test('invoice recognition form fills nested provider defaults', () => {
  const form = invoiceRecognitionFormValue({
    'fallback-threshold': 0.9,
    baidu: { enabled: true, 'api-key-configured': true }
  })

  assert.equal(form['fallback-threshold'], 0.9)
  assert.equal(form.baidu.enabled, true)
  assert.equal(form.baidu['api-key-configured'], true)
  assert.equal(form.baidu['timeout-seconds'], 30)
  assert.equal(form.multimodal['timeout-seconds'], 45)
})

test('invoice recognition payload is detached from form state', () => {
  const form = defaultInvoiceRecognition()
  form.multimodal.model = 'vision-model'
  const payload = invoiceRecognitionPayload(form)

  payload.multimodal.model = 'changed'
  assert.equal(form.multimodal.model, 'vision-model')
})
