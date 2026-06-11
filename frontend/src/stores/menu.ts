import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { api, APIError } from '../api/client'
import { isoWeekKey, dayKeyForDate } from '../utils/isoWeek'
import type { Day, DayMenu, Slot, WeekMenu, Dish } from '../types'

const emptyDayMenu = (): DayMenu => ({
  breakfast: [],
  lunch: [],
  dinner: [],
  snack: [],
})

const emptyWeekMenu = (): WeekMenu => ({
  mon: emptyDayMenu(), tue: emptyDayMenu(), wed: emptyDayMenu(), thu: emptyDayMenu(),
  fri: emptyDayMenu(), sat: emptyDayMenu(), sun: emptyDayMenu(),
})

export const useMenuStore = defineStore('menu', () => {
  const currentWeek = ref<string>(isoWeekKey())
  const currentDay = ref<Day>(dayKeyForDate())
  const weekMenus = reactive<Record<string, WeekMenu>>({})
  const loading = ref(false)
  const error = ref<string | null>(null)

  const currentWeekMenu = computed<WeekMenu>(() => {
    return weekMenus[currentWeek.value] ?? emptyWeekMenu()
  })

  const currentDayMenu = computed<DayMenu>(() => {
    return currentWeekMenu.value[currentDay.value] ?? emptyDayMenu()
  })

  function ensureWeek(week: string): WeekMenu {
    if (!weekMenus[week]) weekMenus[week] = emptyWeekMenu()
    return weekMenus[week]
  }

  async function loadWeek(week: string, force = false): Promise<void> {
    if (!force && weekMenus[week]) return
    loading.value = true
    error.value = null
    try {
      const res = await api.getWeek(week)
      weekMenus[res.week] = res.menu
    } catch (e) {
      error.value = e instanceof APIError ? e.message : '加载菜单失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function appendItem(
    week: string,
    day: Day,
    slot: Slot,
    dishId: string,
    note: string
  ): Promise<void> {
    const res = await api.appendItem(week, day, slot, dishId, note)
    const wm = ensureWeek(res.week)
    const list = wm[day][slot] ?? []
    list.push({ seq: res.item.seq, dish_id: res.item.dish_id, note: res.item.note })
    wm[day][slot] = list
  }

  async function removeItem(
    week: string,
    day: Day,
    slot: Slot,
    seq: number
  ): Promise<void> {
    await api.deleteItem(week, day, slot, seq)
    const wm = ensureWeek(week)
    const list = wm[day][slot] ?? []
    // 后端已重排 seq;前端按数组顺序重排,保持稳定
    wm[day][slot] = list
      .filter((it) => it.seq !== seq)
      .map((it, i) => ({ ...it, seq: i }))
  }

  async function updateItemNote(
    week: string,
    day: Day,
    slot: Slot,
    seq: number,
    note: string
  ): Promise<void> {
    await api.updateItemNote(week, day, slot, seq, note)
    const wm = ensureWeek(week)
    const list = wm[day][slot] ?? []
    const it = list.find((x) => x.seq === seq)
    if (it) it.note = note
  }

  async function shuffleDish(week: string, day: Day, slot: Slot): Promise<Dish> {
    const res = await api.shuffle(week, day, slot)
    await appendItem(week, day, slot, res.dish.id, '')
    return res.dish
  }

  function setDay(day: Day): void {
    currentDay.value = day
  }

  function setWeek(week: string): void {
    currentWeek.value = week
  }

  return {
    currentWeek,
    currentDay,
    weekMenus,
    loading,
    error,
    currentWeekMenu,
    currentDayMenu,
    loadWeek,
    appendItem,
    removeItem,
    updateItemNote,
    shuffleDish,
    setDay,
    setWeek,
  }
})
