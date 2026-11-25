import test from 'node:test'
import assert from 'node:assert/strict'
import { dateKey, normalizeRecurrence, recurrenceLabel, scheduleMatchesDate } from './workCalendar.js'

test('daily recurrence matches every date from its start date', () => {
  const schedule = {
    date: '2026-08-03',
    recurrence: { enabled: true, mode: 'daily' }
  }

  assert.equal(scheduleMatchesDate(schedule, '2026-08-02'), false)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-03'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-04'), true)
  assert.equal(recurrenceLabel(schedule), '每天')
})

test('weekly recurrence matches every selected weekday after the start date', () => {
  const schedule = {
    date: '2026-08-03',
    recurrence: { enabled: true, mode: 'weekly', weekdays: [1, 4] }
  }

  assert.equal(scheduleMatchesDate(schedule, '2026-08-10'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-11'), false)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-13'), true)
  assert.equal(recurrenceLabel(schedule), '每周星期一、星期四')
})

test('monthly recurrence matches every selected day and never renders before its start date', () => {
  const schedule = {
    date: '2026-08-05',
    recurrence: { enabled: true, mode: 'monthly', monthDays: [5, 20] }
  }

  assert.equal(scheduleMatchesDate(schedule, '2026-08-05'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-09-05'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-09-20'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-09-21'), false)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-04'), false)
  assert.equal(recurrenceLabel(schedule), '每月5日、20日')
})

test('recurrence normalization accepts legacy scalar values and cleans array selections', () => {
  const legacy = normalizeRecurrence({ enabled: true, mode: 'weekly', weekday: 3, monthDay: 5 }, '2026-08-05')
  assert.deepEqual(legacy.weekdays, [3])
  assert.deepEqual(legacy.monthDays, [5])

  const multiple = normalizeRecurrence({
    enabled: true,
    mode: 'monthly',
    weekdays: [7, 2, 2, 9],
    monthDays: [31, 1, 31, 0]
  }, '2026-08-05')
  assert.deepEqual(multiple.weekdays, [2, 7])
  assert.deepEqual(multiple.monthDays, [1, 31])
  assert.equal(multiple.weekday, 2)
  assert.equal(multiple.monthDay, 1)
})

test('dateKey keeps local calendar dates stable', () => {
  assert.equal(dateKey(new Date(2026, 7, 5)), '2026-08-05')
})
