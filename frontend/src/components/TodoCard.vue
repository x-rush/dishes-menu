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

async function toggle() {
  await store.toggle(props.todo.id)
  if (!props.todo.completed_at) {
    // 刚勾上 → 5 秒内可撤销
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

function formatDue(d: string | null) {
  if (!d) return ''
  // d = 'YYYY-MM-DD',取 MM-DD
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
  <article :class="['todo-card', { done: todo.completed_at }]">
    <button
      type="button"
      class="check"
      :aria-pressed="!!todo.completed_at"
      :aria-label="todo.completed_at ? '取消完成' : '标记完成'"
      @click="toggle"
    >
      <span v-if="todo.completed_at">✓</span>
    </button>

    <div class="body">
      <div v-if="editing" class="edit-row">
        <input v-model="editText" maxlength="500" @keydown.enter="saveEdit" @keydown.esc="editing = false" />
        <button class="btn btn-primary save" @click="saveEdit">保存</button>
      </div>
      <div v-else class="content-row">
        <p class="content" @click="startEdit">{{ todo.content }}</p>
      </div>

      <div class="meta">
        <span class="sig" :style="{ background: todo.author_color }" :title="`由 ${todo.author_emoji} 添加`">
          {{ todo.author_emoji }}
        </span>
        <span v-if="todo.due_date" class="due">📅 {{ formatDue(todo.due_date) }}</span>
        <span class="time">{{ relativeTime(todo.created_at) }}</span>
      </div>
    </div>

    <button class="btn btn-icon del" :aria-label="`删除 ${todo.content}`" @click="remove">×</button>
  </article>
</template>

<style scoped>
.todo-card {
  display: flex;
  gap: 10px;
  padding: 12px 14px;
  background: var(--color-cream);
  border-radius: var(--radius-md);
  align-items: flex-start;
  transition: opacity 0.3s ease, background 0.3s ease;
}
.todo-card.done { opacity: 0.55; background: var(--color-warm-bg); }
.todo-card.done .content { text-decoration: line-through; color: var(--color-muted); }

.check {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 2px solid var(--color-pink-300);
  background: transparent;
  color: var(--color-pink-500);
  font-size: 16px;
  flex: 0 0 auto;
  margin-top: 2px;
  transition: background 0.2s var(--ease-spring), border-color 0.2s ease, transform 0.15s var(--ease-spring);
}
.check:hover { border-color: var(--color-pink-500); }
.check:active { transform: scale(0.88); }
.todo-card.done .check {
  background: var(--color-pink-400);
  border-color: var(--color-pink-500);
  color: #fff;
}

.body { flex: 1 1 auto; min-width: 0; }
.content-row { display: flex; gap: 8px; align-items: center; }
.content {
  font-size: 15px;
  word-break: break-word;
  cursor: text;
  margin: 0;
}
.edit-row {
  display: flex;
  gap: 6px;
}
.edit-row input { flex: 1 1 auto; font-size: 14px; padding: 6px 10px; }
.save { padding: 4px 12px; font-size: 12px; }

.meta {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  font-size: 12px;
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
  filter: brightness(1.05);
}
.due { color: var(--color-peach-300); }
.time { color: var(--color-muted); }

.del {
  width: 28px;
  height: 28px;
  min-width: 28px;
  min-height: 28px;
  font-size: 16px;
  background: transparent;
  color: var(--color-muted);
  flex: 0 0 auto;
}
.del:hover { background: var(--color-pink-100); color: var(--color-danger); }
</style>
