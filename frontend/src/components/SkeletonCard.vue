<script setup lang="ts"></script>

<template>
  <article class="skeleton card" aria-hidden="true">
    <div class="head">
      <div class="skel-line short"></div>
    </div>
    <div class="body">
      <div class="skel-line medium"></div>
      <div class="skel-line long"></div>
      <div class="skel-line short"></div>
    </div>
  </article>
</template>

<style scoped>
.skeleton {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 140px;
}
.skel-line {
  height: 14px;
  border-radius: 6px;
  background: linear-gradient(90deg,
    var(--color-line) 0%,
    var(--color-line-2) 50%,
    var(--color-line) 100%);
  background-size: 200% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
}
.skel-line.short { width: 35%; }
.skel-line.medium { width: 60%; }
.skel-line.long { width: 90%; }

@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
