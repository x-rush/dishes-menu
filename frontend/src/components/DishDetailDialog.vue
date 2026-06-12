<script setup lang="ts">
// 菜品详情弹窗 — 点开查看做法 + 图片
// 用法:
//   <DishDetailDialog
//     :open="detailOpen"
//     :dish="dish"
//     :show-pick-button="true"
//     @close="detailOpen = false"
//     @pick="onPick"
//   />
import { ref, watch, nextTick } from 'vue'
import type { Dish } from '../types'

const props = defineProps<{
  open: boolean
  dish: Dish | null
  showPickButton?: boolean
}>()

const emit = defineEmits<{
  close: []
  pick: [dish: Dish]
}>()

const dialogRef = ref<HTMLDialogElement | null>(null)
const imageFailed = ref(false)

watch(() => props.open, async (isOpen) => {
  if (isOpen && props.dish) {
    imageFailed.value = false
    await nextTick()
    dialogRef.value?.showModal()
  } else {
    dialogRef.value?.close()
  }
})

function close() {
  emit('close')
}

function onPick() {
  if (props.dish) emit('pick', props.dish)
}

function onBackdropClick(e: MouseEvent) {
  if (e.target === dialogRef.value) close()
}

function onImageError() {
  imageFailed.value = true
}
</script>

<template>
  <dialog ref="dialogRef" @click="onBackdropClick" @close="close">
    <div v-if="dish" class="dlg-body">
      <header class="dlg-header">
        <h3>{{ dish.name }}</h3>
        <button class="btn btn-ghost" @click="close">关闭</button>
      </header>

      <div v-if="dish.image && !imageFailed" class="hero-image">
        <img
          :src="dish.image"
          :alt="dish.name"
          loading="lazy"
          referrerpolicy="no-referrer"
          @error="onImageError"
        />
      </div>

      <div class="meta-row">
        <span class="meta-time">🕐 约 {{ dish.estimated_time || 0 }} 分钟</span>
      </div>

      <div v-if="dish.ingredients.length" class="ingredients">
        <h4 class="meta-title">🥬 食材</h4>
        <div class="chips">
          <span v-for="ing in dish.ingredients" :key="ing" class="chip">{{ ing }}</span>
        </div>
      </div>

      <section v-if="dish.note" class="recipe">
        <h4>📝 做法 / 小贴士</h4>
        <p>{{ dish.note }}</p>
      </section>
      <p v-else class="no-recipe">这道菜还没有写做法,要不要补一下? ✏️</p>

      <footer v-if="showPickButton" class="dlg-footer">
        <button class="btn btn-primary" @click="onPick">选这道</button>
      </footer>
    </div>
  </dialog>
</template>

<style scoped>
.dlg-body {
  padding: 20px;
  min-width: min(320px, 90vw);
  max-width: 480px;
}
.dlg-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}
.dlg-header h3 {
  font-size: 20px;
  font-family: var(--font-display);
  word-break: break-word;
  min-width: 0;
  flex: 1 1 auto;
}
.dlg-header .btn {
  padding: 6px 14px;
  font-size: 13px;
  flex: 0 0 auto;
}

.hero-image {
  width: 100%;
  aspect-ratio: 16 / 10;
  max-height: 220px;
  border-radius: var(--radius-lg);
  overflow: hidden;
  margin-bottom: 14px;
  background: var(--color-cream);
  box-shadow: var(--shadow-sm);
}
.hero-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--color-muted);
}
.meta-time {
  background: var(--color-butter-100);
  color: var(--color-on-butter);
  padding: 4px 12px;
  border-radius: var(--radius-pill);
  font-weight: 600;
}

.ingredients {
  margin-bottom: 14px;
}
.meta-title {
  font-size: 13px;
  margin: 0 0 8px;
  color: var(--color-pink-500);
  font-family: var(--font-display);
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.chip {
  font-size: 13px;
  padding: 4px 12px;
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  border-radius: var(--radius-pill);
  font-weight: 600;
}

.recipe {
  background: var(--color-warm-bg);
  padding: 14px 16px;
  border-radius: var(--radius-md);
  border-left: 3px solid var(--color-pink-400);
  margin-bottom: 14px;
}
.recipe h4 {
  font-size: 13px;
  margin: 0 0 8px;
  color: var(--color-pink-500);
  font-family: var(--font-display);
}
.recipe p {
  font-size: 14px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
}

.no-recipe {
  color: var(--color-muted);
  font-size: 13px;
  text-align: center;
  padding: 12px;
  background: var(--color-cream);
  border-radius: var(--radius-md);
  margin-bottom: 14px;
}

.dlg-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
