import type { Dish, WeekMenu, Slot, Day, MenuItem } from '../types'

const BASE = '' // same-origin in prod; Vite proxy in dev

export class APIError extends Error {
  constructor(public code: string, message: string, public status: number) {
    super(message)
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })
  if (!res.ok) {
    let code = 'HTTP_' + res.status
    let message = res.statusText
    try {
      const body = await res.json()
      if (body?.error) {
        code = body.error.code ?? code
        message = body.error.message ?? message
      }
    } catch {
      /* ignore */
    }
    throw new APIError(code, message, res.status)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export const api = {
  // Dishes
  listDishes: () => request<{ dishes: Dish[] }>('/api/dishes'),
  createDish: (d: Omit<Dish, 'id' | 'created_at' | 'updated_at'>) =>
    request<Dish>('/api/dishes', { method: 'POST', body: JSON.stringify(d) }),
  updateDish: (id: string, d: Omit<Dish, 'id' | 'created_at' | 'updated_at'>) =>
    request<Dish>(`/api/dishes/${id}`, { method: 'PUT', body: JSON.stringify(d) }),
  deleteDish: (id: string) =>
    request<void>(`/api/dishes/${id}`, { method: 'DELETE' }),

  // Menu
  getWeek: (week: string) =>
    request<{ week: string; menu: WeekMenu }>(`/api/menu?week=${encodeURIComponent(week)}`),
  appendItem: (week: string, day: Day, slot: Slot, dishId: string, note: string) =>
    request<{ week: string; day: Day; slot: Slot; item: MenuItem }>(
      `/api/menu/${day}/${slot}?week=${encodeURIComponent(week)}`,
      { method: 'POST', body: JSON.stringify({ dish_id: dishId, note }) }
    ),
  updateItemNote: (week: string, day: Day, slot: Slot, seq: number, note: string) =>
    request<{ week: string; day: Day; slot: Slot; seq: number; note: string }>(
      `/api/menu/${day}/${slot}/${seq}?week=${encodeURIComponent(week)}`,
      { method: 'PUT', body: JSON.stringify({ note }) }
    ),
  deleteItem: (week: string, day: Day, slot: Slot, seq: number) =>
    request<{ week: string; day: Day; slot: Slot; seq: number }>(
      `/api/menu/${day}/${slot}/${seq}?week=${encodeURIComponent(week)}`,
      { method: 'DELETE' }
    ),
  shuffle: (week: string, day: Day, slot: Slot) => {
    const params = new URLSearchParams({ week, day, slot })
    return request<{ dish: Dish; week: string; day: Day; slot: Slot }>(
      `/api/menu/shuffle?${params.toString()}`
    )
  },
}
