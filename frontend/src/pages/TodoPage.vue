<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import TodoInputBar from '../components/TodoInputBar.vue'
import TodoCard from '../components/TodoCard.vue'
import Notebook from '../components/illustrations/Notebook.vue'

const store = useTodoStore()

onMounted(() => {
  store.fetchAll()
})

const sortedOpen = computed(() =>
  [...store.open].sort((a, b) => {
    if (a.due_date && !b.due_date) return -1
    if (!a.due_date && b.due_date) return 1
    if (a.due_date && b.due_date) return a.due_date.localeCompare(b.due_date)
    return b.created_at.localeCompare(a.created_at)
  })
)

const totalCount = computed(() => store.todos.length)
const openCount = computed(() => sortedOpen.value.length)
const doneCount = computed(() => store.done.length)
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
    </header>

    <TodoInputBar />

    <section v-if="store.loading && !store.todos.length" class="loading">
      <span class="spinner"></span>
      <span>加载中…</span>
    </section>

    <section v-else-if="!sortedOpen.length && !doneCount" class="empty">
      <Notebook :size="160" />
      <p class="empty-title">这本本子还空着</p>
      <p class="empty-sub">在下面写一句想做的小事,<br />我们一起慢慢打勾 ✏️</p>
    </section>

    <template v-else>
      <section v-if="sortedOpen.length" class="list-section">
        <h2 class="section-title">
          <span>待办</span>
          <span class="section-line"></span>
        </h2>
        <div class="todo-list">
          <TodoCard v-for="t in sortedOpen" :key="t.id" :todo="t" />
        </div>
      </section>

      <section v-if="doneCount" class="list-section done-section">
        <h2 class="section-title">
          <span>已完成</span>
          <span class="section-line"></span>
        </h2>
        <div class="todo-list">
          <TodoCard v-for="t in store.done" :key="t.id" :todo="t" />
        </div>
      </section>
    </template>

    <div v-if="store.error" class="error">{{ store.error }}</div>
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

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  font-weight: 700;
  color: var(--color-pink-500);
  font-family: var(--font-display);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  margin: 6px 0 0;
}
.section-line {
  flex: 1 1 auto;
  height: 1px;
  background: linear-gradient(90deg, var(--color-pink-200), transparent);
}
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.done-section { margin-top: 4px; }

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
