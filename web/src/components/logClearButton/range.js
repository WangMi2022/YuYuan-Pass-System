const DAY_IN_MILLISECONDS = 24 * 60 * 60 * 1000

const PRESET_DAYS = {
  older7: 7,
  older30: 30,
  older90: 90
}

export const LOG_CLEAR_PRESETS = [
  { value: 'older7', label: '7 天以前' },
  { value: 'older30', label: '30 天以前' },
  { value: 'older90', label: '90 天以前' },
  { value: 'custom', label: '自定义范围' },
  { value: 'all', label: '全部日志' }
]

export const buildLogClearScope = (mode, customRange = [], now = new Date()) => {
  if (mode === 'all') return { clearAll: true }

  if (mode === 'custom') {
    if (!Array.isArray(customRange) || customRange.length !== 2) return null
    const [startTime, endTime] = customRange.map((value) => new Date(value))
    if (
      Number.isNaN(startTime.getTime()) ||
      Number.isNaN(endTime.getTime()) ||
      startTime.getTime() >= endTime.getTime()
    ) {
      return null
    }
    return {
      startTime: startTime.toISOString(),
      endTime: endTime.toISOString()
    }
  }

  const days = PRESET_DAYS[mode]
  if (!days) return null
  return {
    endTime: new Date(now.getTime() - days * DAY_IN_MILLISECONDS).toISOString()
  }
}

export const buildLogCountParams = (scope) => {
  if (!scope) return null
  const { startTime, endTime } = scope
  return {
    page: 1,
    pageSize: 1,
    ...(startTime ? { startTime } : {}),
    ...(endTime ? { endTime } : {})
  }
}
