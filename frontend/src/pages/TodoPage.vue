<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useTodoStore } from '../stores/todo'
import { useUndo } from '../composables/useUndo'
import TodoInputBar from '../components/TodoInputBar.vue'
import TodoCard from '../components/TodoCard.vue'
import TodoDetailModal from '../components/TodoDetailModal.vue'
import Notebook from '../components/illustrations/Notebook.vue'
import type { Todo } from '../types'

const store = useTodoStore()
const undo = useUndo()

onMounted(() => {
  store.fetchAll()
})

// —— Tab + 搜索 ——
const activeTab = ref<'open' | 'done'>('open')
const searchQuery = ref('')

// —— 筛选 ——
const filterEmoji = ref<string | null>(null)  // null = 全部提交人
const filterRange = ref<'all' | 'today' | 'week' | 'month'>('all')

// 候选 emoji:从所有 todo 提取去重
const availableEmojis = computed(() => {
  const set = new Set<string>()
  for (const t of store.todos) {
    if (t.author_emoji) set.add(t.author_emoji)
  }
  return Array.from(set)
})

// Popover 状态
const emojiPopoverOpen = ref(false)
const rangePopoverOpen = ref(false)
const emojiRef = ref<HTMLElement | null>(null)
const rangeRef = ref<HTMLElement | null>(null)

function onDocClick(e: MouseEvent) {
  const target = e.target as Node
  if (emojiPopoverOpen.value && emojiRef.value && !emojiRef.value.contains(target)) {
    emojiPopoverOpen.value = false
  }
  if (rangePopoverOpen.value && rangeRef.value && !rangeRef.value.contains(target)) {
    rangePopoverOpen.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))

const rangeOptions: { value: typeof filterRange.value; label: string; hint: string }[] = [
  { value: 'all', label: '全部时间', hint: '不限' },
  { value: 'today', label: '今天到期', hint: '剩 0 天' },
  { value: 'week', label: '本周到期', hint: '7 天内' },
  { value: 'month', label: '本月到期', hint: '到月底' },
]

const rangeLabel: Record<typeof filterRange.value, string> = {
  all: '全部时间',
  today: '今天',
  week: '本周',
  month: '本月',
}

// —— 详情 modal ——
const detailTodoId = ref<number | null>(null)
const detailTodo = computed(() =>
  detailTodoId.value == null ? null : store.todos.find(t => t.id === detailTodoId.value) ?? null
)

// —— 加载更多(P3) ——
// 一次只渲染前 PAGE_SIZE 条;切换 tab / 改筛选 / 改搜索时重置
const PAGE_SIZE = 50
const visibleCount = ref(PAGE_SIZE)
const visibleList = computed(() => filteredList.value.slice(0, visibleCount.value))
const hasMore = computed(() => filteredList.value.length > visibleCount.value)
function loadMore() {
  visibleCount.value = Math.min(visibleCount.value + PAGE_SIZE, filteredList.value.length)
}
// 切 tab / 改筛选 / 改搜索时把 visibleCount 重置回 PAGE_SIZE
watch([activeTab, filterEmoji, filterRange, searchQuery], () => {
  visibleCount.value = PAGE_SIZE
})
function openDetail(t: Todo) {
  detailTodoId.value = t.id
}
function closeDetail() {
  detailTodoId.value = null
}

async function onModalToggle(t: Todo) {
  // 仅在「标记完成」时记 undo;取消完成是误操作修复,不需要
  const wasOpen = !t.completed_at
  await store.toggle(t.id)
  if (wasOpen) {
    undo.push(`已勾上「${t.content.slice(0, 12)}」`, async () => {
      await store.toggle(t.id)
    })
  }
}

async function onModalRemove(t: Todo) {
  const snapshot = { ...t }
  await store.remove(t.id)
  closeDetail()
  undo.push(`已删除「${snapshot.content.slice(0, 12)}」`, async () => {
    await store.create({
      content: snapshot.content,
      due_date: snapshot.due_date,
      author_emoji: snapshot.author_emoji,
      author_color: snapshot.author_color,
    })
  })
}

// —— 派生列表 ——
// 已完成按完成时间倒序;待办按 pinned → due_date → created_at 排
const sortedOpen = computed(() =>
  [...store.open].sort((a, b) => {
    // 1. 置顶优先
    if (a.pinned !== b.pinned) return a.pinned ? -1 : 1
    // 2. due_date 优先(无 due 的排后)
    if (a.due_date && !b.due_date) return -1
    if (!a.due_date && b.due_date) return 1
    if (a.due_date && b.due_date) return a.due_date.localeCompare(b.due_date)
    // 3. created_at 倒序
    return b.created_at.localeCompare(a.created_at)
  })
)
const sortedDone = computed(() =>
  [...store.done].sort((a, b) =>
    (b.completed_at ?? '').localeCompare(a.completed_at ?? '')
  )
)

const sourceList = computed(() =>
  activeTab.value === 'open' ? sortedOpen.value : sortedDone.value
)

// 链式过滤:tab → emoji → range → search
const filteredList = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  const emoji = filterEmoji.value
  const range = filterRange.value
  if (!q && !emoji && range === 'all') return sourceList.value

  // 准备时间边界(range 共享)
  let rangeStart: Date | null = null
  let rangeEnd: Date | null = null
  if (range !== 'all') {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    rangeStart = today
    if (range === 'today') {
      rangeEnd = new Date(today)
      rangeEnd.setDate(rangeEnd.getDate() + 1)
    } else if (range === 'week') {
      rangeEnd = new Date(today)
      rangeEnd.setDate(rangeEnd.getDate() + 7)
    } else if (range === 'month') {
      rangeEnd = new Date(today.getFullYear(), today.getMonth() + 1, 1)
    }
  }

  return sourceList.value.filter(t => {
    if (emoji && t.author_emoji !== emoji) return false
    if (rangeStart && rangeEnd) {
      // 没 due_date 的不在 today/week/month 范围内
      if (!t.due_date) return false
      const due = new Date(t.due_date + 'T00:00:00')
      if (due < rangeStart || due >= rangeEnd) return false
    }
    if (q && !t.content.toLowerCase().includes(q)) return false
    return true
  })
})

const totalCount = computed(() => store.todos.length)
const openCount = computed(() => sortedOpen.value.length)
const doneCount = computed(() => sortedDone.value.length)
const isFiltering = computed(() =>
  searchQuery.value.trim().length > 0 || filterEmoji.value !== null || filterRange.value !== 'all'
)
const hasActiveFilter = computed(() => filterEmoji.value !== null || filterRange.value !== 'all')
// 全部完成:仅在「待办」tab 且 openCount === 0 但已有过 done 时显示庆祝态
const allDone = computed(() => activeTab.value === 'open' && openCount.value === 0 && doneCount.value > 0)
</script>

<template>
  <div class="todo-page">
    <header class="page-head">
      <div class="hero">
        <h1>💝 想做的小事</h1>
        <p class="muted">写下你们想一起做的事,做完打勾 ✨</p>
      </div>

      <div v-if="totalCount > 0" class="stats">
        <div class="stat">
          <span class="stat-num">{{ openCount }}</span>
          <span class="stat-label">待办</span>
        </div>
        <span class="stat-divider"></span>
        <div class="stat">
          <span class="stat-num done">{{ doneCount }}</span>
          <span class="stat-label">已完成</span>
        </div>
        <span class="stat-divider"></span>
        <div class="stat">
          <span class="stat-num total">{{ totalCount }}</span>
          <span class="stat-label">合计</span>
        </div>
      </div>

      <!-- 搜索框 -->
      <div class="search-bar" :class="{ active: isFiltering }">
        <span class="search-icon" aria-hidden="true">🔍</span>
        <input
          v-model="searchQuery"
          type="search"
          inputmode="search"
          placeholder="找一条待办…"
          aria-label="搜索待办"
          @keydown.escape="searchQuery = ''"
        />
        <button
          v-if="isFiltering"
          type="button"
          class="search-clear"
          aria-label="清空搜索"
          @click="searchQuery = ''"
        >✕</button>
      </div>

      <!-- Tab 切换 + 筛选触发 -->
      <div class="tab-row">
        <div class="tab-bar" role="tablist">
          <button
            type="button"
            role="tab"
            :class="{ active: activeTab === 'open' }"
            :aria-selected="activeTab === 'open'"
            @click="activeTab = 'open'"
          >
            待办 <span class="badge">{{ openCount }}</span>
          </button>
          <button
            type="button"
            role="tab"
            :class="{ active: activeTab === 'done' }"
            :aria-selected="activeTab === 'done'"
            @click="activeTab = 'done'"
          >
            已完成 <span class="badge">{{ doneCount }}</span>
          </button>
        </div>

        <!-- 筛选触发器(emoji + range) -->
        <div class="filter-triggers">
          <div ref="emojiRef" class="filter-trigger-wrap">
            <button
              type="button"
              class="filter-trigger"
              :class="{ active: filterEmoji !== null }"
              :aria-expanded="emojiPopoverOpen"
              :disabled="availableEmojis.length === 0"
              @click.stop="emojiPopoverOpen = !emojiPopoverOpen"
            >
              <span>{{ filterEmoji ?? '🐾' }}</span>
              <span class="caret" :class="{ up: emojiPopoverOpen }">▾</span>
            </button>
            <Transition name="popover">
              <div v-if="emojiPopoverOpen" class="filter-popover" role="dialog" aria-label="按提交人筛选">
                <button
                  type="button"
                  class="filter-option"
                  :class="{ active: filterEmoji === null }"
                  @click="filterEmoji = null; emojiPopoverOpen = false"
                >
                  <span>🐾</span>
                  <span class="option-label">全部</span>
                </button>
                <button
                  v-for="e in availableEmojis"
                  :key="e"
                  type="button"
                  class="filter-option"
                  :class="{ active: filterEmoji === e }"
                  @click="filterEmoji = e; emojiPopoverOpen = false"
                >
                  <span>{{ e }}</span>
                  <span class="option-label">{{ e }}</span>
                </button>
              </div>
            </Transition>
          </div>

          <div ref="rangeRef" class="filter-trigger-wrap">
            <button
              type="button"
              class="filter-trigger"
              :class="{ active: filterRange !== 'all' }"
              :aria-expanded="rangePopoverOpen"
              @click.stop="rangePopoverOpen = !rangePopoverOpen"
            >
              <span aria-hidden="true">📅</span>
              <span>{{ rangeLabel[filterRange] }}</span>
              <span class="caret" :class="{ up: rangePopoverOpen }">▾</span>
            </button>
            <Transition name="popover">
              <div v-if="rangePopoverOpen" class="filter-popover" role="dialog" aria-label="按时间筛选">
                <button
                  v-for="opt in rangeOptions"
                  :key="opt.value"
                  type="button"
                  class="filter-option wide"
                  :class="{ active: filterRange === opt.value }"
                  @click="filterRange = opt.value; rangePopoverOpen = false"
                >
                  <span class="option-label">{{ opt.label }}</span>
                  <span class="option-hint">{{ opt.hint }}</span>
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </div>

      <!-- 激活筛选条件行(可单独移除) -->
      <div v-if="hasActiveFilter" class="filter-row">
        <button
          v-if="filterEmoji"
          type="button"
          class="filter-chip"
          @click="filterEmoji = null"
        >
          <span>{{ filterEmoji }}</span>
          <span class="chip-x">✕</span>
        </button>
        <button
          v-if="filterRange !== 'all'"
          type="button"
          class="filter-chip"
          @click="filterRange = 'all'"
        >
          <span>📅 {{ rangeLabel[filterRange] }}</span>
          <span class="chip-x">✕</span>
        </button>
        <button
          type="button"
          class="filter-clear-all"
          @click="filterEmoji = null; filterRange = 'all'"
        >清除全部</button>
      </div>
    </header>

    <TodoInputBar />

    <section v-if="store.loading && !store.todos.length" class="loading">
      <span class="spinner"></span>
      <span>加载中…</span>
    </section>

    <Transition name="fade-up" mode="out-in">
      <section :key="activeTab" class="list-section">
        <!-- 列表非空 -->
        <div v-if="visibleList.length" class="todo-list">
          <TodoCard
            v-for="t in visibleList"
            :key="t.id"
            :todo="t"
            @open="openDetail"
          />

          <!-- 加载更多(P3) -->
          <button
            v-if="hasMore"
            type="button"
            class="load-more"
            @click="loadMore"
          >
            再加载 {{ Math.min(PAGE_SIZE, filteredList.length - visibleCount) }} 条 · 剩 {{ filteredList.length - visibleCount }} 条
          </button>
        </div>

        <!-- 搜索无匹配 -->
        <div v-else-if="isFiltering" class="empty">
          <Notebook :size="120" />
          <p class="empty-title">没有匹配「{{ searchQuery }}」的待办</p>
          <p class="empty-sub">试试别的关键词,或者点 ✕ 清空 ✨</p>
        </div>

        <!-- 待办 tab + 全部完成🎉 -->
        <div v-else-if="allDone" class="empty all-done">
          <div class="confetti-burst" aria-hidden="true">🎉</div>
          <p class="empty-title">太棒了,所有事都搞定啦</p>
          <p class="empty-sub">今晚可以安心窝在一起了 ✨</p>
        </div>

        <!-- 已完成 tab + 没有已完成 -->
        <div v-else-if="activeTab === 'done'" class="empty">
          <p class="empty-title">还没有完成的事</p>
          <p class="empty-sub">先去待办里打几个勾吧 ✅</p>
        </div>

        <!-- 待办 tab + 没有任何 todo -->
        <div v-else class="empty">
          <Notebook :size="160" />
          <p class="empty-title">这本本子还空着</p>
          <p class="empty-sub">在下面写一句想做的小事,<br />我们一起慢慢打勾 ✏️</p>
        </div>
      </section>
    </Transition>

    <div v-if="store.error" class="error">{{ store.error }}</div>

    <TodoDetailModal
      :todo="detailTodo"
      @close="closeDetail"
      @toggle="onModalToggle"
      @remove="onModalRemove"
    />
  </div>
</template>

<style scoped>
.todo-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 20px 18px 32px;
  max-width: 560px;
  margin: 0 auto;
}

.page-head {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 4px 0;
}
.hero h1 {
  font-size: 28px;
  font-family: var(--font-display);
  color: var(--color-pink-500);
  margin: 0 0 6px;
  letter-spacing: 0.01em;
  line-height: 1.2;
}
.muted {
  font-size: 13.5px;
  color: var(--color-muted);
  line-height: 1.5;
  margin: 0;
}

.stats {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  padding: 10px 16px;
  background: linear-gradient(135deg, #fff0f5 0%, #ffe4d4 100%);
  border-radius: var(--radius-pill);
  align-self: flex-start;
  box-shadow: var(--shadow-sm);
}
:root[data-theme="dark"] .stats {
  background: linear-gradient(135deg, #3a2830 0%, #4a3a30 100%);
}
.stat {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  font-size: 12px;
  color: var(--color-muted);
}
.stat-num {
  font-size: 18px;
  font-weight: 700;
  font-family: var(--font-display);
  color: var(--color-pink-500);
  line-height: 1;
}
.stat-num.done { color: var(--color-mint-300); }
.stat-num.total { color: var(--color-peach-300); }
.stat-label { letter-spacing: 0.04em; }
.stat-divider {
  width: 1px;
  height: 14px;
  background: rgba(0, 0, 0, 0.08);
}
:root[data-theme="dark"] .stat-divider { background: rgba(255, 255, 255, 0.08); }

/* —— 搜索框 —— */
.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--color-pink-50);
  border: 1.5px solid transparent;
  border-radius: var(--radius-pill);
  transition: background 0.15s ease, border-color 0.15s ease;
}
.search-bar.active {
  background: #fff;
  border-color: var(--color-pink-300);
  box-shadow: 0 0 0 3px rgba(248, 165, 194, 0.18);
}
:root[data-theme="dark"] .search-bar {
  background: var(--color-pink-100);
}
:root[data-theme="dark"] .search-bar.active {
  background: var(--color-cream);
}
.search-bar input {
  flex: 1 1 auto;
  border: none;
  background: transparent;
  outline: none;
  font-size: 14px;
  color: var(--color-ink);
  min-height: 0;
  padding: 0;
}
.search-bar input::-webkit-search-cancel-button,
.search-bar input::-webkit-search-decoration { -webkit-appearance: none; appearance: none; }
.search-icon { opacity: 0.55; font-size: 14px; line-height: 1; }
.search-clear {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.06);
  color: var(--color-muted);
  font-size: 11px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  min-height: 0;
  padding: 0;
  transition: background 0.15s ease;
}
.search-clear:hover { background: rgba(0, 0, 0, 0.12); }
:root[data-theme="dark"] .search-clear { background: rgba(255, 255, 255, 0.12); }

/* —— Tab 胶囊 —— */
.tab-bar {
  position: relative;
  display: inline-flex;
  padding: 4px;
  background: var(--color-pink-50);
  border-radius: var(--radius-pill);
  align-self: flex-start;
}
:root[data-theme="dark"] .tab-bar {
  background: rgba(236, 125, 166, 0.14);
}
.tab-bar button {
  position: relative;
  z-index: 1;
  padding: 8px 18px;
  border-radius: var(--radius-pill);
  background: transparent;
  border: none;
  font-size: 13.5px;
  color: var(--color-muted);
  cursor: pointer;
  font-weight: 600;
  transition: color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-height: 0;
  min-width: 0;
}
.tab-bar button.active {
  color: var(--color-pink-500);
  background: #fff;
  box-shadow: var(--shadow-sm);
  font-weight: 700;
}
:root[data-theme="dark"] .tab-bar button.active {
  background: #3a2830;
}
.tab-bar .badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 6px;
  background: var(--color-pink-200);
  color: #fff;
  border-radius: var(--radius-pill);
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}
.tab-bar button.active .badge { background: var(--color-pink-500); }

/* —— Tab + 筛选触发器 同行 —— */
.tab-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}
.filter-triggers {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}
.filter-trigger-wrap {
  position: relative;
}
.filter-trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 7px 12px;
  background: var(--color-cream);
  color: var(--color-muted);
  border: 1.5px solid transparent;
  border-radius: var(--radius-pill);
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  min-height: 36px;
  transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease;
}
.filter-trigger:hover:not(:disabled) {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
}
.filter-trigger.active {
  background: var(--color-pink-100);
  color: var(--color-pink-600);
  border-color: var(--color-pink-300);
  font-weight: 600;
}
.filter-trigger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
:root[data-theme="dark"] .filter-trigger {
  background: var(--color-pink-100);
}
:root[data-theme="dark"] .filter-trigger.active {
  background: rgba(236, 125, 166, 0.18);
}
.filter-trigger .caret {
  font-size: 9px;
  margin-left: 2px;
  transition: transform 0.2s var(--ease-spring);
  display: inline-block;
}
.filter-trigger .caret.up { transform: rotate(180deg); }

.filter-popover {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  z-index: 30;
  min-width: 160px;
  background: #fff;
  border: 1px solid var(--color-line);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
:root[data-theme="dark"] .filter-popover {
  background: var(--color-cream);
  border-color: var(--color-line-2);
}
.filter-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-ink);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  min-height: 0;
  min-width: 0;
  transition: background 0.12s ease;
}
.filter-option:hover {
  background: var(--color-pink-50);
  color: var(--color-pink-600);
}
.filter-option.active {
  background: var(--color-pink-100);
  color: var(--color-pink-600);
  font-weight: 600;
}
.filter-option.wide {
  justify-content: space-between;
  gap: 12px;
}
.filter-option .option-label { flex: 1 1 auto; }
.filter-option .option-hint {
  font-size: 11px;
  color: var(--color-muted);
}

/* —— 激活筛选条件行 —— */
.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  font-size: 12px;
  background: var(--color-pink-100);
  color: var(--color-pink-600);
  border-radius: var(--radius-pill);
  font-weight: 600;
  cursor: pointer;
  border: none;
  min-height: 0;
  min-width: 0;
  transition: background 0.12s ease;
}
.filter-chip:hover { background: var(--color-pink-200); color: #fff; }
.filter-chip .chip-x { opacity: 0.6; font-size: 10px; }
.filter-clear-all {
  font-size: 11.5px;
  color: var(--color-muted);
  background: transparent;
  padding: 4px 8px;
  text-decoration: underline;
  min-height: 0;
  min-width: 0;
}
.filter-clear-all:hover { color: var(--color-pink-500); }

/* Popover transition(共享) */
.popover-enter-active, .popover-leave-active {
  transition: opacity 0.18s ease, transform 0.22s var(--ease-spring);
  transform-origin: top left;
}
.popover-enter-from, .popover-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

/* —— 列表 + fade-up —— */
.list-section { display: flex; flex-direction: column; gap: 10px; }
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.load-more {
  align-self: center;
  margin-top: 6px;
  padding: 8px 18px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--color-pink-500);
  background: var(--color-pink-50);
  border-radius: var(--radius-pill);
  border: 1.5px solid transparent;
  cursor: pointer;
  min-height: 0;
  min-width: 0;
  transition: background 0.15s ease, border-color 0.15s ease, transform 0.15s var(--ease-spring);
}
.load-more:hover {
  background: var(--color-pink-100);
  border-color: var(--color-pink-300);
}
.load-more:active { transform: scale(0.97); }
:root[data-theme="dark"] .load-more {
  background: var(--color-pink-100);
}
.fade-up-enter-active, .fade-up-leave-active {
  transition: opacity 0.2s ease, transform 0.2s var(--ease-spring);
}
.fade-up-enter-from { opacity: 0; transform: translateY(8px); }
.fade-up-leave-to { opacity: 0; transform: translateY(-4px); }

/* —— 空状态 —— */
.empty {
  text-align: center;
  padding: 36px 16px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--color-muted);
}
.empty-title {
  font-family: var(--font-display);
  font-size: 18px;
  color: var(--color-pink-500);
  margin: 4px 0 0;
}
.empty-sub {
  font-size: 13.5px;
  line-height: 1.7;
  color: var(--color-muted);
  margin: 0;
}

/* —— 全部完成庆祝态 —— */
.empty.all-done {
  animation: pop 0.6s var(--ease-spring);
  background: linear-gradient(135deg, #fff0f5 0%, #fff4d4 100%);
  border-radius: var(--radius-lg);
  padding: 48px 20px 36px;
  box-shadow: var(--shadow-sm);
  gap: 14px;
}
:root[data-theme="dark"] .empty.all-done {
  background: linear-gradient(135deg, #3a2830 0%, #4a3a28 100%);
}
.empty.all-done .confetti-burst {
  font-size: 64px;
  line-height: 1;
  animation: pop 0.8s var(--ease-spring), bounce 1.6s ease-in-out 0.8s infinite;
  filter: drop-shadow(0 4px 12px rgba(236, 125, 166, 0.35));
}
.empty.all-done .empty-title {
  font-size: 20px;
  color: var(--color-pink-500);
}
.empty.all-done .empty-sub {
  color: var(--color-peach-300);
  font-weight: 500;
}
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50%      { transform: translateY(-6px); }
}

.loading {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  align-self: center;
  color: var(--color-muted);
  font-size: 13px;
  padding: 16px;
}

.error {
  background: var(--color-pink-100);
  color: var(--color-danger);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
}
</style>