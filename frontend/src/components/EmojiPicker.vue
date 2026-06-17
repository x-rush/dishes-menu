<script setup lang="ts">
const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [string] }>()

const EMOJIS = ['🌸','🌷','🌺','🌻','🌼','🐱','🐶','🐰','🍓','☕','🍡','🎁']

function pick(e: string) {
  emit('update:modelValue', e)
}
</script>

<template>
  <div class="emoji-picker" role="radiogroup" aria-label="选择你的 emoji 签名">
    <button
      v-for="e in EMOJIS"
      :key="e"
      type="button"
      role="radio"
      :aria-checked="props.modelValue === e"
      :class="['emoji-btn', { active: props.modelValue === e }]"
      @click="pick(e)"
    >{{ e }}</button>
  </div>
</template>

<style scoped>
.emoji-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.emoji-btn {
  width: 36px;
  height: 36px;
  font-size: 22px;
  border-radius: var(--radius-sm);
  background: var(--color-cream);
  transition: transform 0.15s var(--ease-spring), background 0.15s ease;
}
.emoji-btn:hover { background: var(--color-pink-50); }
.emoji-btn:active { transform: scale(0.9); }
.emoji-btn.active {
  background: var(--color-pink-100);
  box-shadow: 0 0 0 2px var(--color-pink-400);
}
</style>
