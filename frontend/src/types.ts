export type Slot = 'breakfast' | 'lunch' | 'dinner' | 'snack'

export interface Dish {
  id: string
  name: string
  slots: Slot[]
  ingredients: string[]
  estimated_time: number
  note: string
  image?: string
  created_at: string
  updated_at: string
}

export interface MenuItem {
  seq: number
  dish_id: string
  note: string
}

export interface DayMenu {
  breakfast: MenuItem[]
  lunch: MenuItem[]
  dinner: MenuItem[]
  snack: MenuItem[]
}

export interface WeekMenu {
  mon: DayMenu
  tue: DayMenu
  wed: DayMenu
  thu: DayMenu
  fri: DayMenu
  sat: DayMenu
  sun: DayMenu
}

export const ALL_DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const
export type Day = typeof ALL_DAYS[number]

export const ALL_SLOTS: Slot[] = ['breakfast', 'lunch', 'dinner', 'snack']

export const SLOT_LABELS: Record<Slot, string> = {
  breakfast: '早餐',
  lunch: '午餐',
  dinner: '晚餐',
  snack: '加餐',
}

export const DAY_LABELS: Record<Day, string> = {
  mon: '周一',
  tue: '周二',
  wed: '周三',
  thu: '周四',
  fri: '周五',
  sat: '周六',
  sun: '周日',
}

export function slotsForDay(day: Day): Slot[] {
  return day === 'sat' || day === 'sun'
    ? ['breakfast', 'lunch', 'dinner']
    : ['breakfast', 'snack']
}
