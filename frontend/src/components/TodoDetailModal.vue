<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { Todo } from '../types'
import { useTodoStore } from '../stores/todo'
import DateChipBar from './DateChipBar.vue'

const props = defineProps<{ todo: Todo | null }>()
const emit = defineEmits<{
  close: []
  toggle: [Todo]
  remove: [Todo]
}>()

const store = useTodoStore()
const dialogRef = ref<HTMLDialogElement | null>(null)

// 视图 / 编辑模式
const mode = ref<'view' | 'edit'>('view')
const editText = ref('')
const editDue = ref<string | null>(null)
// 进入 edit 时的原始快照(避免 store 更新触发 watch 重置编辑态)
const originalContent = ref('')
const originalDue = ref<string | null>(null)
const saving = ref(false)
const saveError = ref<string | null>(null)

// 跟随父组件的 todo prop 开关 modal;打开时重置到 view 并同步编辑态
// 注意:编辑过程中(props.todo 因 store 更新而变)不要覆盖 editText/editDue,
// 避免和正在进行的 save 逻辑抢状态
watch(
  () => props.todo,
  (t, prev) => {
    if (t) {
      // 只在「首次打开」(从 null → 有 todo)时同步编辑态
      const justOpened = !prev
      if (justOpened) {
        mode.value = 'view'
        editText.value = t.content
        editDue.value = t.due_date
        originalContent.value = t.content
        originalDue.value = t.due_date
        saveError.value = null
      }
      dialogRef.value?.showModal()
    } else {
      dialogRef.value?.close()
    }
  }
)

function close() {
  dialogRef.value?.close()
  emit('close')
}

// 点击 backdrop 关闭(target === dialogRef 本身)
function onBackdropClick(e: MouseEvent) {
  if (e.target === dialogRef.value) close()
}

// 元信息格式化
function formatDateTime(iso: string | null) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    month: 'numeric', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function formatDueDate(d: string | null) {
  if (!d) return ''
  const [y, m, day] = d.split('-')
  return `${parseInt(m!)} 月 ${parseInt(day!)} 日`
}

const createdLabel = computed(() => formatDateTime(props.todo?.created_at ?? null))
const completedLabel = computed(() => formatDateTime(props.todo?.completed_at ?? null))
const dueLabel = computed(() => formatDueDate(props.todo?.due_date ?? null))
const hasEditChanges = computed(() => {
  if (!props.todo) return false
  if (editText.value.trim() === '') return false
  // 对比快照,避免 store 更新影响"是否有改动"的判断
  return editText.value !== originalContent.value || editDue.value !== originalDue.value
})

function onToggle() {
  if (props.todo) emit('toggle', props.todo)
}
function onEdit() {
  if (!props.todo) return
  mode.value = 'edit'
  editText.value = props.todo.content
  editDue.value = props.todo.due_date
  // 拍快照,避免后续 store 更新触发 watch 重置编辑态
  originalContent.value = props.todo.content
  originalDue.value = props.todo.due_date
  saveError.value = null
}
function onCancelEdit() {
  if (!props.todo) return
  mode.value = 'view'
  editText.value = props.todo.content
  editDue.value = props.todo.due_date
  saveError.value = null
}
async function onSaveEdit() {
  if (!props.todo) return
  const trimmed = editText.value.trim()
  if (!trimmed) {
    saveError.value = '内容不能为空'
    return
  }
  saving.value = true
  saveError.value = null
  try {
    // 对比快照而非 props.todo(后者会被 store 更新改变)
    if (trimmed !== originalContent.value) {
      await store.updateContent(props.todo.id, trimmed)
    }
    if (editDue.value !== originalDue.value) {
      await store.updateDueDate(props.todo.id, editDue.value)
    }
    mode.value = 'view'
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : '保存失败'
  } finally {
    saving.value = false
  }
}
function onRemove() {
  if (props.todo) emit('remove', props.todo)
}
</script>

<template>
  <dialog ref="dialogRef" @close="emit('close')" @click="onBackdropClick">
    <div v-if="todo" class="dlg-body" :data-mode="mode">
      <header class="dlg-head">
        <div class="author">
          <span class="sig" :style="{ background: todo.author_color }">{{ todo.author_emoji }}</span>
          <span class="created">{{ createdLabel }}</span>
          <span v-if="todo.pinned" class="pin" title="已置顶">📌</span>
        </div>
        <button type="button" class="close-btn" aria-label="关闭" @click="close">✕</button>
      </header>

      <!-- 视图模式 -->
      <div v-if="mode === 'view'" class="content">{{ todo.content }}</div>

      <!-- 编辑模式 -->
      <div v-else class="content edit-form">
        <textarea
          v-model="editText"
          class="edit-textarea"
          rows="4"
          maxlength="500"
          placeholder="想做的小事…"
          :disabled="saving"
        ></textarea>
        <div class="edit-due">
          <span class="edit-due-label">📅 截止日期</span>
          <DateChipBar v-model="editDue" />
        </div>
        <p v-if="saveError" class="edit-error">{{ saveError }}</p>
      </div>

      <dl v-if="mode === 'view' && (dueLabel || completedLabel)" class="meta">
        <div v-if="dueLabel" class="meta-row">
          <dt>📅 截止</dt>
          <dd>{{ dueLabel }}</dd>
        </div>
        <div v-if="completedLabel" class="meta-row">
          <dt>✅ 完成于</dt>
          <dd>{{ completedLabel }}</dd>
        </div>
      </dl>

      <footer v-if="mode === 'view'" class="dlg-footer">
        <button
          type="button"
          class="btn"
          :class="todo.completed_at ? 'btn-ghost' : 'btn-primary'"
          @click="onToggle"
        >
          {{ todo.completed_at ? '↻ 取消完成' : '✓ 标记完成' }}
        </button>
        <button type="button" class="btn btn-ghost" @click="onEdit">✏️ 编辑</button>
        <button type="button" class="btn btn-danger-ghost" @click="onRemove">🗑️ 删除</button>
      </footer>
      <footer v-else class="dlg-footer">
        <button
          type="button"
          class="btn btn-ghost"
          :disabled="saving"
          @click="onCancelEdit"
        >取消</button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="saving || !hasEditChanges"
          @click="onSaveEdit"
        >{{ saving ? '保存中…' : '保存' }}</button>
      </footer>
    </div>
  </dialog>
</template>

<style scoped>
.dlg-body {
  padding: 18px 20px 16px;
  min-width: min(360px, 90vw);
  max-width: 480px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.dlg-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.author {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-muted);
  min-width: 0;
}
.sig {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.06), inset 0 -1px 2px rgba(0,0,0,0.08);
  flex: 0 0 auto;
}
.created { white-space: nowrap; }
.pin { font-size: 13px; }
.close-btn {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--color-pink-50);
  color: var(--color-muted);
  font-size: 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  min-height: 0;
  padding: 0;
  transition: background 0.15s ease;
}
.close-btn:hover {
  background: var(--color-pink-100);
  color: var(--color-pink-500);
}
:root[data-theme="dark"] .close-btn {
  background: var(--color-pink-100);
}

.content {
  font-size: 16px;
  line-height: 1.7;
  color: var(--color-ink);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 50vh;
  overflow-y: auto;
  padding: 4px 2px;
}
[data-mode="edit"] .content {
  max-height: none;
  overflow: visible;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.edit-textarea {
  width: 100%;
  border: 1.5px solid var(--color-line-2);
  border-radius: var(--radius-md);
  padding: 10px 12px;
  font: inherit;
  font-size: 15px;
  line-height: 1.6;
  color: var(--color-ink);
  background: var(--color-cream);
  resize: vertical;
  min-height: 90px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.edit-textarea:focus {
  outline: none;
  border-color: var(--color-pink-300);
  box-shadow: 0 0 0 3px rgba(248, 165, 194, 0.18);
}
:root[data-theme="dark"] .edit-textarea {
  background: var(--color-pink-100);
  border-color: var(--color-line-2);
}
.edit-due {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.edit-due-label {
  font-size: 13px;
  color: var(--color-muted);
}
.edit-error {
  margin: 0;
  font-size: 12.5px;
  color: var(--color-danger);
}

.meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 0;
  padding: 12px 14px;
  background: var(--color-pink-50);
  border-radius: var(--radius-md);
  font-size: 12.5px;
}
:root[data-theme="dark"] .meta {
  background: var(--color-pink-100);
}
.meta-row {
  display: flex;
  gap: 8px;
  align-items: baseline;
}
.meta-row dt {
  color: var(--color-muted);
  flex: 0 0 auto;
  margin: 0;
}
.meta-row dd {
  color: var(--color-ink);
  margin: 0;
  font-weight: 500;
}

.dlg-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  flex-wrap: wrap;
}
.btn-danger-ghost {
  background: transparent;
  color: var(--color-danger);
  padding: 10px 18px;
  border-radius: var(--radius-pill);
  font-weight: 600;
  border: 1px solid transparent;
  transition: background 0.15s ease;
}
.btn-danger-ghost:hover {
  background: rgba(226, 109, 109, 0.08);
  border-color: rgba(226, 109, 109, 0.2);
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
