<script setup lang="ts">
// 招手的可爱小狗头特写 — 全 SVG,currentColor 跟随主题色
// 用法: <Mascot :size="80" />
defineProps<{ size?: number | string }>()
</script>

<template>
  <svg
    :width="size ?? 96"
    :height="size ?? 96"
    viewBox="0 0 100 100"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
    class="mascot"
  >
    <!-- 头顶呆毛 -->
    <path
      d="M48 18 Q50 12 52 18 Q54 14 56 20"
      stroke="currentColor"
      stroke-width="2.4"
      fill="none"
      stroke-linecap="round"
    />
    <!-- 左耳(垂下三角) -->
    <path
      d="M22 30 Q8 38 12 62 Q18 76 30 70 Q34 60 30 38 Z"
      fill="currentColor"
      opacity="0.9"
    />
    <!-- 右耳(垂下三角) -->
    <path
      d="M78 30 Q92 38 88 62 Q82 76 70 70 Q66 60 70 38 Z"
      fill="currentColor"
      opacity="0.9"
    />
    <!-- 左内耳粉色 -->
    <path
      d="M22 38 Q14 44 17 60 Q21 68 27 64 Q28 50 26 40 Z"
      fill="#ff8fb1"
      opacity="0.45"
    />
    <!-- 右内耳粉色 -->
    <path
      d="M78 38 Q86 44 83 60 Q79 68 73 64 Q72 50 74 40 Z"
      fill="#ff8fb1"
      opacity="0.45"
    />
    <!-- 脸(圆胖,带颊鼓) -->
    <path
      d="M20 50 Q20 28 50 28 Q80 28 80 50 Q80 86 50 86 Q20 86 20 50 Z"
      fill="currentColor"
      opacity="0.95"
    />
    <!-- 白色口鼻区(浅色脸) -->
    <ellipse cx="50" cy="74" rx="18" ry="10" fill="#fff5e8" opacity="0.55" />
    <!-- 左眼 -->
    <circle cx="38" cy="54" r="3.6" fill="#2d2226" />
    <!-- 右眼 -->
    <circle cx="62" cy="54" r="3.6" fill="#2d2226" />
    <!-- 左眼高光 -->
    <circle cx="39.4" cy="52.6" r="1.2" fill="#fff" />
    <!-- 右眼高光 -->
    <circle cx="63.4" cy="52.6" r="1.2" fill="#fff" />
    <!-- 鼻子 -->
    <ellipse cx="50" cy="66" rx="3.6" ry="2.6" fill="#2d2226" />
    <!-- 鼻子到嘴的竖线 -->
    <line x1="50" y1="68.6" x2="50" y2="72" stroke="#2d2226" stroke-width="1.6" stroke-linecap="round" />
    <!-- 嘴(微笑 Y 形) -->
    <path
      d="M50 72 Q46 76 42.5 74.5"
      stroke="#2d2226"
      stroke-width="1.6"
      fill="none"
      stroke-linecap="round"
    />
    <path
      d="M50 72 Q54 76 57.5 74.5"
      stroke="#2d2226"
      stroke-width="1.6"
      fill="none"
      stroke-linecap="round"
    />
    <!-- 舌头(从嘴里探出) -->
    <path
      d="M48 74.5 Q50 78 52 74.5 Q52 76 50 77 Q48 76 48 74.5 Z"
      fill="#ff8fb1"
      opacity="0.85"
    />
    <!-- 左腮红 -->
    <ellipse cx="28" cy="66" rx="4.4" ry="2.6" fill="#ff8fb1" opacity="0.6" />
    <!-- 右腮红 -->
    <ellipse cx="72" cy="66" rx="4.4" ry="2.6" fill="#ff8fb1" opacity="0.6" />
  </svg>
</template>

<style scoped>
.mascot {
  display: block;
  animation: mascot-bob 2.4s ease-in-out infinite;
}
@keyframes mascot-bob {
  0%, 100% { transform: translateY(0) rotate(-2deg); }
  50%      { transform: translateY(-3px) rotate(2deg); }
}
@media (prefers-reduced-motion: reduce) {
  .mascot { animation: none; }
}
</style>
