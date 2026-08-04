import test from 'node:test'
import assert from 'node:assert/strict'
import { dateKey, recurrenceLabel, scheduleMatchesDate } from './workCalendar.js'

test('weekly recurrence matches the selected weekday after the start date', () => {
  const schedule = {
    date: '2026-08-03',
    recurrence: { enabled: true, mode: 'weekly', weekday: 1 }
  }

  assert.equal(scheduleMatchesDate(schedule, '2026-08-10'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-11'), false)
  assert.equal(recurrenceLabel(schedule), '每周星期一')
})

test('monthly recurrence matches the selected day and never renders before its start date', () => {
  const schedule = {
    date: '2026-08-05',
    recurrence: { enabled: true, mode: 'monthly', monthDay: 5 }
  }

  assert.equal(scheduleMatchesDate(schedule, '2026-08-05'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-09-05'), true)
  assert.equal(scheduleMatchesDate(schedule, '2026-08-04'), false)
  assert.equal(recurrenceLabel(schedule), '每月5日')
})

test('dateKey keeps local calendar dates stable', () => {
  assert.equal(dateKey(new Date(2026, 7, 5)), '2026-08-05')
})
