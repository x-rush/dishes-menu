<script setup lang="ts">
// 空状态插画: 蒸汽飘飘的空盘子
// 用法: <EmptyPlate :size="64" />
defineProps<{ size?: number | string }>()
</script>

<template>
  <svg
    :width="size ?? 64"
    :height="size ?? 64"
    viewBox="0 0 80 80"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
    class="empty-plate"
  >
    <!-- 蒸汽 -->
    <path
      d="M28 22 Q26 14 30 8 Q34 14 32 22"
      stroke="currentColor"
      stroke-width="2.4"
      stroke-linecap="round"
      fill="none"
      opacity="0.55"
    />
    <path
      d="M40 18 Q38 10 42 4 Q46 10 44 18"
      stroke="currentColor"
      stroke-width="2.4"
      stroke-linecap="round"
      fill="none"
      opacity="0.55"
    />
    <path
      d="M52 22 Q50 14 54 8 Q58 14 56 22"
      stroke="currentColor"
      stroke-width="2.4"
      stroke-linecap="round"
      fill="none"
      opacity="0.55"
    />

    <!-- 盘子 -->
    <ellipse cx="40" cy="52" rx="30" ry="6" fill="currentColor" opacity="0.15" />
    <ellipse cx="40" cy="50" rx="30" ry="6" stroke="currentColor" stroke-width="2.4" fill="none" />
    <ellipse cx="40" cy="50" rx="22" ry="4" stroke="currentColor" stroke-width="1.6" fill="none" opacity="0.5" />

    <!-- 点点(碎屑装饰) -->
    <circle cx="22" cy="68" r="1.4" fill="currentColor" opacity="0.5" />
    <circle cx="58" cy="68" r="1.4" fill="currentColor" opacity="0.5" />
    <circle cx="40" cy="72" r="1.2" fill="currentColor" opacity="0.5" />
  </svg>
</template>

<style scoped>
.empty-plate {
  display: block;
  color: var(--color-pink-400);
  animation: float 3s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50%      { transform: translateY(-4px); }
}
@media (prefers-reduced-motion: reduce) {
  .empty-plate { animation: none; }
}
</style>
