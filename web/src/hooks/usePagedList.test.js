import test from 'node:test'
import assert from 'node:assert/strict'
import { effectScope } from 'vue'
import { usePagedList } from './usePagedList.js'

const deferred = () => {
  let resolve
  const promise = new Promise((done) => { resolve = done })
  return { promise, resolve }
}

const createHarness = (options) => {
  const scope = effectScope()
  const list = scope.run(() => usePagedList(options))
  return { ...list, dispose: () => scope.stop() }
}

test('only the latest request may update collection state', async () => {
  const first = deferred()
  const second = deferred()
  let calls = 0
  const list = createHarness({
    defaults: { page: 1, pageSize: 10 },
    request: () => (++calls === 1 ? first.promise : second.promise)
  })

  const firstLoad = list.load()
  const secondLoad = list.load()
  second.resolve({ code: 0, data: { list: ['latest'], total: 1 } })
  await secondLoad
  first.resolve({ code: 0, data: { list: ['stale'], total: 99 } })
  await firstLoad

  assert.deepEqual(list.items.value, ['latest'])
  assert.equal(list.total.value, 1)
  assert.equal(list.loading.value, false)
  list.dispose()
})

test('submit, reset and page changes keep pagination rules local', async () => {
  const requests = []
  const list = createHarness({
    defaults: { page: 1, pageSize: 10, keyword: '' },
    request: async (params) => {
      requests.push(params)
      return { code: 0, data: { list: [], total: 0 } }
    }
  })

  list.search.page = 4
  list.search.keyword = 'asset'
  await list.submit()
  assert.deepEqual(requests.at(-1), { page: 1, pageSize: 10, keyword: 'asset' })

  await list.changePage(3)
  assert.equal(requests.at(-1).page, 3)

  await list.changePageSize(30)
  assert.deepEqual(requests.at(-1), { page: 1, pageSize: 30, keyword: 'asset' })

  await list.reset()
  assert.deepEqual(requests.at(-1), { page: 1, pageSize: 10, keyword: '' })
  list.dispose()
})

test('reloadAfterRemoval steps back when the current page becomes empty', async () => {
  let requestParams
  const list = createHarness({
    defaults: { page: 2, pageSize: 10 },
    request: async (params) => {
      requestParams = params
      return { code: 0, data: { list: [], total: 10 } }
    }
  })

  list.items.value = [{ ID: 1 }, { ID: 2 }]
  await list.reloadAfterRemoval(2)

  assert.equal(requestParams.page, 1)
  assert.equal(list.search.page, 1)
  list.dispose()
})

test('distinguishes request failure from a successful empty collection', async () => {
  let fail = true
  const list = createHarness({
    defaults: { page: 1, pageSize: 10 },
    request: async () => fail
      ? { code: 7, msg: '服务暂不可用' }
      : { code: 0, data: { list: [], total: 0 } }
  })

  await list.load()
  assert.equal(list.error.value, '服务暂不可用')
  assert.equal(list.loaded.value, false)

  fail = false
  await list.load()
  assert.equal(list.error.value, '')
  assert.equal(list.loaded.value, true)
  assert.deepEqual(list.items.value, [])
  list.dispose()
})

test('keeps the last successful collection when refresh throws', async () => {
  let calls = 0
  const list = createHarness({
    defaults: { page: 1, pageSize: 10 },
    request: async () => {
      calls++
      if (calls === 1) return { code: 0, data: { list: ['stable'], total: 1 } }
      throw new Error('网络中断')
    }
  })

  await list.load()
  await list.load()
  assert.deepEqual(list.items.value, ['stable'])
  assert.equal(list.total.value, 1)
  assert.equal(list.loaded.value, true)
  assert.equal(list.error.value, '网络中断')
  assert.equal(list.loading.value, false)
  list.dispose()
})
