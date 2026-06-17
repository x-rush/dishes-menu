import type { Dish, WeekMenu, Slot, Day, MenuItem, Todo, Together } from '../types'

// 同一个 BASE 兼容 dev / prod:dev 模式 BASE_URL = '/' → 走 Vite proxy 到 :8080;prod 模式 BASE_URL = '/forxt/dishes-menu/' → 拼到所有 /api/... 前面
const BASE = import.meta.env.DEV ? '' : import.meta.env.BASE_URL.replace(/\/$/, '')

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

  // Todos
  listTodos: () => request<Todo[]>('/api/todos'),
  createTodo: (input: { content: string; due_date: string | null; author_emoji: string; author_color: string }) =>
    request<Todo>('/api/todos', { method: 'POST', body: JSON.stringify(input) }),
  patchTodo: (id: number, body: { content?: string; completed?: boolean; due_date?: string | null }) =>
    request<Todo>(`/api/todos/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  deleteTodo: (id: number) =>
    request<void>(`/api/todos/${id}`, { method: 'DELETE' }),

  // Together counter
  getTogether: () => request<Together>('/api/together'),
  setTogether: (date: string) =>
    request<void>('/api/together', { method: 'POST', body: JSON.stringify({ date }) }),
}
