<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import TodoInputBar from '../components/TodoInputBar.vue'
import TodoCard from '../components/TodoCard.vue'

const store = useTodoStore()

onMounted(() => {
  store.fetchAll()
})

const sortedOpen = computed(() =>
  [...store.open].sort((a, b) => {
    // 有 due_date 排前;同 due_date 新的在前
    if (a.due_date && !b.due_date) return -1
    if (!a.due_date && b.due_date) return 1
    if (a.due_date && b.due_date) return a.due_date.localeCompare(b.due_date)
    return b.created_at.localeCompare(a.created_at)
  })
)
</script>

<template>
  <div class="todo-page">
    <header class="page-head">
      <h1>💝 想做的小事</h1>
      <p class="muted">写下你们想一起做的事</p>
    </header>

    <TodoInputBar />

    <section v-if="store.loading && !store.todos.length" class="empty">加载中…</section>
    <section v-else-if="!sortedOpen.length && !store.done.length" class="empty">
      <p>还没有任何待办。从上面写一条试试 ✏️</p>
    </section>

    <template v-else>
      <section v-if="sortedOpen.length" class="list-section">
        <h2 class="section-title">待办 · {{ sortedOpen.length }}</h2>
        <div class="todo-list">
          <TodoCard v-for="t in sortedOpen" :key="t.id" :todo="t" />
        </div>
      </section>

      <section v-if="store.done.length" class="list-section done-section">
        <h2 class="section-title">已完成 · {{ store.done.length }}</h2>
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
  gap: 16px;
  padding: 16px;
  max-width: 560px;
  margin: 0 auto;
}
.page-head h1 {
  font-size: 24px;
  font-family: var(--font-display);
  color: var(--color-pink-500);
  margin: 0 0 4px;
}
.muted { font-size: 13px; color: var(--color-muted); }

.section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-pink-500);
  font-family: var(--font-display);
  letter-spacing: 0.04em;
  margin: 8px 0 0;
}
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.done-section { margin-top: 8px; opacity: 0.85; }

.empty {
  text-align: center;
  padding: 32px 16px;
  color: var(--color-muted);
  font-size: 14px;
}
.error {
  background: var(--color-pink-100);
  color: var(--color-danger);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
}
</style>
