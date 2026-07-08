<script setup lang="ts">
import { ref, watch, computed, onMounted, onBeforeUnmount } from 'vue'
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
const open = ref(false)

function onEsc(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) close()
}
onMounted(() => document.addEventListener('keydown', onEsc))
onBeforeUnmount(() => document.removeEventListener('keydown', onEsc))

// 视图 / 编辑模式
const mode = ref<'view' | 'edit'>('view')
const editText = ref('')
const editDue = ref<string | null>(null)
// 进入 edit 时的原始快照(避免 store 更新触发 watch 重置编辑态)
const originalContent = ref('')
const originalDue = ref<string | null>(null)
const saving = ref(false)
const saveError = ref<string | null>(null)

// ── 评论 ──
const commentText = ref('')
const commentEmoji = ref<string>(localStorage.getItem('todo:emoji') ?? '🌸')
const commentColor = ref<string>(localStorage.getItem('todo:color') ?? '#ec7da6')
const postingComment = ref(false)
const commentError = ref<string | null>(null)

// 当前 todo 的评论(从 store 缓存中派生)
const comments = computed(() => (props.todo ? store.commentsByTodo[props.todo.id] ?? [] : []))
const commentsLoading = computed(() => (props.todo ? !!store.commentsLoading[props.todo.id] : false))

function resetCommentInput() {
  commentText.value = ''
  commentError.value = null
}

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
        resetCommentInput()
        // 拉取评论(后端不返回评论内嵌,按需拉)
        store.fetchComments(t.id).catch(() => { /* 静默 */ })
      }
      open.value = true
    } else {
      open.value = false
    }
  }
)

function close() {
  open.value = false
  emit('close')
}

// 点击 backdrop 关闭
function onBackdropClick(e: MouseEvent) {
  if ((e.target as HTMLElement).dataset.backdrop) close()
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

async function submitComment() {
  if (!props.todo) return
  const trimmed = commentText.value.trim()
  if (!trimmed) {
    commentError.value = '写点内容再发送吧'
    return
  }
  if (trimmed.length > 500) {
    commentError.value = '评论最多 500 字'
    return
  }
  postingComment.value = true
  commentError.value = null
  try {
    await store.addComment(props.todo.id, {
      content: trimmed,
      author_emoji: commentEmoji.value,
      author_color: commentColor.value,
    })
    commentText.value = ''
  } catch (e) {
    commentError.value = e instanceof Error ? e.message : '发送失败'
  } finally {
    postingComment.value = false
  }
}

function commentRelative(iso: string) {
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
  <Teleport to="body">
  <div v-if="open" class="modal-overlay" data-backdrop @click="onBackdropClick">
    <div class="modal-card" @click.stop>
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

      <!-- ── 评论(只 view 模式显示) ── -->
      <section v-if="mode === 'view'" class="comments">
        <h3 class="comments-title">💬 评论 <span class="comments-count">{{ comments.length }}</span></h3>

        <div v-if="commentsLoading && comments.length === 0" class="comment-skeleton">加载中…</div>

        <ul v-else-if="comments.length > 0" class="comment-list">
          <li v-for="c in comments" :key="c.id" class="comment-item">
            <span class="comment-sig" :style="{ background: c.author_color }">{{ c.author_emoji }}</span>
            <div class="comment-bubble">
              <p class="comment-content">{{ c.content }}</p>
              <span class="comment-time">{{ commentRelative(c.created_at) }}</span>
            </div>
          </li>
        </ul>

        <p v-else class="comment-empty">还没有评论,说一句吧 ✨</p>

        <form class="comment-form" @submit.prevent="submitComment">
          <span
            class="comment-sig-input"
            :style="{ background: commentColor }"
            :title="`当前签名:${commentEmoji}`"
          >{{ commentEmoji }}</span>
          <input
            v-model="commentText"
            class="comment-input"
            type="text"
            maxlength="500"
            placeholder="补一句…"
            :disabled="postingComment"
          />
          <button
            type="submit"
            class="comment-send"
            :disabled="postingComment || !commentText.trim()"
            aria-label="发送评论"
          >{{ postingComment ? '…' : '发送' }}</button>
        </form>
        <p v-if="commentError" class="comment-error">{{ commentError }}</p>
      </section>

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
    </div>
  </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(58, 46, 54, 0.45);
  backdrop-filter: blur(2px);
  animation: fadeIn 0.2s ease;
}
.modal-card {
  border: none;
  border-radius: var(--radius-lg);
  padding: 0;
  background: var(--color-warm-bg);
  color: var(--color-ink);
  max-width: min(420px, calc(100vw - 32px));
  width: 100%;
  box-shadow: var(--shadow-lg);
  animation: pop 0.28s var(--ease-spring);
  max-height: 90vh;
  overflow-y: auto;
}
.dlg-body {
  padding: 18px 20px 16px;
  min-width: min(360px, 90vw);
  max-width: 480px;
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

/* ── 评论 ── */
.comments {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.comments-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-ink);
  display: flex;
  align-items: center;
  gap: 6px;
}
.comments-count {
  font-size: 11px;
  color: var(--color-muted);
  background: var(--color-pink-50);
  padding: 1px 8px;
  border-radius: var(--radius-pill);
  font-weight: 500;
}
:root[data-theme="dark"] .comments-count {
  background: var(--color-pink-100);
}
.comment-skeleton,
.comment-empty {
  margin: 0;
  padding: 14px 12px;
  text-align: center;
  font-size: 12.5px;
  color: var(--color-muted);
  background: var(--color-pink-50);
  border-radius: var(--radius-md);
}
:root[data-theme="dark"] .comment-skeleton,
:root[data-theme="dark"] .comment-empty {
  background: var(--color-pink-100);
}
.comment-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 30vh;
  overflow-y: auto;
}
.comment-item {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}
.comment-sig,
.comment-sig-input {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  flex: 0 0 auto;
  box-shadow: 0 2px 4px rgba(0,0,0,0.06), inset 0 -1px 2px rgba(0,0,0,0.08);
}
.comment-sig-input { width: 28px; height: 28px; font-size: 14px; }
.comment-bubble {
  flex: 1 1 auto;
  min-width: 0;
  background: var(--color-pink-50);
  padding: 8px 12px;
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  gap: 2px;
}
:root[data-theme="dark"] .comment-bubble {
  background: var(--color-pink-100);
}
.comment-content {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.55;
  color: var(--color-ink);
  word-break: break-word;
  white-space: pre-wrap;
}
.comment-time {
  font-size: 11px;
  color: var(--color-muted);
}
.comment-form {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 4px;
}
.comment-input {
  flex: 1 1 auto;
  min-width: 0;
  background: var(--color-cream);
  border: 1.5px solid var(--color-line-2);
  border-radius: var(--radius-pill);
  padding: 7px 14px;
  font: inherit;
  font-size: 13.5px;
  min-height: 32px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.comment-input:focus {
  outline: none;
  border-color: var(--color-pink-300);
  box-shadow: 0 0 0 3px rgba(248, 165, 194, 0.18);
}
:root[data-theme="dark"] .comment-input {
  background: var(--color-pink-100);
  border-color: var(--color-line-2);
}
.comment-send {
  flex: 0 0 auto;
  padding: 6px 16px;
  border-radius: var(--radius-pill);
  background: linear-gradient(135deg, var(--color-pink-400), var(--color-pink-500));
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  min-height: 32px;
  min-width: 0;
  transition: transform 0.15s var(--ease-spring), opacity 0.15s ease;
}
.comment-send:active { transform: scale(0.94); }
.comment-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: var(--color-pink-100);
  color: var(--color-muted);
}
.comment-error {
  margin: 0;
  font-size: 12px;
  color: var(--color-danger);
}
</style>
