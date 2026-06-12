<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useMenuStore } from '../stores/menu'
import { useDishesStore } from '../stores/dishes'
import WeekTabs from '../components/WeekTabs.vue'
import DishCard from '../components/DishCard.vue'
import AddDishDialog from '../components/AddDishDialog.vue'
import DishPickerSheet from '../components/DishPickerSheet.vue'
import Mascot from '../components/illustrations/Mascot.vue'
import { slotsForDay, ALL_DAYS, DAY_LABELS, type Day, type Slot, type Dish } from '../types'
import { isoWeekKey, dateToDayKey, dayToDate, formatDateString, parseDateString } from '../utils/isoWeek'
import { useSwipeDay } from '../composables/useSwipeDay'
import { useUndo } from '../composables/useUndo'
import SkeletonCard from '../components/SkeletonCard.vue'
import ConfettiBurst from '../components/ConfettiBurst.vue'

const menu = useMenuStore()
const dishes = useDishesStore()
const undo = useUndo()
const router = useRouter()
const { currentDay, currentDayMenu, loading, error } = storeToRefs(menu)

const route = useRoute()

/** URL 周次(校验后回退到本周) */
const week = computed<string>(() => {
  const w = route.params.week
  return typeof w === 'string' && /^202\d-W\d{2}$/.test(w) ? w : isoWeekKey()
})

/** URL 日期(YYYY-MM-DD) — 路由同步切 day 的 source of truth */
const date = computed<string>(() => {
  const d = route.params.date
  if (typeof d === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(d) && parseDateString(d)) {
    return d
  }
  return formatDateString(new Date())
})

/** date → day key(派生,跟 store 无关;路由里只存 date) */
const day = computed<Day>(() => {
  const k = dateToDayKey(date.value)
  return (k ?? 'mon') as Day
})

/** 切 day 方向(给动画用) */
const slideDir = ref<'left' | 'right' | null>(null)

/** 路由变化 → 同步到 store + 拉数据(week 变就重拉,day 变只切不重拉) */
watch(
  week,
  async (newWeek) => {
    menu.setWeek(newWeek)
    menu.setDay(day.value)
    try {
      await menu.loadWeek(newWeek)
    } catch (e) {
      console.error('load week failed', e)
    }
  },
  { immediate: true }
)

/** day 变 → 同步 store,但不重拉数据 */
watch(day, (newDay) => {
  menu.setDay(newDay)
})

const addOpen = ref(false)
const reloadHint = ref<string | null>(null)

const homeRef = ref<HTMLElement | null>(null)

/** 滑到下一天 / 上一天 → router.push(切 day 走 URL,不切 week) */
function gotoNextDay() {
  slideDir.value = 'left'
  const d = parseDateString(date.value)
  if (!d) return
  d.setDate(d.getDate() + 1)
  const newDate = formatDateString(d)
  // 跨周:week 用新 date 重新算
  const newWeek = isoWeekKey(d)
  router.push(`/${newWeek}/${newDate}`)
}
function gotoPrevDay() {
  slideDir.value = 'right'
  const d = parseDateString(date.value)
  if (!d) return
  d.setDate(d.getDate() - 1)
  const newDate = formatDateString(d)
  const newWeek = isoWeekKey(d)
  router.push(`/${newWeek}/${newDate}`)
}
useSwipeDay(homeRef, { onNext: gotoNextDay, onPrev: gotoPrevDay })

const pickerTarget = ref<{ day: Day; slot: Slot } | null>(null)
function openPicker(day: Day, slot: Slot) {
  pickerTarget.value = { day, slot }
}
function closePicker() {
  pickerTarget.value = null
}

const slots = computed(() => slotsForDay(day.value))

const dayLabel = computed(() => `${DAY_LABELS[day.value]}`)
const weekLabel = computed(() => week.value)

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
      menu.loadWeek(week.value, true),
    ])
    reloadHint.value = '已更新 ✓'
    setTimeout(() => { reloadHint.value = null }, 1500)
  } catch {
    reloadHint.value = null
  }
}

/** 整周 16 个 slot 是否全部填满(5×2 工作日 + 2×3 周末 = 16) */
const weekFull = computed(() => {
  const wm = menu.weekMenus[week.value]
  if (!wm) return false
  return ALL_DAYS.every((d) => {
    const slots = slotsForDay(d)
    return slots.every((s) => (wm[d]?.[s]?.length ?? 0) > 0)
  })
})

const confettiShow = ref(false)
const confettiShownForWeek = ref<string | null>(null)

watch(week, () => {
  // 切到新一周:允许重新撒花
  confettiShownForWeek.value = null
})

watch(weekFull, (full) => {
  if (full && confettiShownForWeek.value !== week.value) {
    confettiShownForWeek.value = week.value
    confettiShow.value = true
    setTimeout(() => { confettiShow.value = false }, 2000)
  }
})

function onDishCreated(d: Dish) {
  undo.push(`已加入「${d.name}」到菜品库`, async () => {
    try {
      await dishes.remove(d.id)
    } catch (e) {
      console.error('undo create dish failed', e)
    }
  })
}

/** WeekTabs 点击 day 也走 URL(统一入口) */
function onChangeDay(d: Day) {
  const newDate = dayToDate(week.value, d)
  slideDir.value = null  // tab 点击不动画
  router.push(`/${week.value}/${newDate}`)
}

onMounted(async () => {
  try {
    await dishes.load()
  } catch (e) {
    console.error('initial dishes load failed', e)
  }
})
</script>

<template>
  <div ref="homeRef" class="home">
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

    <WeekTabs @change-day="onChangeDay" />

    <main class="slots">
      <p v-if="error" class="error-banner">{{ error }}</p>

      <Transition
        :name="slideDir ? `day-slide-${slideDir}` : 'day-fade'"
        mode="out-in"
      >
        <div :key="date" class="day-content">
          <template v-if="loading && slots.every((s) => (currentDayMenu[s]?.length ?? 0) === 0)">
            <SkeletonCard v-for="i in Math.max(slots.length, 2)" :key="`skel-${i}`" :style="{ '--enter-delay': (i - 1) * 70 + 'ms' }" />
          </template>

          <template v-else>
            <DishCard
              v-for="(slot, i) in slots"
              :key="slot"
              :week="week"
              :day="day"
              :slot="slot"
              :items="currentDayMenu[slot] ?? []"
              :style="{ '--enter-delay': i * 70 + 'ms' }"
              @pick="openPicker(day, slot)"
            />
          </template>
        </div>
      </Transition>
    </main>

    <button class="fab" aria-label="添加菜品" @click="addOpen = true">+</button>

    <AddDishDialog :open="addOpen" @close="addOpen = false" @saved="dishes.load(true)" @created="onDishCreated" />
    <DishPickerSheet
      :open="pickerTarget !== null"
      :week="week"
      :day="pickerTarget?.day ?? day"
      :slot="pickerTarget?.slot ?? 'breakfast'"
      :picked-dish-ids="pickerTarget ? (currentDayMenu[pickerTarget.slot] ?? []).map((it) => it.dish_id) : []"
      @close="closePicker"
    />

    <ConfettiBurst :show="confettiShow" />
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
  overflow: hidden;
}

.day-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* 切 day 动画:左滑(下一天)= 内容从右进、从左出 */
.day-slide-left-enter-active,
.day-slide-left-leave-active,
.day-slide-right-enter-active,
.day-slide-right-leave-active,
.day-fade-enter-active,
.day-fade-leave-active {
  transition: transform 0.32s var(--ease-out-soft), opacity 0.18s ease;
  will-change: transform, opacity;
}

.day-slide-left-enter-from { transform: translateX(50px); opacity: 0; }
.day-slide-left-leave-to   { transform: translateX(-50px); opacity: 0; }

.day-slide-right-enter-from { transform: translateX(-50px); opacity: 0; }
.day-slide-right-leave-to   { transform: translateX(50px); opacity: 0; }

.day-fade-enter-from { opacity: 0; transform: translateY(6px); }
.day-fade-leave-to   { opacity: 0; transform: translateY(-6px); }

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
