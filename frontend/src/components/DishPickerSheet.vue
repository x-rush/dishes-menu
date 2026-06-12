<script setup lang="ts">
// 底部 sheet:为某 slot 选一道菜
// 用法:
//   <DishPickerSheet
//     :open="pickerOpen"
//     :week="currentWeek" :day="day" :slot="slot"
//     :current-dish-id="..."
//     @close="pickerOpen = false"
//   />
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { useMenuStore } from '../stores/menu'
import { useDishesStore } from '../stores/dishes'
import DishDetailDialog from './DishDetailDialog.vue'
import { SLOT_LABELS, type Day, type Slot, type Dish, type MenuItem } from '../types'

const props = defineProps<{
  open: boolean
  week: string
  day: Day
  slot: Slot
  pickedDishIds?: string[]
}>()

const emit = defineEmits<{
  close: []
  pick: [dish: Dish]
}>()

const menu = useMenuStore()
const dishesStore = useDishesStore()

const query = ref('')
const slotFilter = ref<Slot | null>(null)  // null = 全部,点同一个回 null
const searchInputRef = ref<HTMLInputElement | null>(null)

function toggleSlotFilter(s: Slot) {
  slotFilter.value = slotFilter.value === s ? null : s
}

// 本周已出现的菜品 id 集合(用于降权)
const weekUsedDishIds = computed<Set<string>>(() => {
  const set = new Set<string>()
  const wm = menu.weekMenus[props.week]
  if (!wm) return set
  for (const day of Object.keys(wm) as Day[]) {
    const dm = wm[day] ?? []
    for (const slot of Object.keys(dm) as Slot[]) {
      const list: MenuItem[] = dm[slot] ?? []
      for (const item of list) {
        if (item.dish_id) set.add(item.dish_id)
      }
    }
  }
  // 当前 slot 已选的菜不算"降权"(用户可能想再加同款)
  for (const id of props.pickedDishIds ?? []) set.delete(id)
  return set
})

// filteredDishes — 默认展示全部菜品(slot 只作辅助筛选,不是硬限制)
// 文本搜索:同时匹配 name 和 ingredients
// 时段过滤:slotFilter 选中时只显示适用该时段的菜;null = 全部
const filteredDishes = computed<Dish[]>(() => {
  const q = query.value.trim().toLowerCase()
  const slotFilterVal = slotFilter.value

  return dishesStore.dishes.filter((d) => {
    if (slotFilterVal && !d.slots.includes(slotFilterVal)) return false
    if (!q) return true
    if (d.name.toLowerCase().includes(q)) return true
    for (const ing of d.ingredients) {
      if (ing.toLowerCase().includes(q)) return true
    }
    return false
  })
})

async function pick(d: Dish) {
  emit('pick', d)
  await menu.appendItem(props.week, props.day, props.slot, d.id, '')
  emit('close')
}

function onClose() {
  emit('close')
}

const detailTarget = ref<Dish | null>(null)
function openDetail(d: Dish) {
  detailTarget.value = d
}
function closeDetail() {
  detailTarget.value = null
}
async function onDetailPick(d: Dish) {
  closeDetail()
  await pick(d)
}

// 打开时:重置 query + slotFilter + 自动聚焦搜索框
watch(() => props.open, async (isOpen) => {
  if (isOpen) {
    query.value = ''
    slotFilter.value = null
    document.body.style.overflow = 'hidden'
    await nextTick()
    searchInputRef.value?.focus()
  } else {
    document.body.style.overflow = ''
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
})

// 防止滚动穿透:触摸底部时禁用默认滚动
function onTouchMove(e: TouchEvent) {
  const target = e.target as HTMLElement
  if (target.closest('.dish-grid')) return
  e.preventDefault()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="sheet">
      <div v-if="open" class="sheet-backdrop" @click.self="onClose" @touchmove="onTouchMove">
        <div class="sheet" role="dialog" aria-label="选择菜品">
          <div class="handle" aria-hidden="true"></div>

          <header class="sheet-header">
            <h3>为「{{ SLOT_LABELS[slot] }}」选一道菜</h3>
            <button class="btn btn-ghost" @click="onClose">取消</button>
          </header>

          <div class="search-row">
            <input
              ref="searchInputRef"
              v-model="query"
              type="search"
              class="search-input"
              placeholder="🔍 搜菜名或食材…"
              autocomplete="off"
            />
          </div>

          <div class="slot-filter-row" role="toolbar" aria-label="按时段筛选">
            <span class="slot-filter-label">时段:</span>
            <button
              v-for="s in (['breakfast', 'lunch', 'dinner', 'snack'] as Slot[])"
              :key="s"
              type="button"
              :class="['slot-chip', { on: slotFilter === s }]"
              :aria-pressed="slotFilter === s"
              @click="toggleSlotFilter(s)"
            >{{ SLOT_LABELS[s] }}</button>
            <button
              v-if="slotFilter"
              type="button"
              class="slot-chip-clear"
              @click="slotFilter = null"
              aria-label="清除时段筛选"
            >全部 ✕</button>
          </div>

          <div v-if="filteredDishes.length === 0" class="empty-state">
            <p>没找到合适的菜 😢</p>
            <p class="empty-hint">换个关键词试试,或者去添加新菜品</p>
          </div>

          <div v-else class="dish-grid">
            <div
              v-for="d in filteredDishes"
              :key="d.id"
              class="dish-tile"
              :class="{ selected: (pickedDishIds ?? []).includes(d.id), used: weekUsedDishIds.has(d.id) }"
            >
              <button
                class="dish-tile-main"
                type="button"
                :aria-label="`选 ${d.name}`"
                @click="pick(d)"
              >
                <div class="dish-tile-row1">
                  <img
                    v-if="d.image"
                    :src="d.image"
                    :alt="d.name"
                    class="dish-tile-thumb"
                    loading="lazy"
                    referrerpolicy="no-referrer"
                  />
                  <span class="dish-tile-name">{{ d.name }}</span>
                </div>
                <div class="dish-tile-row2">
                  <span
                    v-if="(pickedDishIds ?? []).includes(d.id)"
                    class="badge selected-badge"
                  >✓ 已选</span>
                  <span
                    v-else-if="weekUsedDishIds.has(d.id)"
                    class="badge used-badge"
                  >本周已用</span>
                  <span
                    v-for="s in d.slots"
                    :key="s"
                    class="slot-tag"
                  >{{ SLOT_LABELS[s] }}</span>
                </div>
                <p v-if="d.note" class="dish-tile-note">{{ d.note }}</p>
              </button>
              <button
                class="dish-tile-info"
                type="button"
                :aria-label="`查看 ${d.name} 详情`"
                @click.stop="openDetail(d)"
              >📖</button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <DishDetailDialog
      :open="detailTarget !== null"
      :dish="detailTarget"
      :show-pick-button="true"
      @close="closeDetail"
      @pick="onDetailPick"
    />
  </Teleport>
</template>

<style scoped>
.sheet-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(58, 46, 54, 0.45);
  backdrop-filter: blur(2px);
  z-index: 200;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  animation: fadeIn 0.2s ease;
}

.sheet {
  width: 100%;
  max-width: 640px;
  max-height: min(85vh, 720px);
  background: var(--color-warm-bg);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  padding: 8px 16px max(20px, env(safe-area-inset-bottom)) 16px;
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  animation: slideUp 0.32s var(--ease-spring);
  overflow: hidden;
}

.handle {
  width: 40px;
  height: 4px;
  border-radius: 2px;
  background: var(--color-line-2);
  margin: 6px auto 12px;
  flex: 0 0 auto;
}

.sheet-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex: 0 0 auto;
  margin-bottom: 12px;
}
.sheet-header h3 {
  font-size: 17px;
  font-weight: 700;
  font-family: var(--font-display);
}

.search-row {
  flex: 0 0 auto;
  margin-bottom: 10px;
}
.search-input {
  width: 100%;
  font-size: 15px;
  background: var(--color-cream);
}

.slot-filter-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
  flex: 0 0 auto;
}
.slot-filter-label {
  font-size: 12px;
  color: var(--color-muted);
  font-weight: 600;
  margin-right: 2px;
}
.slot-chip {
  padding: 4px 12px;
  border-radius: var(--radius-pill);
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-stack);
  min-height: 28px;
  border: 1.5px solid transparent;
  transition: background 0.15s ease, color 0.15s ease, transform 0.15s var(--ease-spring);
}
.slot-chip:active { transform: scale(0.95); }
.slot-chip.on {
  background: var(--color-pink-400);
  color: #fff;
  border-color: var(--color-pink-500);
}
.slot-chip-clear {
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--color-muted);
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-stack);
  min-height: 28px;
  border: 1px dashed var(--color-line-2);
  transition: color 0.15s ease, border-color 0.15s ease;
}
.slot-chip-clear:hover { color: var(--color-danger); border-color: var(--color-danger); }

.empty-state {
  padding: 40px 16px;
  text-align: center;
  color: var(--color-muted);
}
.empty-state p { margin-bottom: 6px; }
.empty-hint { font-size: 13px; }

.dish-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  overflow-y: auto;
  padding: 4px 2px 8px;
  -webkit-overflow-scrolling: touch;
}
@media (min-width: 600px) {
  .dish-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.dish-tile {
  position: relative;
  display: flex;
  flex-direction: row;
  align-items: stretch;
  text-align: left;
  background: var(--color-cream);
  border: 1.5px solid var(--color-line);
  border-radius: var(--radius-md);
  min-height: 92px;
  overflow: hidden;
  transition: border-color 0.18s ease, background 0.18s ease, transform 0.18s var(--ease-spring);
  font-family: inherit;
}
.dish-tile::before {
  content: "";
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: var(--color-line-2);
  transition: background 0.18s ease;
}
.dish-tile.selected {
  border-color: var(--color-pink-400);
  background: var(--color-pink-50);
  box-shadow: var(--shadow-sm);
}
.dish-tile.selected::before { background: var(--color-pink-500); }
.dish-tile.used {
  opacity: 0.85;
}
.dish-tile.used::before { background: var(--color-butter-300); }

.dish-tile-main {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  text-align: left;
  background: transparent;
  border: none;
  padding: 12px 14px;
  cursor: pointer;
  flex: 1 1 auto;
  min-width: 0;
  color: inherit;
  font: inherit;
  border-radius: 0;
  transition: background 0.15s ease;
}
.dish-tile:hover .dish-tile-main {
  background: rgba(248, 165, 194, 0.08);
}
.dish-tile-main:active { transform: scale(0.985); }
.dish-tile-main:focus-visible {
  outline: 2px solid var(--color-pink-400);
  outline-offset: -2px;
}

.dish-tile-row1 {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  margin-bottom: 6px;
}
.dish-tile-thumb {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  object-fit: cover;
  flex: 0 0 auto;
  background: var(--color-warm-bg);
  box-shadow: var(--shadow-sm);
}
.dish-tile-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-ink);
  font-family: var(--font-display);
  word-break: break-word;
  overflow-wrap: anywhere;
  flex: 1 1 auto;
  min-width: 0;
  line-height: 1.3;
}

.dish-tile-row2 {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 6px;
  min-width: 0;
  align-items: center;
}

.dish-tile-note {
  font-size: 12px;
  color: var(--color-muted);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  margin: 0;
  padding-left: 46px;  /* 对齐 thumb 右边缘,看起来整齐 */
}

.dish-tile-info {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--color-warm-bg);
  border: 1px solid var(--color-line);
  font-size: 14px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--color-muted);
  transition: background 0.15s ease, color 0.15s ease, transform 0.15s var(--ease-spring);
}
.dish-tile-info:hover {
  background: var(--color-pink-50);
  color: var(--color-pink-500);
}
.dish-tile-info:active { transform: scale(0.9); }
.dish-tile-info:focus-visible {
  outline: 2px solid var(--color-pink-400);
  outline-offset: 2px;
}

.badge {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  flex: 0 0 auto;
  font-family: var(--font-stack);
  letter-spacing: 0.02em;
}
.selected-badge {
  background: var(--color-pink-400);
  color: #fff;
}
.used-badge {
  background: var(--color-butter-200);
  color: var(--color-on-butter);
}

.slot-tag {
  font-size: 10px;
  font-weight: 500;
  padding: 1px 7px;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--color-muted);
  border: 1px solid var(--color-line-2);
  font-family: var(--font-stack);
  line-height: 1.5;
  flex: 0 0 auto;
}
.dish-tile.selected .slot-tag {
  border-color: var(--color-pink-200);
  color: var(--color-pink-500);
}

.sheet-enter-active, .sheet-leave-active {
  transition: opacity 0.25s ease;
}
.sheet-enter-active .sheet, .sheet-leave-active .sheet {
  transition: transform 0.32s var(--ease-spring);
}
.sheet-enter-from, .sheet-leave-to {
  opacity: 0;
}
.sheet-enter-from .sheet, .sheet-leave-to .sheet {
  transform: translateY(100%);
}
</style>
