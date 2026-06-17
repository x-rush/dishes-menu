<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TodayTogether from './TodayTogether.vue'

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
    if (typeof route.params.week === 'string' && typeof route.params.date === 'string') {
      router.push(`/menu/${route.params.week}/${route.params.date}`)
    } else {
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
      <span class="tab-icon">
        <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
          <path
            d="M5 7h14a1 1 0 0 1 1 1v9.5a2.5 2.5 0 0 1-2.5 2.5h-11A2.5 2.5 0 0 1 4 17.5V8a1 1 0 0 1 1-1z"
            :fill="active === 'menu' ? 'currentColor' : 'none'"
            stroke="currentColor"
            stroke-width="1.6"
          />
          <path
            d="M4 11h16"
            stroke="currentColor"
            :stroke-opacity="active === 'menu' ? 0.4 : 0.7"
            stroke-width="1.4"
          />
          <circle cx="8.5" cy="15.5" r="0.9" fill="currentColor" />
          <circle cx="12" cy="15.5" r="0.9" fill="currentColor" />
          <circle cx="15.5" cy="15.5" r="0.9" fill="currentColor" />
          <path
            d="M9 4.5l1.4 1.6M15 4.5l-1.4 1.6"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            fill="none"
          />
        </svg>
      </span>
      <span class="tab-label">菜单</span>
    </button>

    <div class="tab-center">
      <TodayTogether />
    </div>

    <button
      type="button"
      role="tab"
      :aria-selected="active === 'todo'"
      :class="['tab', { active: active === 'todo' }]"
      @click="go('todo')"
    >
      <span class="tab-icon">
        <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
          <path
            d="M12 19s-6.5-4.35-6.5-9.5C5.5 6.46 7.46 4.5 10 4.5c1.74 0 3.41.81 4.5 2.09 1.09-1.28 2.76-2.09 4.5-2.09 2.54 0 4.5 1.96 4.5 5 0 5.15-6.5 9.5-6.5 9.5z"
            :fill="active === 'todo' ? 'currentColor' : 'none'"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
        </svg>
      </span>
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
  height: calc(68px + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
  background: rgba(255, 248, 243, 0.78);
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  backdrop-filter: saturate(180%) blur(20px);
  border-top: 1px solid rgba(243, 216, 211, 0.6);
  display: grid;
  grid-template-columns: 1fr 96px 1fr;
  align-items: stretch;
  z-index: 25;
}
:root[data-theme="dark"] .tab-bar {
  background: rgba(45, 34, 38, 0.78);
  border-top-color: rgba(74, 58, 66, 0.6);
}

.tab {
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
  padding: 4px 0 0;
}
.tab.active {
  color: var(--color-pink-500);
}
.tab-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  transition: transform 0.2s var(--ease-spring);
}
.tab.active .tab-icon {
  transform: translateY(-2px);
}
.tab.active .tab-icon svg {
  filter: drop-shadow(0 2px 4px rgba(236, 125, 166, 0.35));
}

.tab-center {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  position: relative;
}
</style>
