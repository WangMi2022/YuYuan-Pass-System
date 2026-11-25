import { onScopeDispose, reactive, ref } from 'vue'

const defaultSelectResult = (response) => ({
  items: response.data?.list || [],
  total: Number(response.data?.total || 0)
})

/**
 * Owns the repeated state machine used by server-backed collection pages.
 * Only the latest request may update the list, so rapid filter/page changes
 * cannot be overwritten by an older response that finishes later.
 */
export const usePagedList = ({ defaults, request, selectResult = defaultSelectResult }) => {
  const initialSearch = { ...defaults }
  const search = reactive({ ...initialSearch })
  const items = ref([])
  const total = ref(0)
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref('')
  let latestRequestId = 0

  const load = async () => {
    const requestId = ++latestRequestId
    loading.value = true
    error.value = ''
    try {
      const response = await request({ ...search })
      if (requestId !== latestRequestId) return response
      if (response?.code !== 0) {
        error.value = response?.msg || '数据加载失败，请稍后重试'
        return response
      }

      const result = selectResult(response)
      items.value = result.items || []
      total.value = Number(result.total || 0)
      loaded.value = true
      return response
    } catch (requestError) {
      if (requestId === latestRequestId) {
        error.value = requestError?.message || '数据加载失败，请稍后重试'
      }
      return undefined
    } finally {
      if (requestId === latestRequestId) loading.value = false
    }
  }

  const submit = () => {
    search.page = 1
    return load()
  }

  const reset = (overrides = {}) => {
    const safeOverrides = overrides && Object.getPrototypeOf(overrides) === Object.prototype ? overrides : {}
    Object.assign(search, initialSearch, safeOverrides)
    return load()
  }

  const changePage = (value) => {
    const nextPage = Number(value)
    if (Number.isInteger(nextPage) && nextPage > 0) search.page = nextPage
    return load()
  }

  const changePageSize = (value) => {
    const nextPageSize = Number(value)
    if (Number.isInteger(nextPageSize) && nextPageSize > 0) search.pageSize = nextPageSize
    search.page = 1
    return load()
  }

  const reloadAfterRemoval = (removedCount = 1) => {
    const safeRemovedCount = Math.max(1, Number(removedCount) || 1)
    if (items.value.length <= safeRemovedCount && search.page > 1) search.page--
    return load()
  }

  onScopeDispose(() => {
    latestRequestId++
  })

  return {
    search,
    items,
    total,
    loading,
    loaded,
    error,
    load,
    submit,
    reset,
    changePage,
    changePageSize,
    reloadAfterRemoval
  }
}
