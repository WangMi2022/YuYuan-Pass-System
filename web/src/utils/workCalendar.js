export const WEEKDAY_LABELS = ['一', '二', '三', '四', '五', '六', '日']

export function dateKey(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function fromDateKey(key) {
  return new Date(`${key}T00:00:00`)
}

export function weekdayFromKey(key) {
  const day = fromDateKey(key).getDay()
  return day === 0 ? 7 : day
}

export function createRecurrence(date) {
  const weekday = weekdayFromKey(date)
  const monthDay = fromDateKey(date).getDate()
  return {
    enabled: false,
    mode: 'weekly',
    weekdays: [weekday],
    monthDays: [monthDay],
    weekday,
    monthDay
  }
}

function normalizeIntegerList(values, legacyValue, fallbackValue, min, max) {
  const source = Array.isArray(values) && values.length ? values : [legacyValue]
  const normalized = source
    .map(Number)
    .filter((value) => Number.isInteger(value) && value >= min && value <= max)
  const unique = [...new Set(normalized)].sort((left, right) => left - right)
  return unique.length ? unique : [fallbackValue]
}

export function normalizeRecurrence(recurrence, date) {
  const fallback = createRecurrence(date)
  const weekdays = normalizeIntegerList(recurrence?.weekdays, recurrence?.weekday, fallback.weekday, 1, 7)
  const monthDays = normalizeIntegerList(recurrence?.monthDays, recurrence?.monthDay, fallback.monthDay, 1, 31)
  const mode = ['daily', 'weekly', 'monthly'].includes(recurrence?.mode) ? recurrence.mode : 'weekly'
  return {
    enabled: Boolean(recurrence?.enabled),
    mode,
    weekdays,
    monthDays,
    weekday: weekdays[0],
    monthDay: monthDays[0]
  }
}

export function normalizeSchedule(schedule) {
  return { ...schedule, recurrence: normalizeRecurrence(schedule.recurrence, schedule.date) }
}

export function scheduleMatchesDate(schedule, key) {
  if (!schedule?.date || key < schedule.date) return false
  const recurrence = normalizeRecurrence(schedule.recurrence, schedule.date)
  if (!recurrence.enabled) return key === schedule.date
  if (recurrence.mode === 'daily') return true
  if (recurrence.mode === 'monthly') return recurrence.monthDays.includes(fromDateKey(key).getDate())
  return recurrence.weekdays.includes(weekdayFromKey(key))
}

export function recurrenceLabel(schedule) {
  const recurrence = normalizeRecurrence(schedule.recurrence, schedule.date)
  if (!recurrence.enabled) return ''
  if (recurrence.mode === 'daily') return '每天'
  if (recurrence.mode === 'monthly') return `每月${recurrence.monthDays.map((day) => `${day}日`).join('、')}`
  return `每周${recurrence.weekdays.map((weekday) => `星期${WEEKDAY_LABELS[weekday - 1]}`).join('、')}`
}
