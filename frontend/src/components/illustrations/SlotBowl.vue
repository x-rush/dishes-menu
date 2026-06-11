<script setup lang="ts">
// 时段图标: prop 决定早/午/晚/加餐
// 用法: <SlotBowl slot="breakfast" :size="32" />
import { computed } from 'vue'
import type { Slot } from '../../types'

const props = defineProps<{ slot: Slot; size?: number | string }>()

const iconFor: Record<Slot, 'sun' | 'sun-peak' | 'moon' | 'cookie'> = {
  breakfast: 'sun',
  lunch: 'sun-peak',
  dinner: 'moon',
  snack: 'cookie',
}

const kind = computed(() => iconFor[props.slot])
const size = computed(() => props.size ?? 36)
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 48 48"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
    class="slot-bowl"
  >
    <!-- 碗 -->
    <path
      d="M6 26 Q6 42 24 42 Q42 42 42 26 Z"
      fill="currentColor"
      opacity="0.18"
    />
    <path
      d="M6 26 L42 26"
      stroke="currentColor"
      stroke-width="2.4"
      stroke-linecap="round"
      fill="none"
    />
    <path
      d="M9 28 Q12 38 24 38 Q36 38 39 28"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      fill="none"
      opacity="0.6"
    />

    <!-- 早餐: 太阳 + 光线 -->
    <template v-if="kind === 'sun'">
      <circle cx="24" cy="16" r="6" fill="currentColor" />
      <g stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
        <line x1="24" y1="4"  x2="24" y2="7"  />
        <line x1="24" y1="25" x2="24" y2="28" />
        <line x1="12" y1="16" x2="15" y2="16" />
        <line x1="33" y1="16" x2="36" y2="16" />
        <line x1="15.5" y1="7.5"  x2="17.5" y2="9.5" />
        <line x1="30.5" y1="22.5" x2="32.5" y2="24.5" />
        <line x1="15.5" y1="24.5" x2="17.5" y2="22.5" />
        <line x1="30.5" y1="9.5"  x2="32.5" y2="7.5" />
      </g>
    </template>

    <!-- 午餐: 顶天太阳 + 叶子 -->
    <template v-else-if="kind === 'sun-peak'">
      <circle cx="24" cy="14" r="5" fill="currentColor" />
      <g stroke="currentColor" stroke-width="1.6" stroke-linecap="round">
        <line x1="24" y1="3"  x2="24" y2="6" />
        <line x1="13" y1="14" x2="16" y2="14" />
        <line x1="32" y1="14" x2="35" y2="14" />
        <line x1="16" y1="6"  x2="18" y2="8" />
        <line x1="30" y1="20" x2="32" y2="22" />
      </g>
      <path
        d="M10 20 Q14 14 18 18 Q14 22 10 20 Z"
        fill="currentColor"
        opacity="0.7"
      />
    </template>

    <!-- 晚餐: 月亮 + 星星 -->
    <template v-else-if="kind === 'moon'">
      <path
        d="M28 8 a8 8 0 1 0 6 14 a6 6 0 0 1 -6 -14 z"
        fill="currentColor"
      />
      <circle cx="36" cy="22" r="1.4" fill="currentColor" />
      <circle cx="14" cy="18" r="1"   fill="currentColor" opacity="0.7" />
    </template>

    <!-- 加餐: 饼干 + 巧克力豆 -->
    <template v-else>
      <circle cx="24" cy="16" r="8" fill="currentColor" />
      <circle cx="21" cy="14" r="1.4" fill="#3a2e36" />
      <circle cx="27" cy="18" r="1.4" fill="#3a2e36" />
      <circle cx="24" cy="13" r="1.2" fill="#3a2e36" />
      <circle cx="22" cy="19" r="1"   fill="#3a2e36" />
    </template>
  </svg>
</template>

<style scoped>
.slot-bowl {
  display: block;
  flex: 0 0 auto;
}
</style>
