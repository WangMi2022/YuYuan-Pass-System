import assert from 'node:assert/strict'
import test from 'node:test'

import { createRiskScanPoller } from './riskScan.js'

const deferred = () => {
  let resolve
  const promise = new Promise((done) => { resolve = done })
  return { promise, resolve }
}

const flushPromises = () => new Promise((resolve) => setImmediate(resolve))

test('risk scan polling waits for each request and refreshes once after completion', async () => {
  const scheduled = []
  const requests = []
  const progress = []
  const completed = []
  const timers = {
    setTimeout(callback) { scheduled.push(callback); return scheduled.length },
    clearTimeout() {}
  }
  const poller = createRiskScanPoller({
    timers,
    interval: 2500,
    requestStatus(runId) {
      const request = deferred()
      requests.push({ runId, request })
      return request.promise
    },
    onProgress(run) { progress.push(run.status) },
    onComplete(run) { completed.push(run.status) }
  })

  poller.start(42)
  assert.equal(scheduled.length, 1)

  const firstPoll = scheduled.shift()
  const firstResult = firstPoll()
  assert.equal(requests.length, 1)
  assert.equal(scheduled.length, 0)

  requests[0].request.resolve({ ID: 42, status: 'running' })
  await firstResult
  assert.deepEqual(progress, ['running'])
  assert.equal(scheduled.length, 1)

  const finalPoll = scheduled.shift()
  const finalResult = finalPoll()
  assert.equal(requests.length, 2)
  assert.equal(scheduled.length, 0)

  requests[1].request.resolve({ ID: 42, status: 'success' })
  await finalResult
  await flushPromises()
  assert.deepEqual(progress, ['running', 'success'])
  assert.deepEqual(completed, ['success'])
  assert.equal(scheduled.length, 0)
})

test('starting a new scan invalidates an older in-flight response', async () => {
  const scheduled = []
  const requests = []
  const progress = []
  const timers = {
    setTimeout(callback) { scheduled.push(callback); return scheduled.length },
    clearTimeout() {}
  }
  const poller = createRiskScanPoller({
    timers,
    requestStatus(runId) {
      const request = deferred()
      requests.push({ runId, request })
      return request.promise
    },
    onProgress(run) { progress.push(run.ID) }
  })

  poller.start(7)
  const oldPoll = scheduled.shift()
  const oldResult = oldPoll()
  poller.start(8)

  requests[0].request.resolve({ ID: 7, status: 'success' })
  await oldResult
  assert.deepEqual(progress, [])
  assert.equal(scheduled.length, 1)
})
