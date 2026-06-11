<script setup lang="ts">
import { onMounted, ref, watch, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useMenuStore } from '../stores/menu'
import { ALL_DAYS, DAY_LABELS, type Day } from '../types'

const menu = useMenuStore()
const { currentDay } = storeToRefs(menu)
const scrollerRef = ref<HTMLElement | null>(null)

function pick(day: Day) {
  menu.setDay(day)
}

async function scrollToActive() {
  await nextTick()
  const el = scrollerRef.value?.querySelector<HTMLElement>(`[data-day="${currentDay.value}"]`)
  el?.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' })
}

onMounted(() => {
  scrollToActive()
})

watch(currentDay, () => {
  scrollToActive()
})
</script>

<template>
  <nav class="week-tabs" ref="scrollerRef" aria-label="选择星期">
    <button
      v-for="day in ALL_DAYS"
      :key="day"
      :data-day="day"
      :class="['day-tab', { active: currentDay === day }]"
      :aria-pressed="currentDay === day"
      @click="pick(day)"
    >
      {{ DAY_LABELS[day] }}
    </button>
  </nav>
</template>

<style scoped>
.week-tabs {
  display: flex;
  gap: 8px;
  padding: 8px 16px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}
.week-tabs::-webkit-scrollbar { display: none; }

.day-tab {
  flex: 0 0 auto;
  scroll-snap-align: center;
  padding: 10px 18px;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--color-muted);
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-display);
  transition: background 0.18s var(--ease-out-soft), color 0.15s ease, transform 0.15s var(--ease-spring);
  white-space: nowrap;
}
.day-tab:active { transform: scale(0.96); }

.day-tab.active {
  background: var(--color-pink-400);
  color: #fff;
  font-weight: 600;
  box-shadow: var(--shadow-sm);
}

.day-tab:not(.active):hover {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
}
</style>
