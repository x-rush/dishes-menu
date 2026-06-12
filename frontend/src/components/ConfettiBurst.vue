<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ show: boolean }>()

const COLORS = ['#f8a5c2', '#ec7da6', '#d65a8a', '#ffe799', '#f7d560', '#7dd3a3', '#ffa980']

const particles = computed(() =>
  Array.from({ length: 18 }, () => ({
    left: 20 + Math.random() * 60,
    delay: Math.random() * 0.25,
    duration: 1.2 + Math.random() * 0.6,
    drift: (Math.random() - 0.5) * 120,
    rotation: 360 + Math.random() * 540,
    color: COLORS[Math.floor(Math.random() * COLORS.length)],
    size: 6 + Math.random() * 6,
  }))
)
</script>

<template>
  <div v-if="show" class="confetti-burst" aria-hidden="true">
    <span
      v-for="(p, i) in particles"
      :key="i"
      class="confetti"
      :style="{
        left: p.left + '%',
        backgroundColor: p.color,
        width: p.size + 'px',
        height: p.size + 'px',
        animationDelay: p.delay + 's',
        animationDuration: p.duration + 's',
        '--drift': p.drift + 'px',
        '--rot': p.rotation + 'deg',
      }"
    ></span>
  </div>
</template>

<style scoped>
.confetti-burst {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 100;
  overflow: hidden;
}
.confetti {
  position: absolute;
  top: -20px;
  display: block;
  border-radius: 2px;
  animation: confetti-fall 1.5s ease-in forwards;
}

@keyframes confetti-fall {
  0% {
    transform: translate(0, 0) rotate(0);
    opacity: 1;
  }
  100% {
    transform: translate(var(--drift, 0), 110vh) rotate(var(--rot, 720deg));
    opacity: 0;
  }
}
</style>
