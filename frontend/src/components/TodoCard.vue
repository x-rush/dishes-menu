<script setup lang="ts">
import { ref } from 'vue'
import type { Todo } from '../types'
import { useTodoStore } from '../stores/todo'
import { useUndo } from '../composables/useUndo'

const props = defineProps<{ todo: Todo }>()
const store = useTodoStore()
const undo = useUndo()
const editing = ref(false)
const editText = ref(props.todo.content)
const showActions = ref(false)
let pressTimer: number | null = null

async function toggle() {
  await store.toggle(props.todo.id)
  if (!props.todo.completed_at) {
    undo.push(`已勾上「${props.todo.content.slice(0, 12)}」`, async () => {
      await store.toggle(props.todo.id)
    })
  }
}

function startEdit() {
  editing.value = true
  editText.value = props.todo.content
}

async function saveEdit() {
  const trimmed = editText.value.trim()
  if (!trimmed || trimmed === props.todo.content) {
    editing.value = false
    return
  }
  await store.updateContent(props.todo.id, trimmed)
  editing.value = false
}

async function remove() {
  showActions.value = false
  const snapshot = { ...props.todo }
  await store.remove(props.todo.id)
  undo.push(`已删除「${snapshot.content.slice(0, 12)}」`, async () => {
    await store.create({
      content: snapshot.content,
      due_date: snapshot.due_date,
      author_emoji: snapshot.author_emoji,
      author_color: snapshot.author_color,
    })
  })
}

function onPressStart() {
  pressTimer = window.setTimeout(() => {
    showActions.value = true
  }, 520)
}
function onPressEnd() {
  if (pressTimer !== null) {
    clearTimeout(pressTimer)
    pressTimer = null
  }
}

function formatDue(d: string | null) {
  if (!d) return ''
  return d.slice(5)
}

function relativeTime(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const min = Math.floor(ms / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h} 小时前`
  const d = Math.floor(h / 24)
  return `${d} 天前`
}
</script>

<template>
  <article
    :class="['todo-card', { done: todo.completed_at, showActions }]"
    @touchstart.passive="onPressStart"
    @touchend.passive="onPressEnd"
    @touchcancel.passive="onPressEnd"
    @mousedown="onPressStart"
    @mouseup="onPressEnd"
    @mouseleave="onPressEnd"
  >
    <button
      type="button"
      class="check"
      :aria-pressed="!!todo.completed_at"
      :aria-label="todo.completed_at ? '取消完成' : '标记完成'"
      @click="toggle"
    >
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
        <path
          d="M5 12.5l4.5 4.5L19 7.5"
          fill="none"
          stroke="currentColor"
          stroke-width="3"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="check-path"
        />
      </svg>
    </button>

    <div class="body">
      <div v-if="editing" class="edit-row">
        <input
          v-model="editText"
          maxlength="500"
          @keydown.enter="saveEdit"
          @keydown.esc="editing = false"
        />
        <button class="btn btn-primary save" @click="saveEdit">保存</button>
      </div>
      <div v-else class="content-row">
        <p class="content" @click="startEdit">{{ todo.content }}</p>
      </div>

      <div class="meta">
        <span
          class="sig"
          :style="{ background: todo.author_color }"
          :title="`由 ${todo.author_emoji} 添加`"
        >{{ todo.author_emoji }}</span>
        <span v-if="todo.due_date" class="due">📅 {{ formatDue(todo.due_date) }}</span>
        <span class="time">{{ relativeTime(todo.created_at) }}</span>
      </div>
    </div>

    <Transition name="actions">
      <div v-if="showActions" class="action-overlay" @click="showActions = false">
        <button class="action-btn danger" @click.stop="remove">删除这条</button>
        <button class="action-btn ghost" @click.stop="showActions = false">取消</button>
      </div>
    </Transition>
  </article>
</template>

<style scoped>
.todo-card {
  position: relative;
  display: flex;
  gap: 12px;
  padding: 14px 16px;
  background: linear-gradient(135deg, #fff8f3 0%, #fff0f5 100%);
  border-radius: var(--radius-lg);
  align-items: flex-start;
  box-shadow: var(--shadow-sm);
  transition:
    opacity 0.3s ease,
    transform 0.2s var(--ease-spring),
    box-shadow 0.25s var(--ease-out-soft);
  overflow: hidden;
}
.todo-card:active { transform: scale(0.99); }
:root[data-theme="dark"] .todo-card {
  background: linear-gradient(135deg, #2d2226 0%, #3a2830 100%);
}

.todo-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--color-pink-300);
  opacity: 0.5;
}
.todo-card.done {
  background: var(--color-warm-bg);
  box-shadow: none;
}
:root[data-theme="dark"] .todo-card.done {
  background: #241b1f;
}
.todo-card.done::before { opacity: 0.2; }
.todo-card.done .content {
  text-decoration: line-through;
  text-decoration-color: var(--color-pink-300);
  text-decoration-thickness: 1.5px;
  color: var(--color-muted);
}
.todo-card.done .meta { opacity: 0.7; }

.check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid var(--color-pink-300);
  background: transparent;
  color: #fff;
  flex: 0 0 auto;
  margin-top: 1px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition:
    background 0.2s var(--ease-spring),
    border-color 0.2s ease,
    transform 0.15s var(--ease-spring);
}
.check:hover { border-color: var(--color-pink-500); }
.check:active { transform: scale(0.88); }
.todo-card.done .check {
  background: linear-gradient(135deg, var(--color-pink-400), var(--color-pink-500));
  border-color: var(--color-pink-500);
  box-shadow: 0 2px 6px rgba(236, 125, 166, 0.35);
}
.check-path {
  stroke-dasharray: 24;
  stroke-dashoffset: 24;
  transition: stroke-dashoffset 0.3s var(--ease-out-soft);
}
.todo-card.done .check-path { stroke-dashoffset: 0; }

.body { flex: 1 1 auto; min-width: 0; }
.content-row { display: flex; gap: 8px; align-items: center; }
.content {
  font-size: 15px;
  line-height: 1.55;
  word-break: break-word;
  cursor: text;
  margin: 0;
  color: var(--color-ink);
}
.edit-row { display: flex; gap: 6px; }
.edit-row input { flex: 1 1 auto; font-size: 14px; padding: 6px 10px; }
.save { padding: 4px 12px; font-size: 12px; }

.meta {
  display: flex;
  gap: 8px;
  margin-top: 8px;
  font-size: 11.5px;
  color: var(--color-muted);
  align-items: center;
  flex-wrap: wrap;
}
.sig {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.06), inset 0 -1px 2px rgba(0,0,0,0.08);
  flex: 0 0 auto;
}
.due {
  color: var(--color-peach-300);
  background: rgba(255, 169, 128, 0.12);
  padding: 2px 7px;
  border-radius: var(--radius-pill);
  font-weight: 500;
}
.time { color: var(--color-muted); }

.action-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 248, 243, 0.96);
  -webkit-backdrop-filter: blur(4px);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  z-index: 5;
  animation: fadeIn 0.15s ease;
}
:root[data-theme="dark"] .action-overlay {
  background: rgba(45, 34, 38, 0.96);
}
.action-btn {
  padding: 8px 18px;
  border-radius: var(--radius-pill);
  font-size: 14px;
  font-weight: 600;
  transition: transform 0.15s var(--ease-spring), background 0.15s ease;
}
.action-btn:active { transform: scale(0.95); }
.action-btn.danger {
  background: linear-gradient(135deg, #ffa0a0, var(--color-danger));
  color: #fff;
  box-shadow: 0 4px 12px rgba(226, 109, 109, 0.35);
}
.action-btn.ghost {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
}
:root[data-theme="dark"] .action-btn.ghost {
  background: var(--color-pink-100);
}

.actions-enter-active, .actions-leave-active { transition: opacity 0.15s ease; }
.actions-enter-from, .actions-leave-to { opacity: 0; }
</style>
