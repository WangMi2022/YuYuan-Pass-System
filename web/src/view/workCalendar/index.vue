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
            <el-popover
              v-model:visible="dateNavigatorVisible"
              placement="bottom-start"
              :width="260"
              trigger="click"
              popper-class="work-calendar-date-popper"
              @show="syncDateNavigator"
            >
              <template #reference>
                <button type="button" class="month-picker-trigger" aria-label="选择查看日期">
                  <span>浏览月份</span>
                  <strong>{{ miniMonthLabel }}</strong>
                  <el-icon><ArrowDown /></el-icon>
                </button>
              </template>
              <div class="date-navigator">
                <div class="date-navigator-year">
                  <el-button :icon="ArrowLeft" circle text aria-label="上一年" @click="changePickerYear(-1)" />
                  <el-select v-model="pickerYear" filterable aria-label="选择年份">
                    <el-option v-for="year in availableYears" :key="year" :label="`${year} 年`" :value="year" />
                  </el-select>
                  <el-button :icon="ArrowRight" circle text aria-label="下一年" @click="changePickerYear(1)" />
                </div>
                <div class="date-navigator-months" aria-label="选择月份">
                  <button
                    v-for="month in 12"
                    :key="month"
                    type="button"
                    :class="{ 'is-active': pickerMonth === month }"
                    @click="pickerMonth = month"
                  >
                    {{ month }}月
                  </button>
                </div>
                <div class="date-navigator-footer">
                  <el-select v-model="pickerDay" aria-label="选择日期">
                    <el-option v-for="day in pickerDays" :key="day" :label="`${day} 日`" :value="day" />
                  </el-select>
                  <el-button type="primary" @click="applyDateNavigator">查看</el-button>
                </div>
              </div>
            </el-popover>
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
            <div class="section-tools">
              <span>{{ activeTypes.length }}/{{ scheduleTypes.length }}</span>
              <el-popover
                v-model:visible="typeManagerVisible"
                placement="right-start"
                :width="320"
                trigger="click"
                popper-class="work-calendar-type-popper"
                @hide="cancelTypeEdit"
              >
                <template #reference>
                  <span class="type-manager-trigger">
                    <el-tooltip content="管理日程类型" placement="top">
                      <el-button :icon="Setting" circle text aria-label="管理日程类型" />
                    </el-tooltip>
                  </span>
                </template>
                <div class="type-manager">
                  <div class="type-manager-heading">
                    <strong>日程类型</strong>
                    <span>{{ scheduleTypes.length }} 项</span>
                  </div>
                  <div class="type-manager-list">
                    <div v-for="type in scheduleTypes" :key="type.value" class="type-manager-item">
                      <i :style="{ background: type.color }" />
                      <span class="type-manager-label">{{ type.label }}</span>
                      <el-tooltip content="编辑" placement="top">
                        <el-button :icon="EditPen" circle text aria-label="编辑类型" @click="startEditType(type)" />
                      </el-tooltip>
                      <el-popconfirm
                        :title="typeDeleteTitle(type.value)"
                        :disabled="!canDeleteType(type.value)"
                        @confirm="removeScheduleType(type.value)"
                      >
                        <template #reference>
                          <span class="type-delete-trigger" :title="typeDeleteTitle(type.value)">
                            <el-button
                              :icon="Delete"
                              circle
                              text
                              type="danger"
                              :disabled="!canDeleteType(type.value)"
                              aria-label="删除类型"
                            />
                          </span>
                        </template>
                      </el-popconfirm>
                    </div>
                  </div>
                  <div v-if="typeDraft" class="type-editor">
                    <el-color-picker v-model="typeDraft.color" aria-label="类型颜色" />
                    <el-input
                      v-model="typeDraft.label"
                      maxlength="12"
                      placeholder="类型名称"
                      @keyup.enter="saveScheduleType"
                    />
                    <el-button @click="cancelTypeEdit">取消</el-button>
                    <el-button type="primary" @click="saveScheduleType">保存</el-button>
                  </div>
                  <el-button
                    v-else
                    class="add-type-button"
                    :icon="Plus"
                    text
                    :disabled="scheduleTypes.length >= maxScheduleTypes"
                    @click="startAddType"
                  >
                    新增类型
                  </el-button>
                </div>
              </el-popover>
            </div>
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
          <el-form-item :label="draft.recurrence.enabled ? '开始日期' : '日期'" required>
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
        <section class="schedule-repeat-settings" aria-labelledby="schedule-repeat-title">
          <div class="schedule-repeat-header">
            <div>
              <strong id="schedule-repeat-title">重复日程</strong>
              <small>{{ repeatSummary }}</small>
            </div>
            <el-button
              class="repeat-toggle-button"
              :type="draft.recurrence.enabled ? 'primary' : 'default'"
              :plain="!draft.recurrence.enabled"
              size="small"
              :aria-pressed="draft.recurrence.enabled"
              :aria-expanded="draft.recurrence.enabled"
              aria-controls="schedule-repeat-editor"
              :aria-label="draft.recurrence.enabled ? '关闭重复日程' : '启用重复日程'"
              @click="toggleRecurrence"
            >
              {{ draft.recurrence.enabled ? '重复已启用' : '启用重复' }}
            </el-button>
          </div>
          <div
            v-if="draft.recurrence.enabled"
            id="schedule-repeat-editor"
            class="schedule-repeat-editor"
          >
            <el-radio-group
              v-model="draft.recurrence.mode"
              class="repeat-mode-toggle"
              aria-label="重复周期"
            >
              <el-radio-button label="weekly">每周</el-radio-button>
              <el-radio-button label="monthly">每月</el-radio-button>
            </el-radio-group>

            <div v-if="draft.recurrence.mode === 'weekly'" class="repeat-rule-row">
              <span>每周选择</span>
              <div class="weekday-picker" role="group" aria-label="选择每周星期几">
                <button
                  v-for="weekday in weekdayOptions"
                  :key="weekday.value"
                  type="button"
                  :class="{ 'is-active': draft.recurrence.weekday === weekday.value }"
                  :aria-pressed="draft.recurrence.weekday === weekday.value"
                  :aria-label="`每周${weekday.label}`"
                  :title="`每周${weekday.label}`"
                  @click="draft.recurrence.weekday = weekday.value"
                >
                  {{ weekday.shortLabel }}
                </button>
              </div>
            </div>

            <div v-else class="repeat-rule-row">
              <span>每月选择</span>
              <el-select v-model="draft.recurrence.monthDay" class="repeat-month-select" aria-label="选择每月几日">
                <el-option v-for="day in 31" :key="day" :label="`每月 ${day} 日`" :value="day" />
              </el-select>
            </div>
            <small class="repeat-rule-hint">每月没有该日期的月份将跳过本次日程</small>
          </div>
          <small v-else class="repeat-disabled-hint">启用后可选择每周星期或每月几日</small>
        </section>
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
import { ArrowDown, ArrowLeft, ArrowRight, Clock, Delete, EditPen, Plus, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import {
  WEEKDAY_LABELS,
  createRecurrence,
  dateKey,
  fromDateKey,
  normalizeSchedule,
  scheduleMatchesDate,
  weekdayFromKey
} from '@/utils/workCalendar'

defineOptions({ name: 'WorkCalendar' })

const eventStorageKey = 'gva-work-calendar-events'
const typeStorageKey = 'gva-work-calendar-types'
const maxScheduleTypes = 12
const weekdays = WEEKDAY_LABELS
const weekdayOptions = WEEKDAY_LABELS.map((label, index) => ({
  value: index + 1,
  label: `星期${label}`,
  shortLabel: `周${label}`
}))
const defaultScheduleTypes = [
  { value: 'task', label: '工作任务', color: '#4f7cf3' },
  { value: 'meeting', label: '会议沟通', color: '#7a61d4' },
  { value: 'asset', label: '资产盘点', color: '#18a678' },
  { value: 'reminder', label: '到期提醒', color: '#d9773c' }
]
const typePalette = ['#4f7cf3', '#7a61d4', '#18a678', '#d9773c', '#d94f70', '#168aad', '#6b8e23', '#b26bce']

const today = new Date()
const viewedMonth = ref(new Date(today.getFullYear(), today.getMonth(), 1))
const selectedDate = ref(dateKey(today))
const schedules = ref([])
const scheduleTypes = ref(defaultScheduleTypes.map((type) => ({ ...type })))
const activeTypes = ref(scheduleTypes.value.map((type) => type.value))
const dialogVisible = ref(false)
const editingId = ref('')
const draft = ref(createDraft(selectedDate.value))
const dateNavigatorVisible = ref(false)
const pickerYear = ref(today.getFullYear())
const pickerMonth = ref(today.getMonth() + 1)
const pickerDay = ref(today.getDate())
const typeManagerVisible = ref(false)
const editingTypeValue = ref('')
const typeDraft = ref(null)

const monthFormatter = new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long' })
const dateFormatter = new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' })
const lunarFormatter = new Intl.DateTimeFormat('zh-CN-u-ca-chinese', { month: 'short', day: 'numeric' })

const monthLabel = computed(() => monthFormatter.format(viewedMonth.value))
const miniMonthLabel = computed(() => `${viewedMonth.value.getFullYear()} 年 ${viewedMonth.value.getMonth() + 1} 月`)
const availableYears = computed(() => Array.from({ length: 131 }, (_, index) => 1970 + index))
const pickerDays = computed(() => Array.from(
  { length: new Date(pickerYear.value, pickerMonth.value, 0).getDate() },
  (_, index) => index + 1
))
const selectedDateLabel = computed(() => dateFormatter.format(fromDateKey(selectedDate.value)))
const repeatSummary = computed(() => {
  const recurrence = draft.value.recurrence
  if (!recurrence?.enabled) return '未启用，仅创建当天日程'
  if (recurrence.mode === 'monthly') return `每月 ${recurrence.monthDay} 日`
  return `每周${weekdayOptions.find((item) => item.value === recurrence.weekday)?.label || '星期一'}`
})
const calendarDays = computed(() => buildCalendarDays(viewedMonth.value))
const visibleSchedules = computed(() => schedules.value
  .filter((schedule) => activeTypes.value.includes(schedule.type))
  .sort((left, right) => `${left.date} ${left.time}`.localeCompare(`${right.date} ${right.time}`)))
const visibleSchedulesByDate = computed(() => {
  const grouped = new Map()
  for (const day of calendarDays.value) {
    const items = visibleSchedules.value
      .filter((schedule) => scheduleMatchesDate(schedule, day.key))
      .map((schedule) => ({ ...schedule, occurrenceDate: day.key }))
      .sort((left, right) => left.time.localeCompare(right.time))
    if (items.length) grouped.set(day.key, items)
  }
  return grouped
})
const selectedSchedules = computed(() => visibleSchedulesByDate.value.get(selectedDate.value) || [])
const monthSchedules = computed(() => calendarDays.value
  .filter((day) => day.isCurrentMonth)
  .reduce((total, day) => total + schedules.value.filter((schedule) => scheduleMatchesDate(schedule, day.key)).length, 0))

watch(schedules, persistSchedules, { deep: true })
watch(scheduleTypes, persistScheduleTypes, { deep: true })
watch([pickerYear, pickerMonth], () => {
  pickerDay.value = Math.min(pickerDay.value, pickerDays.value.length)
})
watch(() => draft.value.date, (date) => {
  if (!date || draft.value.recurrence.enabled) return
  const parsedDate = fromDateKey(date)
  draft.value.recurrence.weekday = weekdayFromKey(date)
  draft.value.recurrence.monthDay = parsedDate.getDate()
})

onMounted(() => {
  try {
    const savedTypes = JSON.parse(window.localStorage.getItem(typeStorageKey) || '[]')
    if (Array.isArray(savedTypes)) {
      const seenValues = new Set()
      const validTypes = savedTypes.filter((type) => {
        if (!isValidScheduleType(type) || seenValues.has(type.value)) return false
        seenValues.add(type.value)
        return true
      })
      if (validTypes.length) scheduleTypes.value = validTypes
    }
  } catch {
    window.localStorage.removeItem(typeStorageKey)
  }
  activeTypes.value = scheduleTypes.value.map((type) => type.value)

  try {
    const savedSchedules = JSON.parse(window.localStorage.getItem(eventStorageKey) || '[]')
    if (Array.isArray(savedSchedules)) {
      schedules.value = savedSchedules.filter(isValidSchedule).map(normalizeSchedule)
    }
  } catch {
    window.localStorage.removeItem(eventStorageKey)
  }
})

function createDraft(date) {
  return {
    title: '',
    date,
    time: '09:00',
    type: scheduleTypes.value[0]?.value || 'task',
    note: '',
    recurrence: createRecurrence(date)
  }
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
  return scheduleTypes.value.find((type) => type.value === value) || scheduleTypes.value[0] || defaultScheduleTypes[0]
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

function syncDateNavigator() {
  const selected = fromDateKey(selectedDate.value)
  pickerYear.value = viewedMonth.value.getFullYear()
  pickerMonth.value = viewedMonth.value.getMonth() + 1
  pickerDay.value = selected.getFullYear() === pickerYear.value && selected.getMonth() + 1 === pickerMonth.value
    ? selected.getDate()
    : 1
}

function changePickerYear(offset) {
  pickerYear.value = Math.min(2100, Math.max(1970, pickerYear.value + offset))
}

function applyDateNavigator() {
  const date = new Date(pickerYear.value, pickerMonth.value - 1, pickerDay.value)
  viewedMonth.value = new Date(date.getFullYear(), date.getMonth(), 1)
  selectedDate.value = dateKey(date)
  dateNavigatorVisible.value = false
}

function openCreate(date = selectedDate.value) {
  editingId.value = ''
  draft.value = createDraft(date)
  dialogVisible.value = true
}

function toggleRecurrence() {
  draft.value.recurrence.enabled = !draft.value.recurrence.enabled
}

function editSchedule(schedule) {
  editingId.value = schedule.id
  const source = { ...schedule }
  delete source.occurrenceDate
  draft.value = normalizeSchedule(source)
  dialogVisible.value = true
}

function saveSchedule() {
  const title = draft.value.title.trim()
  if (!title) {
    ElMessage.warning('请填写日程名称')
    return
  }

  const schedule = normalizeSchedule({
    ...draft.value,
    id: editingId.value || `schedule-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    title
  })
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

function startAddType() {
  if (scheduleTypes.value.length >= maxScheduleTypes) return
  editingTypeValue.value = ''
  typeDraft.value = {
    label: '',
    color: typePalette[scheduleTypes.value.length % typePalette.length]
  }
}

function startEditType(type) {
  editingTypeValue.value = type.value
  typeDraft.value = { label: type.label, color: type.color }
}

function cancelTypeEdit() {
  editingTypeValue.value = ''
  typeDraft.value = null
}

function saveScheduleType() {
  const label = typeDraft.value?.label.trim()
  if (!label) {
    ElMessage.warning('请填写类型名称')
    return
  }
  const duplicated = scheduleTypes.value.some((type) =>
    type.label.toLocaleLowerCase() === label.toLocaleLowerCase() && type.value !== editingTypeValue.value)
  if (duplicated) {
    ElMessage.warning('类型名称不能重复')
    return
  }

  if (editingTypeValue.value) {
    const index = scheduleTypes.value.findIndex((type) => type.value === editingTypeValue.value)
    if (index !== -1) {
      scheduleTypes.value.splice(index, 1, {
        ...scheduleTypes.value[index],
        label,
        color: typeDraft.value.color
      })
    }
    ElMessage.success('日程类型已更新')
  } else {
    const value = `type-${Date.now()}-${Math.random().toString(16).slice(2)}`
    scheduleTypes.value.push({ value, label, color: typeDraft.value.color })
    activeTypes.value.push(value)
    ElMessage.success('日程类型已新增')
  }
  cancelTypeEdit()
}

function scheduleTypeUsage(value) {
  return schedules.value.filter((schedule) => schedule.type === value).length
}

function canDeleteType(value) {
  return scheduleTypes.value.length > 1 && scheduleTypeUsage(value) === 0
}

function typeDeleteTitle(value) {
  const usage = scheduleTypeUsage(value)
  if (usage) return `已有 ${usage} 项日程使用该类型，不能删除`
  if (scheduleTypes.value.length === 1) return '至少保留一个日程类型'
  return '确认删除这个日程类型？'
}

function removeScheduleType(value) {
  if (!canDeleteType(value)) return
  scheduleTypes.value = scheduleTypes.value.filter((type) => type.value !== value)
  activeTypes.value = activeTypes.value.filter((typeValue) => typeValue !== value)
  if (draft.value.type === value) draft.value.type = scheduleTypes.value[0].value
  if (editingTypeValue.value === value) cancelTypeEdit()
  ElMessage.success('日程类型已删除')
}

function persistSchedules() {
  window.localStorage.setItem(eventStorageKey, JSON.stringify(schedules.value))
}

function persistScheduleTypes() {
  window.localStorage.setItem(typeStorageKey, JSON.stringify(scheduleTypes.value))
}

function isValidSchedule(schedule) {
  return schedule && typeof schedule.id === 'string' && typeof schedule.title === 'string' &&
    /^\d{4}-\d{2}-\d{2}$/.test(schedule.date) && /^\d{2}:\d{2}$/.test(schedule.time) &&
    scheduleTypes.value.some((type) => type.value === schedule.type)
}

function isValidScheduleType(type) {
  return type && typeof type.value === 'string' && type.value && typeof type.label === 'string' &&
    type.label.trim() && /^#[0-9a-f]{6}$/i.test(type.color)
}
</script>

<style scoped lang="scss">
.work-calendar { min-height: 100%; }
.calendar-workspace { display: grid; min-width: 0; grid-template-columns: 254px minmax(0, 1fr); gap: 14px; }
.calendar-sidebar { min-width: 0; align-self: start; overflow: hidden; }
.sidebar-section { padding: 15px 16px; }
.sidebar-section + .sidebar-section { border-top: 1px solid var(--na-border); }
.section-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-width: 0; margin-bottom: 12px; }
.section-heading > div:not(.section-tools) { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.section-heading span { color: var(--na-muted-foreground); font-size: .6875rem; }
.section-heading strong { overflow: hidden; color: var(--na-foreground); font-size: .8125rem; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.section-heading--plain { margin-bottom: 10px; }
.section-heading--plain > span, .section-heading > .el-icon { flex: 0 0 auto; color: var(--na-muted-foreground); font-size: .75rem; }
.month-picker-trigger { display: grid; min-width: 0; grid-template-columns: minmax(0, 1fr) 14px; gap: 2px 5px; padding: 0; border: 0; background: transparent; color: var(--na-foreground); text-align: left; }
.month-picker-trigger > span { grid-column: 1 / -1; }
.month-picker-trigger strong { min-width: 0; }
.month-picker-trigger .el-icon { align-self: center; color: var(--na-muted-foreground); font-size: .6875rem; transition: transform 160ms ease; }
.month-picker-trigger:hover strong { color: var(--na-primary); }
.month-picker-trigger:focus-visible { border-radius: 5px; outline-offset: 4px; }
.section-tools { display: flex; flex: 0 0 auto; align-items: center; gap: 3px; }
.section-tools :deep(.el-button) { width: 25px; min-width: 25px; height: 25px; padding: 0; }
.type-manager-trigger { display: inline-flex; }
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
.date-navigator { display: grid; gap: 11px; }
.date-navigator-year { display: grid; grid-template-columns: 28px minmax(0, 1fr) 28px; align-items: center; gap: 6px; }
.date-navigator-year :deep(.el-button) { width: 28px; min-width: 28px; height: 28px; padding: 0; }
.date-navigator-year :deep(.el-select__wrapper) { min-height: 30px; box-shadow: none; }
.date-navigator-months { display: grid; grid-template-columns: repeat(3, 1fr); gap: 5px; }
.date-navigator-months button { min-height: 32px; padding: 0 6px; border: 1px solid transparent; border-radius: 7px; background: var(--na-muted); color: var(--na-foreground); font-size: .75rem; font-weight: 560; }
.date-navigator-months button:hover { border-color: var(--na-border-strong); background: var(--na-card); }
.date-navigator-months button.is-active { border-color: var(--na-primary); background: var(--na-primary); color: var(--na-on-primary); }
.date-navigator-footer { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 8px; padding-top: 10px; border-top: 1px solid var(--na-border); }
.date-navigator-footer :deep(.el-select__wrapper) { min-height: 32px; }
.type-manager { display: grid; gap: 10px; }
.type-manager-heading { display: flex; align-items: center; justify-content: space-between; }
.type-manager-heading strong { color: var(--na-foreground); font-size: .8125rem; font-weight: 650; }
.type-manager-heading span { color: var(--na-muted-foreground); font-size: .6875rem; }
.type-manager-list { display: grid; max-height: 252px; overflow-y: auto; }
.type-manager-item { display: grid; min-width: 0; grid-template-columns: 9px minmax(0, 1fr) 26px 26px; align-items: center; gap: 7px; min-height: 35px; border-bottom: 1px solid var(--na-border); }
.type-manager-item > i { width: 8px; height: 8px; border-radius: 2px; }
.type-manager-label { overflow: hidden; color: var(--na-foreground); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.type-delete-trigger { display: inline-flex; }
.type-manager-item :deep(.el-button) { width: 26px; min-width: 26px; height: 26px; padding: 0; }
.type-editor { display: grid; grid-template-columns: 32px minmax(0, 1fr) auto auto; align-items: center; gap: 6px; padding-top: 2px; }
.type-editor :deep(.el-color-picker__trigger) { width: 30px; height: 30px; padding: 3px; }
.type-editor :deep(.el-input__wrapper) { min-height: 32px; }
.type-editor :deep(.el-button) { min-height: 32px; padding: 0 10px; }
.add-type-button { justify-self: start; }
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
.schedule-repeat-settings { display: grid; gap: 12px; margin: 2px 0 16px; padding: 12px; border: 1px solid var(--na-border); border-radius: 10px; background: color-mix(in srgb, var(--na-primary) 3%, var(--na-card)); }
.schedule-repeat-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.schedule-repeat-header > div { display: grid; min-width: 0; gap: 3px; }
.schedule-repeat-header strong { color: var(--na-foreground); font-size: .8125rem; font-weight: 650; }
.schedule-repeat-header small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.repeat-toggle-button { flex: 0 0 auto; min-width: 88px; }
.schedule-repeat-editor { display: grid; gap: 11px; padding-top: 11px; border-top: 1px solid var(--na-border); }
.repeat-mode-toggle { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.repeat-mode-toggle :deep(.el-radio-button__inner) { width: 100%; padding: 7px 10px; }
.repeat-rule-row { display: grid; grid-template-columns: 72px minmax(0, 1fr); align-items: center; gap: 10px; }
.repeat-rule-row > span { color: var(--na-muted-foreground); font-size: .75rem; }
.weekday-picker { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 4px; }
.weekday-picker button { min-width: 0; min-height: 30px; padding: 0 3px; border: 1px solid var(--na-border); border-radius: 7px; background: var(--na-card); color: var(--na-foreground); font-size: .6875rem; white-space: nowrap; }
.weekday-picker button:hover { border-color: var(--na-primary); color: var(--na-primary); }
.weekday-picker button:focus-visible { position: relative; z-index: 1; border-color: var(--na-primary); outline: 2px solid color-mix(in srgb, var(--na-primary) 30%, transparent); outline-offset: 1px; }
.weekday-picker button.is-active { border-color: var(--na-primary); background: var(--na-primary); color: var(--na-on-primary); font-weight: 650; }
.repeat-month-select { width: 100%; }
.repeat-rule-hint, .repeat-disabled-hint { color: var(--na-muted-foreground); font-size: .6875rem; line-height: 1.45; }
.repeat-rule-hint { margin-left: 82px; }
.repeat-disabled-hint { display: block; margin-top: -4px; }
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
  .repeat-rule-row { grid-template-columns: 1fr; gap: 6px; }
}
</style>
