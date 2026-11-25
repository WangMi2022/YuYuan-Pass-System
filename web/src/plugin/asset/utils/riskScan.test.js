import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import { createRiskScanPoller } from './riskScan.js'

const deferred = () => {
  let resolve
  const promise = new Promise((done) => { resolve = done })
  return { promise, resolve }
}

const flushPromises = () => new Promise((resolve) => setImmediate(resolve))

test('scan history keeps multiple flexible columns so the table fills its workspace', async () => {
  const filename = fileURLToPath(new URL('../view/risk.vue', import.meta.url))
  const source = await readFile(filename, 'utf8')
  const scanPane = source.match(/<el-tab-pane name="scans">([\s\S]*?)<\/el-tab-pane>/)?.[1] || ''
  const columns = [...scanPane.matchAll(/<el-table-column\b([^>]*)>/g)].map((match) => match[1])
  const flexibleColumns = columns.filter((attributes) => /\bmin-width=/.test(attributes))

  assert.ok(columns.length > 0, 'scan history columns must be present')
  assert.ok(
    flexibleColumns.length >= 4,
    `expected at least 4 flexible scan columns, received ${flexibleColumns.length}`
  )
})

test('scan history exposes guarded selected and finished-record cleanup actions', async () => {
  const viewFilename = fileURLToPath(new URL('../view/risk.vue', import.meta.url))
  const apiFilename = fileURLToPath(new URL('../api/risk.js', import.meta.url))
  const [source, apiSource] = await Promise.all([
    readFile(viewFilename, 'utf8'),
    readFile(apiFilename, 'utf8')
  ])
  const scanPane = source.match(/<el-tab-pane name="scans">([\s\S]*?)<\/el-tab-pane>/)?.[1] || ''

  assert.match(scanPane, /type="selection"[^>]+:selectable="isDeletableScan"/)
  assert.match(scanPane, /@click="deleteSelectedScans"/)
  assert.match(scanPane, /@click="clearFinishedScans"/)
  assert.match(scanPane, /@click="deleteScan\(row\)"/)
  assert.match(source, /scan\.status !== 'running'/)
  assert.match(apiSource, /url: '\/assetRisk\/scans', method: 'delete', data/)
})

test('risk events expose terminal-only selected and history cleanup actions', async () => {
  const viewFilename = fileURLToPath(new URL('../view/risk.vue', import.meta.url))
  const apiFilename = fileURLToPath(new URL('../api/risk.js', import.meta.url))
  const [source, apiSource] = await Promise.all([
    readFile(viewFilename, 'utf8'),
    readFile(apiFilename, 'utf8')
  ])
  const eventPane = source.match(/<el-tab-pane name="events">([\s\S]*?)<el-tab-pane name="rules">/)?.[1] || ''

  assert.match(eventPane, /@click="clearRiskHistory"/)
  assert.match(eventPane, /@click="deleteSelectedEvents"/)
  assert.match(eventPane, /@click\.stop="deleteEvent\(row\)"/)
  assert.match(source, /\['resolved', 'ignored'\]\.includes\(risk\?\.status\)/)
  assert.match(apiSource, /url: '\/assetRisk\/events', method: 'delete', data/)
})

test('risk tables keep horizontal overflow local and fixed columns non-overlapping', async () => {
  const filename = fileURLToPath(new URL('../view/risk.vue', import.meta.url))
  const source = await readFile(filename, 'utf8')
  const eventPane = source.match(/<el-tab-pane name="events">([\s\S]*?)<el-tab-pane name="rules">/)?.[1] || ''
  const scanPane = source.match(/<el-tab-pane name="scans">([\s\S]*?)<\/el-tab-pane>/)?.[1] || ''

  assert.equal((source.match(/class="risk-table-shell"/g) || []).length, 3)
  assert.match(eventPane, /type="selection"[^>]+fixed="left"/)
  assert.match(scanPane, /label="结果" min-width="220"(?![^>]*fixed="right")/)
  assert.match(scanPane, /label="操作" width="176" fixed="right"/)
})

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
