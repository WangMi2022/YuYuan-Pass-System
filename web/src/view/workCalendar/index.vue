<template>
  <main class="na-page na-page--list work-calendar">
    <AppPageHeader
      title-id="work-calendar-title"
      title="工作日历"
      description="集中查看和安排个人工作日程"
    >
      <template #actions>
        <el-button type="primary" :icon="Plus" @click="openCreate()">新建日程</el-button>
      </template>
    </AppPageHeader>

    <section class="calendar-workspace" aria-label="工作日历">
      <aside class="na-panel calendar-sidebar">
        <div class="sidebar-section mini-calendar-section">
          <div class="section-heading">
            <div>
              <span>浏览月份</span>
              <strong>{{ miniMonthLabel }}</strong>
            </div>
            <div class="month-nav">
              <el-tooltip content="上个月" placement="top">
                <el-button :icon="ArrowLeft" circle text aria-label="上个月" @click="changeMonth(-1)" />
              </el-tooltip>
              <el-tooltip content="下个月" placement="top">
                <el-button :icon="ArrowRight" circle text aria-label="下个月" @click="changeMonth(1)" />
              </el-tooltip>
            </div>
          </div>
          <div class="mini-weekdays" aria-hidden="true">
            <span v-for="weekday in weekdays" :key="weekday">{{ weekday }}</span>
          </div>
          <div class="mini-calendar-grid">
            <button
              v-for="day in calendarDays"
              :key="day.key"
              type="button"
              :class="{ 'is-other': !day.isCurrentMonth, 'is-selected': isSelected(day), 'is-today': day.isToday }"
              :aria-label="formatDate(day.key)"
              @click="selectDay(day)"
            >
              {{ day.date.getDate() }}
            </button>
          </div>
        </div>

        <div class="sidebar-section schedule-filter-section">
          <div class="section-heading section-heading--plain">
            <div><strong>日程类型</strong></div>
            <span>{{ activeTypes.length }}/{{ scheduleTypes.length }}</span>
          </div>
          <el-checkbox-group v-model="activeTypes" class="schedule-filters">
            <el-checkbox v-for="type in scheduleTypes" :key="type.value" :label="type.value">
              <i :style="{ background: type.color }" />
              <span>{{ type.label }}</span>
            </el-checkbox>
          </el-checkbox-group>
        </div>

        <div class="sidebar-section selected-day-section">
          <div class="section-heading section-heading--plain">
            <div>
              <span>选中日期</span>
              <strong>{{ selectedDateLabel }}</strong>
            </div>
            <el-icon><Clock /></el-icon>
          </div>
          <div v-if="selectedSchedules.length" class="selected-schedule-list">
            <button
              v-for="schedule in selectedSchedules"
              :key="schedule.id"
              type="button"
              @click="editSchedule(schedule)"
            >
              <i :style="{ background: typeInfo(schedule.type).color }" />
              <span>{{ schedule.time }}</span>
              <strong>{{ schedule.title }}</strong>
            </button>
          </div>
          <button v-else type="button" class="empty-day" @click="openCreate(selectedDate)">
            <el-icon><Plus /></el-icon>
            <span>添加当天日程</span>
          </button>
        </div>
      </aside>

      <section class="na-panel calendar-main">
        <header class="calendar-toolbar">
          <div class="calendar-title-group">
            <span>月视图</span>
            <h2>{{ monthLabel }}</h2>
            <small>农历 {{ lunarText(viewedMonth) }}</small>
          </div>
          <div class="calendar-actions">
            <span class="month-count">本月 {{ monthSchedules.length }} 项</span>
            <el-button @click="goToToday">今天</el-button>
            <div class="month-nav month-nav--main">
              <el-tooltip content="上个月" placement="top">
                <el-button :icon="ArrowLeft" circle aria-label="上个月" @click="changeMonth(-1)" />
              </el-tooltip>
              <el-tooltip content="下个月" placement="top">
                <el-button :icon="ArrowRight" circle aria-label="下个月" @click="changeMonth(1)" />
              </el-tooltip>
            </div>
          </div>
        </header>

        <div class="calendar-board-scroll">
          <div class="calendar-board">
            <div class="calendar-weekdays" aria-hidden="true">
              <span v-for="weekday in weekdays" :key="weekday">周{{ weekday }}</span>
            </div>
            <div class="calendar-days">
              <div
                v-for="day in calendarDays"
                :key="day.key"
                class="calendar-day"
                :class="{
                  'is-other': !day.isCurrentMonth,
                  'is-selected': isSelected(day),
                  'is-today': day.isToday
                }"
                role="button"
                tabindex="0"
                :aria-label="`${formatDate(day.key)}，${visibleEventsFor(day.key).length} 项日程`"
                @click="selectDay(day)"
                @keydown.enter.prevent="selectDay(day)"
                @keydown.space.prevent="selectDay(day)"
              >
                <div class="day-heading">
                  <span>{{ day.date.getDate() }}</span>
                  <small>{{ lunarText(day.date) }}</small>
                </div>
                <div class="day-events">
                  <button
                    v-for="schedule in visibleEventsFor(day.key).slice(0, 3)"
                    :key="schedule.id"
                    type="button"
                    :style="{ '--schedule-color': typeInfo(schedule.type).color }"
                    @click.stop="editSchedule(schedule)"
                  >
                    <i />
                    <span>{{ schedule.time }}</span>
                    <strong>{{ schedule.title }}</strong>
                  </button>
                  <button
                    v-if="visibleEventsFor(day.key).length > 3"
                    type="button"
                    class="more-schedules"
                    @click.stop="selectDay(day)"
                  >
                    还有 {{ visibleEventsFor(day.key).length - 3 }} 项
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </section>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑日程' : '新建日程'" width="480px" destroy-on-close>
      <el-form label-position="top" @submit.prevent="saveSchedule">
        <el-form-item label="日程名称" required>
          <el-input v-model="draft.title" maxlength="60" show-word-limit placeholder="输入日程名称" />
        </el-form-item>
        <div class="schedule-form-grid">
          <el-form-item label="日期" required>
            <el-date-picker v-model="draft.date" type="date" value-format="YYYY-MM-DD" :clearable="false" />
          </el-form-item>
          <el-form-item label="时间" required>
            <el-time-select
              v-model="draft.time"
              start="00:00"
              step="00:15"
              end="23:45"
              :clearable="false"
            />
          </el-form-item>
        </div>
        <el-form-item label="日程类型" required>
          <el-select v-model="draft.type">
            <el-option v-for="type in scheduleTypes" :key="type.value" :label="type.label" :value="type.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="draft.note" type="textarea" :rows="3" maxlength="160" show-word-limit placeholder="补充说明（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-actions">
          <el-popconfirm v-if="editingId" title="确认删除这条日程？" @confirm="removeSchedule(editingId)">
            <template #reference>
              <el-tooltip content="删除日程" placement="top">
                <el-button :icon="Delete" circle type="danger" plain aria-label="删除日程" />
              </el-tooltip>
            </template>
          </el-popconfirm>
          <span />
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveSchedule">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight, Clock, Delete, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'

defineOptions({ name: 'WorkCalendar' })

const storageKey = 'gva-work-calendar-events'
const weekdays = ['一', '二', '三', '四', '五', '六', '日']
const scheduleTypes = [
  { value: 'task', label: '工作任务', color: '#4f7cf3' },
  { value: 'meeting', label: '会议沟通', color: '#7a61d4' },
  { value: 'asset', label: '资产盘点', color: '#18a678' },
  { value: 'reminder', label: '到期提醒', color: '#d9773c' }
]

const today = new Date()
const viewedMonth = ref(new Date(today.getFullYear(), today.getMonth(), 1))
const selectedDate = ref(dateKey(today))
const schedules = ref([])
const activeTypes = ref(scheduleTypes.map((type) => type.value))
const dialogVisible = ref(false)
const editingId = ref('')
const draft = ref(createDraft(selectedDate.value))

const monthFormatter = new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long' })
const dateFormatter = new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' })
const lunarFormatter = new Intl.DateTimeFormat('zh-CN-u-ca-chinese', { month: 'short', day: 'numeric' })

const monthLabel = computed(() => monthFormatter.format(viewedMonth.value))
const miniMonthLabel = computed(() => `${viewedMonth.value.getFullYear()} 年 ${viewedMonth.value.getMonth() + 1} 月`)
const selectedDateLabel = computed(() => dateFormatter.format(fromDateKey(selectedDate.value)))
const calendarDays = computed(() => buildCalendarDays(viewedMonth.value))
const visibleSchedules = computed(() => schedules.value
  .filter((schedule) => activeTypes.value.includes(schedule.type))
  .sort((left, right) => `${left.date} ${left.time}`.localeCompare(`${right.date} ${right.time}`)))
const visibleSchedulesByDate = computed(() => {
  const grouped = new Map()
  for (const schedule of visibleSchedules.value) {
    const items = grouped.get(schedule.date) || []
    items.push(schedule)
    grouped.set(schedule.date, items)
  }
  return grouped
})
const selectedSchedules = computed(() => visibleSchedulesByDate.value.get(selectedDate.value) || [])
const monthSchedules = computed(() => schedules.value.filter((schedule) => {
  const date = fromDateKey(schedule.date)
  return date.getFullYear() === viewedMonth.value.getFullYear() && date.getMonth() === viewedMonth.value.getMonth()
}))

watch(schedules, persistSchedules, { deep: true })

onMounted(() => {
  try {
    const savedSchedules = JSON.parse(window.localStorage.getItem(storageKey) || '[]')
    if (Array.isArray(savedSchedules)) {
      schedules.value = savedSchedules.filter(isValidSchedule)
    }
  } catch {
    window.localStorage.removeItem(storageKey)
  }
})

function createDraft(date) {
  return { title: '', date, time: '09:00', type: 'task', note: '' }
}

function dateKey(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function fromDateKey(key) {
  return new Date(`${key}T00:00:00`)
}

function buildCalendarDays(month) {
  const firstDay = new Date(month.getFullYear(), month.getMonth(), 1)
  const mondayOffset = (firstDay.getDay() + 6) % 7
  const startDate = new Date(firstDay)
  startDate.setDate(startDate.getDate() - mondayOffset)
  const todayKey = dateKey(new Date())

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(startDate)
    date.setDate(startDate.getDate() + index)
    return {
      date,
      key: dateKey(date),
      isCurrentMonth: date.getMonth() === month.getMonth(),
      isToday: dateKey(date) === todayKey
    }
  })
}

function formatDate(key) {
  return dateFormatter.format(fromDateKey(key))
}

function lunarText(date) {
  try {
    return lunarFormatter.format(date).replace(/\s/g, '')
  } catch {
    return ''
  }
}

function isSelected(day) {
  return day.key === selectedDate.value
}

function typeInfo(value) {
  return scheduleTypes.find((type) => type.value === value) || scheduleTypes[0]
}

function visibleEventsFor(key) {
  return visibleSchedulesByDate.value.get(key) || []
}

function selectDay(day) {
  selectedDate.value = day.key
  if (!day.isCurrentMonth) {
    viewedMonth.value = new Date(day.date.getFullYear(), day.date.getMonth(), 1)
  }
}

function changeMonth(offset) {
  viewedMonth.value = new Date(viewedMonth.value.getFullYear(), viewedMonth.value.getMonth() + offset, 1)
}

function goToToday() {
  const now = new Date()
  viewedMonth.value = new Date(now.getFullYear(), now.getMonth(), 1)
  selectedDate.value = dateKey(now)
}

function openCreate(date = selectedDate.value) {
  editingId.value = ''
  draft.value = createDraft(date)
  dialogVisible.value = true
}

function editSchedule(schedule) {
  editingId.value = schedule.id
  draft.value = { ...schedule }
  dialogVisible.value = true
}

function saveSchedule() {
  const title = draft.value.title.trim()
  if (!title) {
    ElMessage.warning('请填写日程名称')
    return
  }

  const schedule = {
    ...draft.value,
    id: editingId.value || `schedule-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    title
  }
  const index = schedules.value.findIndex((item) => item.id === schedule.id)
  if (index === -1) {
    schedules.value = [...schedules.value, schedule]
  } else {
    schedules.value.splice(index, 1, schedule)
  }

  selectedDate.value = schedule.date
  viewedMonth.value = new Date(fromDateKey(schedule.date).getFullYear(), fromDateKey(schedule.date).getMonth(), 1)
  dialogVisible.value = false
  ElMessage.success(editingId.value ? '日程已更新' : '日程已创建')
}

function removeSchedule(id) {
  schedules.value = schedules.value.filter((schedule) => schedule.id !== id)
  dialogVisible.value = false
  ElMessage.success('日程已删除')
}

function persistSchedules() {
  window.localStorage.setItem(storageKey, JSON.stringify(schedules.value))
}

function isValidSchedule(schedule) {
  return schedule && typeof schedule.id === 'string' && typeof schedule.title === 'string' &&
    /^\d{4}-\d{2}-\d{2}$/.test(schedule.date) && /^\d{2}:\d{2}$/.test(schedule.time) &&
    scheduleTypes.some((type) => type.value === schedule.type)
}
</script>

<style scoped lang="scss">
.work-calendar { min-height: 100%; }
.calendar-workspace { display: grid; min-width: 0; grid-template-columns: 254px minmax(0, 1fr); gap: 14px; }
.calendar-sidebar { min-width: 0; align-self: start; overflow: hidden; }
.sidebar-section { padding: 15px 16px; }
.sidebar-section + .sidebar-section { border-top: 1px solid var(--na-border); }
.section-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-width: 0; margin-bottom: 12px; }
.section-heading > div { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.section-heading span { color: var(--na-muted-foreground); font-size: .6875rem; }
.section-heading strong { overflow: hidden; color: var(--na-foreground); font-size: .8125rem; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.section-heading--plain { margin-bottom: 10px; }
.section-heading--plain > span, .section-heading > .el-icon { flex: 0 0 auto; color: var(--na-muted-foreground); font-size: .75rem; }
.month-nav { display: flex; flex: 0 0 auto; align-items: center; gap: 1px; }
.month-nav :deep(.el-button) { width: 26px; min-width: 26px; height: 26px; padding: 0; }
.mini-weekdays, .mini-calendar-grid { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); text-align: center; }
.mini-weekdays { margin-bottom: 4px; color: var(--na-muted-foreground); font-size: .625rem; font-weight: 600; line-height: 22px; }
.mini-calendar-grid { gap: 2px; }
.mini-calendar-grid button { display: inline-grid; width: 100%; min-width: 0; height: 26px; place-items: center; padding: 0; border: 1px solid transparent; border-radius: 7px; background: transparent; color: var(--na-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.mini-calendar-grid button:hover { background: var(--na-muted); }
.mini-calendar-grid button.is-other { color: color-mix(in srgb, var(--na-muted-foreground) 52%, transparent); }
.mini-calendar-grid button.is-today { color: var(--na-primary); font-weight: 700; }
.mini-calendar-grid button.is-selected { border-color: var(--na-primary); background: var(--na-primary); color: var(--na-on-primary); font-weight: 700; }
.schedule-filters { display: grid; gap: 4px; }
.schedule-filters :deep(.el-checkbox) { width: 100%; height: 26px; margin-right: 0; }
.schedule-filters :deep(.el-checkbox__label) { display: inline-flex; min-width: 0; align-items: center; gap: 7px; padding-left: 7px; color: var(--na-foreground); font-size: .75rem; }
.schedule-filters i { width: 7px; height: 7px; border-radius: 2px; }
.selected-day-section { min-height: 144px; }
.selected-schedule-list { display: grid; gap: 4px; }
.selected-schedule-list button { display: grid; min-width: 0; grid-template-columns: 7px 38px minmax(0, 1fr); align-items: center; gap: 7px; min-height: 28px; padding: 0 5px; border: 0; border-radius: 6px; background: transparent; color: var(--na-foreground); text-align: left; }
.selected-schedule-list button:hover { background: var(--na-muted); }
.selected-schedule-list i { width: 6px; height: 6px; border-radius: 50%; }
.selected-schedule-list span { color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.selected-schedule-list strong { overflow: hidden; font-size: .75rem; font-weight: 550; text-overflow: ellipsis; white-space: nowrap; }
.empty-day { display: inline-flex; align-items: center; gap: 6px; min-height: 28px; padding: 0 4px; border: 0; background: transparent; color: var(--na-muted-foreground); font-size: .75rem; }
.empty-day:hover { color: var(--na-primary); }
.empty-day .el-icon { font-size: .875rem; }
.calendar-main { min-width: 0; overflow: hidden; }
.calendar-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 14px; min-height: 73px; padding: 13px 18px; border-bottom: 1px solid var(--na-border); }
.calendar-title-group { display: grid; min-width: 0; grid-template-columns: auto auto; align-items: baseline; gap: 0 9px; }
.calendar-title-group > span { color: var(--na-primary); font-size: .6875rem; font-weight: 650; }
.calendar-title-group h2 { margin: 0; color: var(--na-foreground); font-size: 1.125rem; font-weight: 680; letter-spacing: 0; }
.calendar-title-group small { grid-column: 1 / -1; margin-top: 3px; color: var(--na-muted-foreground); font-size: .6875rem; }
.calendar-actions { display: flex; align-items: center; gap: 8px; }
.month-count { color: var(--na-muted-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; white-space: nowrap; }
.month-nav--main { margin-left: 2px; padding-left: 8px; border-left: 1px solid var(--na-border); }
.month-nav--main :deep(.el-button) { width: 30px; min-width: 30px; height: 30px; }
.calendar-board-scroll { overflow: auto; }
.calendar-board { min-width: 760px; }
.calendar-weekdays, .calendar-days { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); }
.calendar-weekdays { border-bottom: 1px solid var(--na-border); background: color-mix(in srgb, var(--na-muted) 68%, var(--na-card)); }
.calendar-weekdays span { padding: 9px 10px; color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 650; text-align: right; }
.calendar-days { grid-auto-rows: minmax(104px, 1fr); }
.calendar-day { min-width: 0; min-height: 112px; padding: 8px 8px 6px; border-right: 1px solid var(--na-border); border-bottom: 1px solid var(--na-border); background: var(--na-card); outline: none; transition: background-color 140ms ease; }
.calendar-day:nth-child(7n) { border-right: 0; }
.calendar-day:hover { background: color-mix(in srgb, var(--na-primary) 3%, var(--na-card)); }
.calendar-day:focus-visible { position: relative; z-index: 1; box-shadow: inset 0 0 0 2px var(--na-primary); }
.calendar-day.is-other { background: color-mix(in srgb, var(--na-muted) 58%, var(--na-card)); }
.calendar-day.is-other .day-heading { opacity: .5; }
.calendar-day.is-selected { background: color-mix(in srgb, var(--na-primary) 6%, var(--na-card)); }
.calendar-day.is-today .day-heading > span { display: grid; width: 23px; height: 23px; place-items: center; border-radius: 50%; background: var(--na-primary); color: var(--na-on-primary); }
.day-heading { display: flex; align-items: center; justify-content: space-between; gap: 6px; min-height: 24px; }
.day-heading > span { color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; font-weight: 640; }
.day-heading small { overflow: hidden; color: var(--na-muted-foreground); font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }
.day-events { display: grid; gap: 3px; margin-top: 5px; }
.day-events > button { display: grid; min-width: 0; grid-template-columns: 5px 34px minmax(0, 1fr); align-items: center; gap: 5px; min-height: 22px; padding: 0 4px; border: 0; border-radius: 4px; background: color-mix(in srgb, var(--schedule-color) 11%, var(--na-card)); color: var(--na-foreground); text-align: left; }
.day-events > button:hover { background: color-mix(in srgb, var(--schedule-color) 18%, var(--na-card)); }
.day-events i { width: 5px; height: 5px; border-radius: 50%; background: var(--schedule-color); }
.day-events span { color: var(--na-muted-foreground); font-size: .625rem; font-variant-numeric: tabular-nums; }
.day-events strong { overflow: hidden; font-size: .6875rem; font-weight: 560; text-overflow: ellipsis; white-space: nowrap; }
.day-events .more-schedules { display: block; padding-left: 4px; background: transparent; color: var(--na-primary); font-size: .6875rem; font-weight: 600; }
.dialog-actions { display: grid; grid-template-columns: auto 1fr auto auto; align-items: center; gap: 8px; }
.schedule-form-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 12px; }
.schedule-form-grid :deep(.el-date-editor), .schedule-form-grid :deep(.el-select) { width: 100%; }

@media (max-width: 1080px) {
  .calendar-workspace { grid-template-columns: 224px minmax(0, 1fr); }
  .calendar-sidebar .sidebar-section { padding-right: 13px; padding-left: 13px; }
}

@media (max-width: 820px) {
  .calendar-workspace { grid-template-columns: 1fr; }
  .calendar-sidebar { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .calendar-sidebar .sidebar-section + .sidebar-section { border-top: 0; border-left: 1px solid var(--na-border); }
  .calendar-sidebar .selected-day-section { grid-column: 1 / -1; border-top: 1px solid var(--na-border); border-left: 0; }
}

@media (max-width: 620px) {
  .calendar-toolbar { align-items: flex-start; flex-direction: column; }
  .calendar-actions { width: 100%; justify-content: space-between; }
  .month-count { margin-right: auto; }
  .calendar-sidebar { grid-template-columns: 1fr; }
  .calendar-sidebar .sidebar-section + .sidebar-section { border-top: 1px solid var(--na-border); border-left: 0; }
  .calendar-sidebar .selected-day-section { grid-column: auto; }
  .schedule-form-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
