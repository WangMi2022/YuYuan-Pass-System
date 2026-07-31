import assert from 'node:assert/strict'
import test from 'node:test'
import { invoiceUploadError, MAX_INVOICE_UPLOAD_BYTES } from './upload.js'

test('accepts JPEG, PNG and PDF invoice files', () => {
  for (const type of ['image/jpeg', 'image/png', 'application/pdf']) {
    assert.equal(invoiceUploadError({ type, size: 1024 }), '')
  }
})

test('rejects unsupported, empty and oversized files', () => {
  assert.match(invoiceUploadError({ type: 'text/plain', size: 1024 }), /JPG、PNG 或 PDF/)
  assert.match(invoiceUploadError({ type: 'application/pdf', size: 0 }), /10MB/)
  assert.match(invoiceUploadError({ type: 'application/pdf', size: MAX_INVOICE_UPLOAD_BYTES + 1 }), /10MB/)
})
