import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, APIError } from '../api/client'
import type { Dish } from '../types'

export const useDishesStore = defineStore('dishes', () => {
  const dishes = ref<Dish[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const byId = computed<Record<string, Dish>>(() => {
    const m: Record<string, Dish> = {}
    for (const d of dishes.value) m[d.id] = d
    return m
  })

  async function load(force = false): Promise<void> {
    if (!force && dishes.value.length > 0) return
    loading.value = true
    error.value = null
    try {
      const res = await api.listDishes()
      dishes.value = res.dishes
    } catch (e) {
      error.value = e instanceof APIError ? e.message : '加载菜品失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function create(input: Omit<Dish, 'id' | 'created_at' | 'updated_at'>): Promise<Dish> {
    const created = await api.createDish(input)
    dishes.value = [...dishes.value, created]
    return created
  }

  async function update(id: string, input: Omit<Dish, 'id' | 'created_at' | 'updated_at'>): Promise<Dish> {
    const updated = await api.updateDish(id, input)
    dishes.value = dishes.value.map((d) => (d.id === id ? updated : d))
    return updated
  }

  async function remove(id: string): Promise<void> {
    await api.deleteDish(id)
    dishes.value = dishes.value.filter((d) => d.id !== id)
  }

  function getById(id: string): Dish | undefined {
    return byId.value[id]
  }

  return {
    dishes,
    loading,
    error,
    byId,
    load,
    create,
    update,
    remove,
    getById,
  }
})
