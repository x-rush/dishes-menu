import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, APIError } from '../api/client'
import type { Todo } from '../types'

export const useTodoStore = defineStore('todo', () => {
  const todos = ref<Todo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const open = computed(() => todos.value.filter(t => !t.completed_at))
  const done = computed(() => todos.value.filter(t => !!t.completed_at))

  async function fetchAll(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      todos.value = await api.listTodos()
    } catch (e) {
      error.value = e instanceof APIError ? e.message : '加载待办失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function create(input: { content: string; due_date: string | null; author_emoji: string; author_color: string }): Promise<Todo> {
    const created = await api.createTodo(input)
    todos.value = [created, ...todos.value]  // 最新在前
    return created
  }

  async function toggle(id: number): Promise<void> {
    const current = todos.value.find(t => t.id === id)
    if (!current) return
    // 显式发送目标状态 (后端会 toggle,但 FE 把意图写明)
    const updated = await api.patchTodo(id, { completed: !current.completed_at })
    todos.value = todos.value.map(t => (t.id === id ? updated : t))
  }

  async function updateContent(id: number, content: string): Promise<void> {
    const updated = await api.patchTodo(id, { content })
    todos.value = todos.value.map(t => (t.id === id ? updated : t))
  }

  async function remove(id: number): Promise<void> {
    await api.deleteTodo(id)
    todos.value = todos.value.filter(t => t.id !== id)
  }

  return { todos, loading, error, open, done, fetchAll, create, toggle, updateContent, remove }
})