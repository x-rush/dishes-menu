import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, APIError } from '../api/client'
import type { Todo, TodoComment } from '../types'

export const useTodoStore = defineStore('todo', () => {
  const todos = ref<Todo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // ── 评论(按 todo_id 缓存,打开 modal 时按需拉,关闭不清理以便来回切换) ──
  const commentsByTodo = ref<Record<number, TodoComment[]>>({})
  const commentsLoading = ref<Record<number, boolean>>({})

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

  async function updateDueDate(id: number, dueDate: string | null): Promise<void> {
    const updated = await api.patchTodo(id, { due_date: dueDate })
    todos.value = todos.value.map(t => (t.id === id ? updated : t))
  }

  async function togglePin(id: number): Promise<void> {
    const current = todos.value.find(t => t.id === id)
    if (!current) return
    // FE 写明目标状态(后端 TogglePin 用 `pinned = NOT pinned` 原子反转,值会被忽略,
    // 但传显式值的好处是 audit log / 未来兼容 read-modify-write 都有清晰意图)
    const updated = await api.patchTodo(id, { pinned: !current.pinned })
    todos.value = todos.value.map(t => (t.id === id ? updated : t))
  }

  async function remove(id: number): Promise<void> {
    await api.deleteTodo(id)
    todos.value = todos.value.filter(t => t.id !== id)
    // 清理缓存中的评论(todo_comments 有 FK CASCADE,服务端已删,本地也不留)
    delete commentsByTodo.value[id]
    delete commentsLoading.value[id]
  }

  // ── 评论 ──
  async function fetchComments(todoId: number): Promise<void> {
    commentsLoading.value[todoId] = true
    try {
      const list = await api.listComments(todoId)
      commentsByTodo.value = { ...commentsByTodo.value, [todoId]: list }
    } finally {
      commentsLoading.value[todoId] = false
    }
  }

  async function addComment(todoId: number, input: { content: string; author_emoji: string; author_color: string }): Promise<void> {
    const created = await api.addComment(todoId, input)
    const existing = commentsByTodo.value[todoId] ?? []
    commentsByTodo.value = { ...commentsByTodo.value, [todoId]: [...existing, created] }
  }

  return {
    todos, loading, error,
    open, done,
    commentsByTodo, commentsLoading,
    fetchAll, create, toggle, updateContent, updateDueDate, togglePin, remove,
    fetchComments, addComment,
  }
})