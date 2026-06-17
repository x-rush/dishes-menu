<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps<{ modelValue: string | null }>()
const emit = defineEmits<{ 'update:modelValue': [string | null] }>()

const rootRef = ref<HTMLElement | null>(null)
const popoverOpen = ref(false)

const today = new Date()
const todayStr = (() => {
  const y = today.getFullYear()
  const m = String(today.getMonth() + 1).padStart(2, '0')
  const d = String(today.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
})()

const viewYear = ref(today.getFullYear())
const viewMonth = ref(today.getMonth()) // 0-indexed

function fmt(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const displayLabel = computed(() => {
  if (!props.modelValue) return '选个日期'
  if (props.modelValue === todayStr) return '今天'
  const t = new Date(props.modelValue + 'T00:00:00')
  const weekday = ['日', '一', '二', '三', '四', '五', '六'][t.getDay()]
  return `${props.modelValue.slice(5)} · 周${weekday}`
})

const monthLabel = computed(() => `${viewYear.value} 年 ${viewMonth.value + 1} 月`)
const weekdayLabels = ['日', '一', '二', '三', '四', '五', '六']

const calendarCells = computed(() => {
  const firstDay = new Date(viewYear.value, viewMonth.value, 1)
  const startDow = firstDay.getDay() // 0=Sun
  const daysInMonth = new Date(viewYear.value, viewMonth.value + 1, 0).getDate()
  const cells: { date: Date; inMonth: boolean }[] = []
  // leading days from previous month
  for (let i = startDow - 1; i >= 0; i--) {
    const d = new Date(viewYear.value, viewMonth.value, -i)
    cells.push({ date: d, inMonth: false })
  }
  // in-month days
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ date: new Date(viewYear.value, viewMonth.value, d), inMonth: true })
  }
  // trailing days to fill 42 cells (6 weeks)
  while (cells.length < 42) {
    const last = cells[cells.length - 1].date
    const next = new Date(last)
    next.setDate(next.getDate() + 1)
    cells.push({ date: next, inMonth: false })
  }
  return cells.slice(0, 42)
})

function isPast(d: Date) {
  return d < new Date(today.getFullYear(), today.getMonth(), today.getDate())
}
function isToday(d: Date) { return fmt(d) === todayStr }
function isSelected(d: Date) { return props.modelValue === fmt(d) }

function selectDate(d: Date) {
  if (isPast(d)) return
  emit('update:modelValue', fmt(d))
  popoverOpen.value = false
}

function goToday() {
  emit('update:modelValue', todayStr)
  popoverOpen.value = false
}

function clear() {
  emit('update:modelValue', null)
  popoverOpen.value = false
}

function prevMonth() {
  if (viewMonth.value === 0) {
    viewMonth.value = 11
    viewYear.value--
  } else {
    viewMonth.value--
  }
}

function nextMonth() {
  if (viewMonth.value === 11) {
    viewMonth.value = 0
    viewYear.value++
  } else {
    viewMonth.value++
  }
}

function togglePopover() {
  popoverOpen.value = !popoverOpen.value
  if (popoverOpen.value) {
    // 打开时把视图对齐到当前选中日期(或今天)的月份
    const anchor = props.modelValue
      ? new Date(props.modelValue + 'T00:00:00')
      : today
    viewYear.value = anchor.getFullYear()
    viewMonth.value = anchor.getMonth()
  }
}

function onDocClick(e: MouseEvent) {
  if (!rootRef.value) return
  if (!rootRef.value.contains(e.target as Node)) {
    popoverOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
})
</script>

<template>
  <div ref="rootRef" class="date-chip-bar">
    <button
      type="button"
      class="picker"
      :class="{ active: props.modelValue, open: popoverOpen }"
      @click.stop="togglePopover"
    >
      <span class="picker-icon" aria-hidden="true">📅</span>
      <span class="picker-label">{{ displayLabel }}</span>
      <span class="picker-caret" :class="{ up: popoverOpen }" aria-hidden="true">▾</span>
    </button>
    <button
      v-if="props.modelValue"
      type="button"
      class="chip clear"
      @click.stop="clear"
    >清除</button>

    <Transition name="popover">
      <div v-if="popoverOpen" class="popover" role="dialog" aria-label="选择日期">
        <div class="cal-head">
          <button type="button" class="nav" @click="prevMonth" aria-label="上个月">‹</button>
          <span class="cal-title">{{ monthLabel }}</span>
          <button type="button" class="nav" @click="nextMonth" aria-label="下个月">›</button>
        </div>
        <div class="cal-weekdays">
          <span v-for="w in weekdayLabels" :key="w">{{ w }}</span>
        </div>
        <div class="cal-grid">
          <button
            v-for="(c, i) in calendarCells"
            :key="i"
            type="button"
            :class="['cell', {
              out: !c.inMonth,
              past: isPast(c.date),
              today: isToday(c.date),
              selected: isSelected(c.date),
            }]"
            :disabled="isPast(c.date)"
            @click="selectDate(c.date)"
          >{{ c.date.getDate() }}</button>
        </div>
        <div class="cal-footer">
          <button type="button" class="link" @click="goToday">今天</button>
          <button v-if="props.modelValue" type="button" class="link danger" @click="clear">清除</button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.date-chip-bar {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  position: relative;
}

.picker {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  padding: 7px 12px;
  background: var(--color-cream);
  color: var(--color-muted);
  border-radius: var(--radius-pill);
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
  border: 1.5px solid transparent;
  min-height: 36px;
  font-weight: 500;
}
.picker:hover { background: var(--color-pink-50); color: var(--color-pink-500); }
.picker.active { color: var(--color-pink-600); font-weight: 600; }
.picker.open {
  background: #fff;
  border-color: var(--color-pink-300);
  box-shadow: 0 0 0 3px rgba(248, 165, 194, 0.18);
  color: var(--color-pink-600);
}
:root[data-theme="dark"] .picker {
  background: var(--color-pink-100);
}
:root[data-theme="dark"] .picker.open {
  background: var(--color-cream);
}

.picker-icon { font-size: 14px; line-height: 1; }
.picker-caret {
  font-size: 10px;
  margin-left: 2px;
  transition: transform 0.2s var(--ease-spring);
  display: inline-block;
}
.picker-caret.up { transform: rotate(180deg); }

.chip.clear {
  font-size: 12px;
  padding: 5px 10px;
  background: transparent;
  color: var(--color-danger);
  border-radius: var(--radius-pill);
  text-decoration: underline;
}
.chip.clear:hover { background: var(--color-pink-50); }

.popover {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 80;
  background: #fff;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: 14px;
  min-width: 286px;
  border: 1px solid var(--color-line);
}
:root[data-theme="dark"] .popover {
  background: var(--color-cream);
  border-color: var(--color-line-2);
}

.cal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.cal-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-ink);
  font-family: var(--font-display);
}
.nav {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease, transform 0.15s var(--ease-spring);
  min-height: 0;
  min-width: 0;
  line-height: 1;
  padding: 0;
}
.nav:hover { background: var(--color-pink-100); }
.nav:active { transform: scale(0.9); }

.cal-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
  margin-bottom: 4px;
}
.cal-weekdays span {
  text-align: center;
  font-size: 11px;
  color: var(--color-muted);
  font-weight: 600;
  padding: 4px 0;
}

.cal-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}
.cell {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--color-ink);
  border-radius: 50%;
  background: transparent;
  font-weight: 500;
  transition: background 0.12s ease, color 0.12s ease, transform 0.12s var(--ease-spring);
  min-height: 0;
  min-width: 0;
  padding: 0;
  line-height: 1;
}
.cell:hover:not(:disabled):not(.selected) {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
}
.cell.out { color: var(--color-line-2); }
.cell.past { color: var(--color-line-2); cursor: not-allowed; }
.cell.today {
  border: 1.5px solid var(--color-pink-300);
  font-weight: 700;
}
.cell.selected {
  background: linear-gradient(135deg, var(--color-pink-400), var(--color-pink-500));
  color: #fff;
  font-weight: 700;
  box-shadow: 0 2px 6px rgba(236, 125, 166, 0.4);
}
.cell:active:not(:disabled) { transform: scale(0.88); }

.cal-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid var(--color-line);
}
.link {
  font-size: 12px;
  color: var(--color-pink-500);
  font-weight: 600;
  padding: 4px 10px;
  background: transparent;
  border-radius: var(--radius-sm);
  min-height: 0;
  min-width: 0;
  transition: background 0.12s ease;
}
.link:hover { background: var(--color-pink-50); }
.link.danger { color: var(--color-danger); }
.link.danger:hover { background: rgba(226, 109, 109, 0.08); }

.popover-enter-active, .popover-leave-active {
  transition: opacity 0.18s ease, transform 0.22s var(--ease-spring);
  transform-origin: top left;
}
.popover-enter-from, .popover-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
