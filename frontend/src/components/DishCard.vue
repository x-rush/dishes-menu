<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useMenuStore } from '../stores/menu'
import { useDishesStore } from '../stores/dishes'
import SlotBowl from './illustrations/SlotBowl.vue'
import EmptyPlate from './illustrations/EmptyPlate.vue'
import DishDetailDialog from './DishDetailDialog.vue'
import { SLOT_LABELS, type Day, type Slot, type MenuItem } from '../types'
import { useLongPress } from '../composables/useLongPress'
import { useUndo } from '../composables/useUndo'

const props = defineProps<{
  week: string
  day: Day
  slot: Slot
  items: MenuItem[]
}>()

const emit = defineEmits<{
  pick: []
}>()

const menu = useMenuStore()
const dishes = useDishesStore()
const undo = useUndo()

const cardRef = ref<HTMLElement | null>(null)
const shuffling = ref(false)
const latestSeq = ref<number | null>(null)  // 最新加入的 item.seq,用于触发 flip-in 动画
const noteOpen = ref(false)
const noteText = ref('')
const noteTargetSeq = ref<number | null>(null)
const noteDialogRef = ref<HTMLDialogElement | null>(null)
const error = ref<string | null>(null)
const detailOpen = ref(false)
const detailDishId = ref<string | null>(null)

function getDish(item: MenuItem) {
  return dishes.getById(item.dish_id)
}

const detailDish = computed(() =>
  detailDishId.value ? dishes.getById(detailDishId.value) ?? null : null
)

const slotLabel = computed(() => SLOT_LABELS[props.slot])
const isEmpty = computed(() => props.items.length === 0)

async function onShuffle() {
  if (shuffling.value) return
  shuffling.value = true
  error.value = null
  try {
    const newDish = await menu.shuffleDish(props.week, props.day, props.slot)
    // 找到刚加入的 item 的 seq
    const items = menu.weekMenus[props.week]?.[props.day]?.[props.slot] ?? []
    const added = items[items.length - 1]
    if (added) {
      latestSeq.value = added.seq
      // 动画 1.2s 后清掉
      const seqSnapshot = added.seq
      setTimeout(() => {
        if (latestSeq.value === seqSnapshot) latestSeq.value = null
      }, 1200)
    }
    undo.push(`已加入 ${newDish.name}`, async () => {
      // 撤销:删掉 slot 最后一条
      const cur = menu.weekMenus[props.week]?.[props.day]?.[props.slot] ?? []
      const last = cur[cur.length - 1]
      if (last) await menu.removeItem(props.week, props.day, props.slot, last.seq)
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : '换一换失败'
  } finally {
    shuffling.value = false
  }
}

function onPick() {
  emit('pick')
}

function openDetail(dishId: string) {
  detailDishId.value = dishId
  detailOpen.value = true
}

function closeDetail() {
  detailOpen.value = false
  detailDishId.value = null
}

function openNote(item: MenuItem) {
  noteTargetSeq.value = item.seq
  noteText.value = item.note ?? ''
  noteOpen.value = true
  nextTick(() => noteDialogRef.value?.showModal())
}

function closeNote() {
  noteOpen.value = false
  noteDialogRef.value?.close()
  noteTargetSeq.value = null
}

function onNoteBackdropClick(e: MouseEvent) {
  if (e.target === noteDialogRef.value) closeNote()
}

async function saveNote() {
  if (noteTargetSeq.value === null) {
    closeNote()
    return
  }
  await menu.updateItemNote(
    props.week,
    props.day,
    props.slot,
    noteTargetSeq.value,
    noteText.value.trim()
  )
  closeNote()
}

async function removeItem(item: MenuItem) {
  const captured = { dish_id: item.dish_id, note: item.note ?? '' }
  const dishName = getDish(item)?.name ?? '菜品'
  await menu.removeItem(props.week, props.day, props.slot, item.seq)
  undo.push(`已删除 ${dishName}`, async () => {
    await menu.appendItem(props.week, props.day, props.slot, captured.dish_id, captured.note)
  })
}

// 长按 = 换一换(vibrate 50ms 由 useLongPress 内部处理)
useLongPress(cardRef, { onLong: onShuffle })
</script>

<template>
  <article ref="cardRef" class="dish-card card" :class="{ empty: isEmpty }">
    <header class="head">
      <div class="slot-line">
        <SlotBowl :slot="slot" :size="32" class="slot-icon" />
        <span class="slot-label">{{ slotLabel }}</span>
        <span v-if="!isEmpty" class="count-badge">{{ items.length }} 道</span>
      </div>
    </header>

    <div v-if="isEmpty" class="body empty-body">
      <EmptyPlate :size="64" />
      <p class="empty-text">还没安排,要不要来一道?</p>
      <div class="actions empty-actions">
        <button class="btn btn-primary" @click="onPick">
          ✨ 选菜品
        </button>
        <button class="btn btn-ghost" :disabled="shuffling" @click="onShuffle">
          <span v-if="shuffling" class="spinner" aria-hidden="true"></span>
          <span v-else>🎲 随便来</span>
        </button>
      </div>
    </div>

    <template v-else>
      <div class="body item-list">
        <div
          v-for="(item, idx) in items"
          :key="`${item.dish_id}-${item.seq}-${idx}`"
          :class="['item-row', { 'is-new': item.seq === latestSeq }]"
        >
          <button
            type="button"
            class="dish-name-row"
            :aria-label="`查看 ${getDish(item)?.name ?? ''} 的详情`"
            :disabled="!getDish(item)"
            @click="openDetail(item.dish_id)"
          >
            <img
              v-if="getDish(item)?.image"
              :src="getDish(item)!.image"
              :alt="getDish(item)!.name"
              class="dish-thumb"
              loading="lazy"
              referrerpolicy="no-referrer"
            />
            <h3 class="dish-name">{{ getDish(item)?.name ?? '（菜品已删除）' }}</h3>
          </button>
          <p v-if="getDish(item)?.note" class="dish-note">📝 {{ getDish(item)!.note }}</p>
          <p v-if="item.note" class="slot-note">💬 {{ item.note }}</p>
          <div class="item-actions">
            <button class="btn btn-ghost" @click="openNote(item)">📝 备注</button>
            <button class="btn btn-icon" :aria-label="`删除 ${getDish(item)?.name ?? ''}`" @click="removeItem(item)">×</button>
          </div>
          <hr v-if="idx < items.length - 1" class="item-sep" />
        </div>
      </div>
    </template>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="!isEmpty" class="actions footer-actions">
      <button class="btn btn-primary" @click="onPick">✨ 再选一道</button>
      <button class="btn btn-ghost" :disabled="shuffling" @click="onShuffle">
        <span v-if="shuffling" class="spinner" aria-hidden="true"></span>
        <span v-else>🎲 随便来</span>
      </button>
    </div>

    <dialog ref="noteDialogRef" @click="onNoteBackdropClick" @close="noteOpen = false">
      <div class="dlg-body">
        <header class="dlg-header">
          <h3>本次备注</h3>
          <button class="btn btn-ghost" @click="closeNote">取消</button>
        </header>
        <p class="dlg-hint">比如:少辣、不吃香菜、想吃热的…</p>
        <textarea
          v-model="noteText"
          rows="4"
          placeholder="写一句…"
          maxlength="100"
        ></textarea>
        <div class="dlg-footer">
          <span class="counter">{{ noteText.length }}/100</span>
          <button class="btn btn-primary" :disabled="noteTargetSeq === null" @click="saveNote">保存</button>
        </div>
      </div>
    </dialog>

    <DishDetailDialog
      :open="detailOpen"
      :dish="detailDish"
      :show-pick-button="false"
      @close="closeDetail"
    />
  </article>
</template>

<style scoped>
.dish-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: relative;
  overflow: hidden;
  animation: fadeUp 0.5s var(--ease-out-soft) var(--enter-delay, 0ms) backwards;
  touch-action: pan-y;  /* 允许垂直滚动,水平留给 useSwipe */
}
.dish-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: linear-gradient(180deg, var(--color-pink-300), var(--color-pink-500));
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  opacity: 0.5;
}
.dish-card:hover { box-shadow: var(--shadow-md); }
.dish-card.empty { background: var(--color-pink-100); }
.dish-card.empty::before { background: linear-gradient(180deg, var(--color-butter-200), var(--color-butter-300)); opacity: 0.7; }

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.slot-line {
  display: flex;
  align-items: center;
  gap: 8px;
}
.slot-icon {
  color: var(--color-pink-400);
}
.slot-label {
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--color-pink-500);
  font-family: var(--font-display);
}
.count-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  background: var(--color-pink-100);
  color: var(--color-pink-500);
  border-radius: var(--radius-pill);
}

.body { display: flex; flex-direction: column; gap: 8px; }
.empty-body {
  align-items: center;
  padding: 8px 0 4px;
  gap: 12px;
}
.empty-text {
  color: var(--color-muted);
  font-size: 14px;
  text-align: center;
}

.item-list { gap: 4px; }
.item-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 6px 0;
}
/* shuffle 加入新菜时的翻牌动画 */
.item-row.is-new {
  animation: flip-in 0.5s var(--ease-spring);
}
@keyframes flip-in {
  0%   { transform: scale(0.85) rotateX(-22deg); opacity: 0; }
  60%  { transform: scale(1.04) rotateX(4deg); opacity: 1; }
  100% { transform: scale(1) rotateX(0); }
}
.item-sep {
  border: none;
  border-top: 1px dashed var(--color-line);
  margin: 4px 0;
}
.item-actions {
  display: flex;
  gap: 6px;
  margin-top: 2px;
  align-items: center;
}
.item-actions .btn { padding: 4px 10px; font-size: 12px; flex: 0 0 auto; }
.item-actions .btn-icon { margin-left: auto; }

.dish-name {
  font-size: 19px;
  font-weight: 700;
  margin: 0;
  word-break: break-word;
  font-family: var(--font-display);
}

.dish-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  text-align: left;
  font: inherit;
  color: inherit;
  min-width: 0;
  width: 100%;
  border-radius: var(--radius-sm);
  transition: opacity 0.15s ease;
}
.dish-name-row:disabled { cursor: default; }
.dish-name-row:not(:disabled):hover { opacity: 0.78; }
.dish-name-row:not(:disabled):active { transform: scale(0.99); }
.dish-name-row:focus-visible {
  outline: 2px solid var(--color-pink-400);
  outline-offset: 4px;
}

.dish-thumb {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  object-fit: cover;
  flex: 0 0 auto;
  background: var(--color-cream);
  box-shadow: var(--shadow-sm);
}

.dish-note, .slot-note {
  font-size: 13px;
  color: var(--color-muted);
  background: var(--color-warm-bg);
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--color-pink-200);
  word-break: break-word;
  white-space: pre-wrap;
}
.slot-note {
  background: var(--color-pink-50);
  border-left-color: var(--color-pink-400);
  color: var(--color-ink);
}

.error {
  color: var(--color-danger);
  font-size: 13px;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}
.actions .btn { flex: 1 1 auto; min-width: 90px; padding: 8px 14px; font-size: 13px; }
.actions .btn-icon { flex: 0 0 auto; }
.footer-actions {
  border-top: 1px dashed var(--color-line);
  padding-top: 10px;
  margin-top: 4px;
}

.dlg-body { padding: 20px; }
.dlg-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.dlg-header h3 { font-size: 17px; }
.dlg-header .btn { padding: 6px 14px; font-size: 13px; }
.dlg-hint { font-size: 12px; color: var(--color-muted); margin-bottom: 12px; }
.dlg-body textarea {
  width: 100%;
  resize: vertical;
  font-size: 15px;
}
.dlg-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 14px;
}
.counter { font-size: 12px; color: var(--color-muted); }
</style>
