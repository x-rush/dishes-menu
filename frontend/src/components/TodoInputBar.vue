<script setup lang="ts">
import { ref, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import EmojiPicker from './EmojiPicker.vue'
import DateChipBar from './DateChipBar.vue'

const store = useTodoStore()
const content = ref('')
const dueDate = ref<string | null>(null)
const emoji = ref<string>(localStorage.getItem('todo:emoji') ?? '🌸')
const color = ref<string>(localStorage.getItem('todo:color') ?? '#ec7da6')

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
}
</script>

<template>
  <div class="input-bar">
    <div class="top-row">
      <button
        type="button"
        class="sig-avatar"
        :style="{ background: color }"
        :title="`当前签名:${emoji}`"
      >{{ emoji }}</button>

      <input
        v-model="content"
        class="content-input"
        placeholder="想做点什么?写一句…"
        maxlength="500"
        @keydown.enter="submit"
      />

      <button
        class="send-btn"
        :class="{ active: canSubmit }"
        :disabled="!canSubmit"
        :aria-label="canSubmit ? '添加' : '写点内容才能加'"
        @click="submit"
      >
        <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
          <path
            d="M5 12h12M13 6l6 6-6 6"
            fill="none"
            stroke="currentColor"
            stroke-width="2.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </div>

    <div class="bottom-row">
      <div class="emoji-row">
        <span class="row-label">签名</span>
        <EmojiPicker :model-value="emoji" @update:model-value="pickEmoji" />
      </div>
      <div class="date-row">
        <span class="row-label">截止</span>
        <DateChipBar v-model="dueDate" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.input-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 14px 12px;
  background: linear-gradient(180deg, #fff8f3 0%, #fff0f5 100%);
  border-radius: var(--radius-lg);
  border: 1.5px solid rgba(255, 192, 220, 0.4);
  box-shadow: var(--shadow-md);
}
:root[data-theme="dark"] .input-bar {
  background: linear-gradient(180deg, #2d2226 0%, #3a2830 100%);
  border-color: rgba(107, 74, 90, 0.4);
}

.top-row {
  display: flex;
  gap: 10px;
  align-items: center;
}
.sig-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 22px;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 3px 8px rgba(0,0,0,0.08), inset 0 -2px 3px rgba(0,0,0,0.1);
  border: 2px solid #fff8f3;
  transition: transform 0.15s var(--ease-spring);
}
.sig-avatar:active { transform: scale(0.92); }
:root[data-theme="dark"] .sig-avatar {
  border-color: #2d2226;
}

.content-input {
  flex: 1 1 auto;
  background: rgba(255, 255, 255, 0.65);
  border: 1.5px solid rgba(235, 196, 190, 0.6);
  border-radius: var(--radius-pill);
  padding: 10px 16px;
  font-size: 15px;
  min-height: 40px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}
.content-input:focus {
  background: #fff;
  border-color: var(--color-pink-300);
  box-shadow: 0 0 0 3px rgba(255, 176, 206, 0.2);
}
:root[data-theme="dark"] .content-input {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(235, 196, 190, 0.2);
  color: var(--color-ink);
}
:root[data-theme="dark"] .content-input:focus {
  background: rgba(255, 255, 255, 0.1);
}

.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex: 0 0 auto;
  background: var(--color-pink-100);
  color: var(--color-pink-300);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.15s var(--ease-spring), background 0.2s ease, box-shadow 0.2s ease;
}
.send-btn.active {
  background: linear-gradient(135deg, var(--color-pink-400), var(--color-pink-500));
  color: #fff;
  box-shadow: 0 4px 12px rgba(236, 125, 166, 0.4);
}
.send-btn.active:active { transform: scale(0.92); }
.send-btn:disabled { cursor: not-allowed; }

.bottom-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 10px;
  border-top: 1px dashed rgba(243, 216, 211, 0.7);
}
:root[data-theme="dark"] .bottom-row {
  border-top-color: rgba(74, 58, 66, 0.7);
}
.emoji-row, .date-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.row-label {
  font-size: 11px;
  color: var(--color-muted);
  letter-spacing: 0.06em;
  font-weight: 600;
  flex: 0 0 28px;
  text-transform: uppercase;
}
</style>
