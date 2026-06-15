<script setup lang="ts">
const props = defineProps<{ modelValue: string | null }>()
const emit = defineEmits<{ 'update:modelValue': [string | null] }>()

function fmt(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const today = new Date()
const todayStr = fmt(today)

const tomorrow = new Date(today); tomorrow.setDate(tomorrow.getDate() + 1)
const tomorrowStr = fmt(tomorrow)

const dayAfter = new Date(today); dayAfter.setDate(dayAfter.getDate() + 2)
const dayAfterStr = fmt(dayAfter)

const nextWeek = new Date(today); nextWeek.setDate(nextWeek.getDate() + 7)
const nextWeekStr = fmt(nextWeek)

const chips = [
  { label: '今天', value: todayStr },
  { label: '明天', value: tomorrowStr },
  { label: '后天', value: dayAfterStr },
  { label: '下周', value: nextWeekStr },
]

function pick(v: string) {
  // 再点同一个 → 取消
  emit('update:modelValue', props.modelValue === v ? null : v)
}

function clear() {
  emit('update:modelValue', null)
}
</script>

<template>
  <div class="date-chip-bar">
    <button
      v-for="c in chips"
      :key="c.value"
      type="button"
      :class="['chip', { active: props.modelValue === c.value }]"
      @click="pick(c.value)"
    >{{ c.label }} · {{ c.value.slice(5) }}</button>
    <button
      v-if="props.modelValue"
      type="button"
      class="chip clear"
      @click="clear"
    >清除</button>
  </div>
</template>

<style scoped>
.date-chip-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.chip {
  font-size: 12px;
  padding: 5px 10px;
  background: var(--color-cream);
  color: var(--color-muted);
  border-radius: var(--radius-pill);
  transition: background 0.15s ease, color 0.15s ease;
}
.chip:hover { background: var(--color-pink-50); color: var(--color-pink-500); }
.chip.active {
  background: var(--color-pink-200);
  color: var(--color-pink-600);
  font-weight: 600;
}
.chip.clear {
  background: transparent;
  color: var(--color-danger);
  text-decoration: underline;
}
</style>
