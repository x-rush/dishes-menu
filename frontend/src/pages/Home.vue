<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useMenuStore } from '../stores/menu'
import { useDishesStore } from '../stores/dishes'
import WeekTabs from '../components/WeekTabs.vue'
import DishCard from '../components/DishCard.vue'
import AddDishDialog from '../components/AddDishDialog.vue'
import DishPickerSheet from '../components/DishPickerSheet.vue'
import Mascot from '../components/illustrations/Mascot.vue'
import { slotsForDay, DAY_LABELS, type Day, type Slot } from '../types'

const menu = useMenuStore()
const dishes = useDishesStore()
const { currentWeek, currentDay, currentDayMenu, loading, error } = storeToRefs(menu)

const addOpen = ref(false)
const reloadHint = ref<string | null>(null)

// Picker 协调:Home 持有目标,DishCard 触发打开,Sheet 消费
const pickerTarget = ref<{ day: Day; slot: Slot } | null>(null)
function openPicker(day: Day, slot: Slot) {
  pickerTarget.value = { day, slot }
}
function closePicker() {
  pickerTarget.value = null
}

const slots = computed(() => slotsForDay(currentDay.value))

const dayLabel = computed(() => `${DAY_LABELS[currentDay.value]}`)
const weekLabel = computed(() => currentWeek.value)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 5)  return '夜深了,先休息吧 🌙'
  if (h < 11) return '今天吃点什么呢? ☀️'
  if (h < 14) return '午餐想好做什么了吗? 🍱'
  if (h < 18) return '下午茶来点啥呀? ☕'
  if (h < 22) return '晚餐要不要换换口味? 🌸'
  return '夜深了,先休息吧 🌙'
})

async function refresh() {
  reloadHint.value = '刷新中…'
  try {
    await Promise.all([
      dishes.load(true),
      menu.loadWeek(currentWeek.value, true),
    ])
    reloadHint.value = '已更新 ✓'
    setTimeout(() => { reloadHint.value = null }, 1500)
  } catch {
    reloadHint.value = null
  }
}

onMounted(async () => {
  try {
    await Promise.all([
      dishes.load(),
      menu.loadWeek(currentWeek.value),
    ])
  } catch (e) {
    console.error('initial load failed', e)
  }
})
</script>

<template>
  <div class="home">
    <header class="hero">
      <div class="hero-left">
        <Mascot :size="72" class="hero-mascot" />
        <div class="hero-text">
          <h1 class="hero-greeting">{{ greeting }}</h1>
          <p class="hero-sub">{{ weekLabel }} · {{ dayLabel }}</p>
        </div>
      </div>
      <button class="btn btn-icon refresh-btn" aria-label="刷新" @click="refresh" :disabled="loading">
        <span v-if="loading" class="spinner" aria-hidden="true"></span>
        <span v-else>⟳</span>
      </button>
      <Transition name="hint">
        <p v-if="reloadHint" class="reload-hint">{{ reloadHint }}</p>
      </Transition>
    </header>

    <WeekTabs />

    <main class="slots">
      <p v-if="error" class="error-banner">{{ error }}</p>

      <div v-if="loading && (currentDayMenu.breakfast?.length ?? 0) === 0 && (currentDayMenu.lunch?.length ?? 0) === 0 && (currentDayMenu.dinner?.length ?? 0) === 0 && (currentDayMenu.snack?.length ?? 0) === 0" class="empty">
        <span class="spinner"></span>
        <p>加载中…</p>
      </div>

      <template v-else>
        <DishCard
          v-for="(slot, i) in slots"
          :key="slot"
          :week="currentWeek"
          :day="currentDay"
          :slot="slot"
          :items="currentDayMenu[slot] ?? []"
          :style="{ '--enter-delay': i * 70 + 'ms' }"
          @pick="openPicker(currentDay, slot)"
        />
      </template>
    </main>

    <button class="fab" aria-label="添加菜品" @click="addOpen = true">+</button>

    <AddDishDialog :open="addOpen" @close="addOpen = false" @saved="dishes.load(true)" />
    <DishPickerSheet
      :open="pickerTarget !== null"
      :week="currentWeek"
      :day="pickerTarget?.day ?? currentDay"
      :slot="pickerTarget?.slot ?? 'breakfast'"
      :picked-dish-ids="pickerTarget ? (currentDayMenu[pickerTarget.slot] ?? []).map((it) => it.dish_id) : []"
      @close="closePicker"
    />
  </div>
</template>

<style scoped>
.home {
  max-width: 640px;
  margin: 0 auto;
  padding-bottom: 96px;
  position: relative;
}

.hero {
  position: sticky;
  top: 0;
  z-index: 10;
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-rows: auto auto;
  gap: 4px 12px;
  align-items: center;
  padding: 18px 16px 12px;
  background: linear-gradient(180deg, var(--color-pink-50) 0%, var(--color-warm-bg) 100%);
  border-bottom: 1px solid var(--color-line);
}

.hero-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.hero-mascot {
  color: var(--color-pink-400);
  flex: 0 0 auto;
}

.hero-text {
  min-width: 0;
}

.hero-greeting {
  font-size: 19px;
  font-weight: 700;
  color: var(--color-ink);
  margin: 0;
  font-family: var(--font-display);
  letter-spacing: 0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.hero-sub {
  font-size: 12px;
  color: var(--color-muted);
  margin-top: 2px;
  font-weight: 500;
}

.refresh-btn {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  font-size: 18px;
  width: 40px;
  height: 40px;
  min-height: 40px;
  min-width: 40px;
  border-radius: 50%;
}

.reload-hint {
  grid-column: 1 / -1;
  font-size: 12px;
  color: var(--color-success);
  margin-top: 2px;
}

.hint-enter-active, .hint-leave-active {
  transition: opacity 0.2s ease, transform 0.2s var(--ease-spring);
}
.hint-enter-from, .hint-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.slots {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.error-banner {
  background: rgba(226, 109, 109, 0.1);
  color: var(--color-danger);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  border-left: 3px solid var(--color-danger);
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 16px;
}
</style>
