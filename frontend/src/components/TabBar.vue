<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const active = computed<'menu' | 'todo'>(() => {
  return route.path.startsWith('/todo') ? 'todo' : 'menu'
})

function go(target: 'menu' | 'todo') {
  if (target === active.value) return
  if (target === 'todo') {
    router.push('/todo')
  } else {
    // 回菜单:保留当前的 week + date(如果是在 /menu 命名空间下),否则跳今天
    if (typeof route.params.week === 'string' && typeof route.params.date === 'string') {
      router.push(`/menu/${route.params.week}/${route.params.date}`)
    } else {
      // 兜底:跳 / 让 redirect 处理
      router.push('/')
    }
  }
}
</script>

<template>
  <nav class="tab-bar" role="tablist">
    <button
      type="button"
      role="tab"
      :aria-selected="active === 'menu'"
      :class="['tab', { active: active === 'menu' }]"
      @click="go('menu')"
    >
      <span class="tab-icon">🍱</span>
      <span class="tab-label">菜单</span>
    </button>
    <div class="tab-spacer"></div>
    <button
      type="button"
      role="tab"
      :aria-selected="active === 'todo'"
      :class="['tab', { active: active === 'todo' }]"
      @click="go('todo')"
    >
      <span class="tab-icon">💝</span>
      <span class="tab-label">待办</span>
    </button>
  </nav>
</template>

<style scoped>
.tab-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: calc(60px + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-cream);
  border-top: 1px solid var(--color-line);
  display: flex;
  align-items: stretch;
  z-index: 25;
  box-shadow: 0 -4px 12px rgba(0,0,0,0.04);
}
.tab {
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: transparent;
  color: var(--color-muted);
  font-size: 11px;
  font-weight: 600;
  min-height: 44px;
  transition: color 0.2s ease;
}
.tab.active {
  color: var(--color-pink-500);
}
.tab-icon {
  font-size: 24px;
  transition: transform 0.2s var(--ease-spring);
}
.tab.active .tab-icon {
  transform: translateY(-2px) scale(1.1);
}
.tab-spacer {
  flex: 0 0 90px;  /* 中间留位置给 heart FAB */
}
</style>
