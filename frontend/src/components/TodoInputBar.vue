<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import EmojiPicker from './EmojiPicker.vue'
import DateChipBar from './DateChipBar.vue'

const store = useTodoStore()
const content = ref('')
const dueDate = ref<string | null>(null)
const emoji = ref<string>(localStorage.getItem('todo:emoji') ?? '🌸')
const color = ref<string>(localStorage.getItem('todo:color') ?? '#ec7da6')
const expanded = ref(false)

const EMOJI_COLORS: Record<string, string> = {
  '🌸': '#ec7da6', '🌷': '#d65a8a', '🌺': '#e26d6d', '🌻': '#f7d560',
  '🌼': '#ffc9a8', '🐱': '#ffa980', '🐶': '#d2a679', '🐰': '#e8d5e0',
  '🍓': '#e26d6d', '☕': '#a0826d', '🍡': '#ffc7dc', '🎁': '#7dd3a3',
}

function pickEmoji(e: string) {
  emoji.value = e
  color.value = EMOJI_COLORS[e] ?? '#ec7da6'
  localStorage.setItem('todo:emoji', e)
  localStorage.setItem('todo:color', color.value)
}

watch(emoji, (e) => {
  color.value = EMOJI_COLORS[e] ?? '#ec7da6'
})

const canSubmit = computed(() => content.value.trim().length > 0)

async function submit() {
  if (!canSubmit.value) return
  await store.create({
    content: content.value.trim(),
    due_date: dueDate.value,
    author_emoji: emoji.value,
    author_color: color.value,
  })
  content.value = ''
  dueDate.value = null
  expanded.value = false
}
</script>

<template>
  <div :class="['input-bar card', { expanded }]">
    <div class="top-row">
      <button
        type="button"
        class="sig-btn"
        :title="`当前签名:${emoji} (点选切换)`"
        :style="{ background: color }"
        @click="expanded = !expanded"
      >{{ emoji }}</button>

      <input
        v-model="content"
        class="content-input"
        placeholder="想做点什么?写一句…"
        maxlength="500"
        @keydown.enter="submit"
        @focus="expanded = true"
      />

      <button class="btn btn-primary submit" :disabled="!canSubmit" @click="submit">＋</button>
    </div>

    <div v-if="expanded" class="extras">
      <div class="extras-row">
        <span class="extras-label">签名:</span>
        <EmojiPicker :model-value="emoji" @update:model-value="pickEmoji" />
      </div>
      <div class="extras-row">
        <span class="extras-label">截止:</span>
        <DateChipBar v-model="dueDate" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.input-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--color-pink-50);
  border: 1.5px solid var(--color-pink-100);
}
.top-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.sig-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 22px;
  flex: 0 0 auto;
  filter: brightness(1.05);
  transition: transform 0.15s var(--ease-spring);
}
.sig-btn:active { transform: scale(0.92); }
.content-input {
  flex: 1 1 auto;
  background: var(--color-cream);
  border: 1.5px solid var(--color-line-2);
  border-radius: var(--radius-sm);
  padding: 10px 14px;
  font-size: 15px;
}
.submit {
  width: 44px;
  height: 40px;
  font-size: 22px;
  padding: 0;
  flex: 0 0 auto;
}
.extras {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 6px;
  border-top: 1px dashed var(--color-line);
}
.extras-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.extras-label {
  font-size: 12px;
  color: var(--color-muted);
  width: 36px;
  flex: 0 0 auto;
}
</style>
