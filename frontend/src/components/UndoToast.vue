<script setup lang="ts">
import { useUndo } from '../composables/useUndo'

const { currentTask, performUndo, dismiss } = useUndo()
</script>

<template>
  <Transition name="toast">
    <div
      v-if="currentTask"
      class="undo-toast"
      role="status"
      aria-live="polite"
      @click="dismiss"
    >
      <span class="undo-msg">{{ currentTask.message }}</span>
      <button class="undo-btn" @click.stop="performUndo">撤销</button>
      <div class="undo-progress" aria-hidden="true"></div>
    </div>
  </Transition>
</template>

<style scoped>
.undo-toast {
  position: fixed;
  left: 50%;
  bottom: max(96px, calc(env(safe-area-inset-bottom) + 96px));
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  background: var(--color-ink);
  color: var(--color-warm-bg);
  border-radius: var(--radius-pill);
  box-shadow: var(--shadow-lg);
  z-index: 90;
  font-size: 14px;
  font-weight: 500;
  max-width: calc(100vw - 32px);
  overflow: hidden;
}

.undo-msg {
  flex: 0 1 auto;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.undo-btn {
  flex: 0 0 auto;
  background: var(--color-pink-400);
  color: #fff;
  font-weight: 700;
  font-size: 13px;
  padding: 5px 14px;
  border-radius: var(--radius-pill);
  min-height: 32px;
  transition: background 0.15s ease, transform 0.15s var(--ease-spring);
}
.undo-btn:hover { background: var(--color-pink-500); }
.undo-btn:active { transform: scale(0.94); }

/* 5s 进度条:CSS 动画 + forwards,正好和 useUndo 的 expiry 对齐 */
.undo-progress {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 2px;
  width: 100%;
  background: var(--color-pink-400);
  transform-origin: left;
  animation: progress-shrink 5s linear forwards;
}

@keyframes progress-shrink {
  from { transform: scaleX(1); }
  to   { transform: scaleX(0); }
}

/* 出现/消失:从下方滑入 */
.toast-enter-active, .toast-leave-active {
  transition: opacity 0.22s ease, transform 0.28s var(--ease-spring);
}
.toast-enter-from, .toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}
</style>
