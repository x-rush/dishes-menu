<script setup lang="ts">
import { useTheme } from '../composables/useTheme'

const { isDark, toggle } = useTheme()
function onToggle() { toggle() }
</script>

<template>
  <button
    class="theme-toggle"
    type="button"
    :aria-label="isDark ? '切到亮色模式' : '切到暗色模式'"
    :aria-pressed="isDark"
    @click="onToggle"
  >
    <Transition name="theme-icon" mode="out-in">
      <span v-if="isDark" key="sun" class="icon">☀️</span>
      <span v-else key="moon" class="icon">🌙</span>
    </Transition>
  </button>
</template>

<style scoped>
.theme-toggle {
  position: fixed;
  /* 上移到 TabBar 之上(TabBar 高 68px + 12px 间距),不再挡菜单/待办 tab */
  left: max(16px, env(safe-area-inset-left));
  bottom: calc(68px + env(safe-area-inset-bottom) + 12px);
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  font-size: 18px;
  box-shadow: var(--shadow-sm);
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.18s ease, transform 0.15s var(--ease-spring);
}
.theme-toggle:hover {
  background: var(--color-pink-100);
}
.theme-toggle:active {
  transform: scale(0.92);
}

.icon {
  display: inline-block;
  line-height: 1;
}

.theme-icon-enter-active, .theme-icon-leave-active {
  transition: opacity 0.18s ease, transform 0.18s var(--ease-spring);
  position: absolute;
}
.theme-icon-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.4);
}
.theme-icon-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.4);
}
</style>
