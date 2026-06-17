<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Todo } from '../types'
import { useTodoStore } from '../stores/todo'
import { useUndo } from '../composables/useUndo'

const props = defineProps<{ todo: Todo }>()
const emit = defineEmits<{ open: [Todo] }>()
const store = useTodoStore()
const undo = useUndo()
const showActions = ref(false)
let pressTimer: number | null = null

async function toggle(e: Event) {
  e.stopPropagation()
  await store.toggle(props.todo.id)
  if (!props.todo.completed_at) {
    undo.push(`已勾上「${props.todo.content.slice(0, 12)}」`, async () => {
      await store.toggle(props.todo.id)
    })
  }
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

async function togglePin() {
  showActions.value = false
  const wasPinned = props.todo.pinned
  await store.togglePin(props.todo.id)
  undo.push(
    wasPinned ? `已取消置顶「${props.todo.content.slice(0, 12)}」` : `已置顶「${props.todo.content.slice(0, 12)}」`,
    async () => { await store.togglePin(props.todo.id) }
  )
}

function onPressStart(e: Event) {
  // 长按只响应触屏/鼠标主键,不响应 .check 等内部按钮
  if ((e.target as HTMLElement).closest('button')) return
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

function openDetail() {
  emit('open', props.todo)
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

// 已完成 tab 显示完成时间戳,待办 tab 显示创建时间
const displayTime = computed(() => {
  const ts = props.todo.completed_at ?? props.todo.created_at
  return relativeTime(ts)
})
</script>

<template>
  <article
    :class="['todo-card', { done: todo.completed_at, showActions, pinned: todo.pinned }]"
    role="button"
    tabindex="0"
    :aria-label="`待办: ${todo.content}`"
    @click="openDetail"
    @keydown.enter="openDetail"
    @keydown.space.prevent="openDetail"
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
      <p class="content">{{ todo.content }}</p>

      <div class="meta">
        <span
          class="sig"
          :style="{ background: todo.author_color }"
          :title="`由 ${todo.author_emoji} 添加`"
        >{{ todo.author_emoji }}</span>
        <span v-if="todo.pinned" class="pin" title="已置顶">📌</span>
        <span v-if="todo.due_date" class="due">📅 {{ formatDue(todo.due_date) }}</span>
        <span class="time">
          {{ todo.completed_at ? '✓ ' : '' }}{{ displayTime }}
        </span>
      </div>
    </div>

    <Transition name="actions">
      <div v-if="showActions" class="action-overlay" @click.stop="showActions = false">
        <button class="action-btn primary" @click.stop="togglePin">
          {{ todo.pinned ? '↩️ 取消置顶' : '📌 置顶' }}
        </button>
        <button class="action-btn danger" @click.stop="remove">🗑️ 删除</button>
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
  cursor: pointer;
}
.todo-card:hover { box-shadow: var(--shadow-md); }
.todo-card:active { transform: scale(0.99); }
.todo-card:focus-visible {
  outline: none;
  box-shadow: var(--shadow-md), 0 0 0 3px rgba(248, 165, 194, 0.45);
}
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
.todo-card.pinned {
  background: linear-gradient(135deg, #fffaf0 0%, #fff4d4 100%);
}
:root[data-theme="dark"] .todo-card.pinned {
  background: linear-gradient(135deg, #4a4028 0%, #3a3020 100%);
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
.content {
  font-size: 15px;
  line-height: 1.55;
  word-break: break-word;
  cursor: pointer;
  margin: 0;
  color: var(--color-ink);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

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
.pin {
  font-size: 12px;
  line-height: 1;
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
.action-btn.primary {
  background: linear-gradient(135deg, #fff4d4, #ffe69e);
  color: #8a6a1a;
  box-shadow: 0 4px 12px rgba(247, 213, 96, 0.35);
}
:root[data-theme="dark"] .action-btn.primary {
  background: linear-gradient(135deg, #5a4a28, #6a5430);
  color: #f7d560;
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