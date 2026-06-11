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
const searchInputRef = ref<HTMLInputElement | null>(null)

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

// filteredDishes — 文本搜索同时匹配 name 和 ingredients
// 用户搜"鸡蛋"既能找到《番茄炒蛋》(命中 name)
// 也能找到《蛋花汤》(命中 ingredients) — 只要任一字段命中即保留
const filteredDishes = computed<Dish[]>(() => {
  const q = query.value.trim().toLowerCase()
  const slot = props.slot

  return dishesStore.dishes.filter((d) => {
    if (!d.slots.includes(slot)) return false
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

// 打开时:重置 query + 自动聚焦搜索框
watch(() => props.open, async (isOpen) => {
  if (isOpen) {
    query.value = ''
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
              <div
                class="dish-tile-main"
                role="button"
                tabindex="0"
                :aria-label="`选 ${d.name}`"
                @click="pick(d)"
                @keydown.enter.prevent="pick(d)"
                @keydown.space.prevent="pick(d)"
              >
                <div class="dish-tile-head">
                  <img
                    v-if="d.image"
                    :src="d.image"
                    :alt="d.name"
                    class="dish-tile-thumb"
                    loading="lazy"
                    referrerpolicy="no-referrer"
                  />
                  <span class="dish-tile-name">{{ d.name }}</span>
                  <span v-if="(pickedDishIds ?? []).includes(d.id)" class="badge selected-badge">已选</span>
                  <span v-else-if="weekUsedDishIds.has(d.id)" class="badge used-badge">已用</span>
                </div>
                <p v-if="d.note" class="dish-tile-note">{{ d.note }}</p>
              </div>
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
  margin-bottom: 12px;
}
.search-input {
  width: 100%;
  font-size: 15px;
  background: var(--color-cream);
}

.empty-state {
  padding: 40px 16px;
  text-align: center;
  color: var(--color-muted);
}
.empty-state p { margin-bottom: 6px; }
.empty-hint { font-size: 13px; }

.dish-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  overflow-y: auto;
  padding: 4px 2px 8px;
  -webkit-overflow-scrolling: touch;
}
@media (min-width: 600px) {
  .dish-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.dish-tile {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  text-align: left;
  background: var(--color-cream);
  border: 1.5px solid var(--color-line);
  border-radius: var(--radius-md);
  min-height: 88px;
  transition: border-color 0.15s ease, background 0.15s ease;
  font-family: inherit;
}
.dish-tile.selected {
  border-color: var(--color-pink-400);
  background: var(--color-pink-50);
  box-shadow: var(--shadow-sm);
}
.dish-tile.used {
  opacity: 0.78;
}

.dish-tile-main {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
  background: transparent;
  border: none;
  padding: 10px 12px;
  cursor: pointer;
  flex: 1 1 auto;
  min-width: 0;
  color: inherit;
  font: inherit;
  border-radius: var(--radius-md);
  transition: transform 0.15s var(--ease-spring);
}
.dish-tile:hover .dish-tile-main {
  background: var(--color-pink-50);
}
.dish-tile-main:active { transform: scale(0.97); }
.dish-tile-main:focus-visible {
  outline: 2px solid var(--color-pink-400);
  outline-offset: -2px;
}

.dish-tile-head {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  margin-bottom: 6px;
  padding-right: 36px;
  min-width: 0;
}
.dish-tile-thumb {
  width: 30px;
  height: 30px;
  border-radius: 6px;
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
  flex: 1 1 auto;
  min-width: 0;
}

.dish-tile-info {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
  width: 28px;
  height: 28px;
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
  font-weight: 600;
  padding: 2px 6px;
  border-radius: var(--radius-pill);
  flex: 0 0 auto;
  font-family: var(--font-stack);
}
.selected-badge {
  background: var(--color-pink-400);
  color: #fff;
}
.used-badge {
  background: var(--color-butter-200);
  color: #6b5a1f;
}

.dish-tile-note {
  font-size: 11px;
  color: var(--color-muted);
  margin-top: 2px;
  line-height: 1.35;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
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
