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
  return {
    enabled: false,
    mode: 'weekly',
    weekday: weekdayFromKey(date),
    monthDay: fromDateKey(date).getDate()
  }
}

export function normalizeRecurrence(recurrence, date) {
  const fallback = createRecurrence(date)
  const weekday = Number(recurrence?.weekday)
  const monthDay = Number(recurrence?.monthDay)
  return {
    enabled: Boolean(recurrence?.enabled),
    mode: recurrence?.mode === 'monthly' ? 'monthly' : 'weekly',
    weekday: Number.isInteger(weekday) && weekday >= 1 && weekday <= 7 ? weekday : fallback.weekday,
    monthDay: Number.isInteger(monthDay) && monthDay >= 1 && monthDay <= 31 ? monthDay : fallback.monthDay
  }
}

export function normalizeSchedule(schedule) {
  return { ...schedule, recurrence: normalizeRecurrence(schedule.recurrence, schedule.date) }
}

export function scheduleMatchesDate(schedule, key) {
  if (!schedule?.date || key < schedule.date) return false
  const recurrence = normalizeRecurrence(schedule.recurrence, schedule.date)
  if (!recurrence.enabled) return key === schedule.date
  if (recurrence.mode === 'monthly') return fromDateKey(key).getDate() === recurrence.monthDay
  return weekdayFromKey(key) === recurrence.weekday
}

export function recurrenceLabel(schedule) {
  const recurrence = normalizeRecurrence(schedule.recurrence, schedule.date)
  if (!recurrence.enabled) return ''
  if (recurrence.mode === 'monthly') return `每月${recurrence.monthDay}日`
  return `每周星期${WEEKDAY_LABELS[recurrence.weekday - 1]}`
}
