<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useDishesStore } from '../stores/dishes'
import { ALL_SLOTS, SLOT_LABELS, type Slot } from '../types'

const dishes = useDishesStore()

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [], saved: [] }>()

const dialogRef = ref<HTMLDialogElement | null>(null)

interface DishForm {
  name: string
  slots: Slot[]
  estimatedTime: number
  ingredients: string[]
  note: string
  image: string
}

const MAX_INGREDIENTS = 20
const MAX_INGREDIENT_LEN = 16

const form = reactive<DishForm>({
  name: '',
  slots: [],
  estimatedTime: 30,
  ingredients: [],
  note: '',
  image: '',
})

const ingredientInput = ref('')
const error = ref<string | null>(null)
const submitting = ref(false)

function reset() {
  form.name = ''
  form.slots = []
  form.estimatedTime = 30
  form.ingredients = []
  form.note = ''
  form.image = ''
  ingredientInput.value = ''
  error.value = null
}

function toggleSlot(s: Slot) {
  if (form.slots.includes(s)) {
    form.slots = form.slots.filter((x) => x !== s)
  } else {
    form.slots = [...form.slots, s]
  }
}

// 添加食材:trim → 跳过空/重复 → 上限静默忽略 → 清空输入框
function addIngredient() {
  const v = ingredientInput.value.trim().slice(0, MAX_INGREDIENT_LEN)
  if (!v) {
    ingredientInput.value = ''
    return
  }
  if (form.ingredients.length >= MAX_INGREDIENTS) {
    ingredientInput.value = ''
    return
  }
  if (form.ingredients.includes(v)) {
    ingredientInput.value = ''
    return
  }
  form.ingredients = [...form.ingredients, v]
  ingredientInput.value = ''
}

function removeIngredient(i: number) {
  form.ingredients = form.ingredients.filter((_, idx) => idx !== i)
}

function close() {
  dialogRef.value?.close()
}

function onClose() {
  reset()
  emit('close')
}

async function submit() {
  if (submitting.value) return
  if (!form.name.trim()) {
    error.value = '请填写菜名'
    return
  }
  if (form.slots.length === 0) {
    error.value = '请至少选择一个时段'
    return
  }
  if (form.estimatedTime < 0 || form.estimatedTime > 999) {
    error.value = '预计用时需在 0~999 分钟之间'
    return
  }
  submitting.value = true
  error.value = null
  try {
    await dishes.create({
      name: form.name.trim(),
      slots: form.slots,
      ingredients: form.ingredients.filter((x) => x.trim()),
      estimated_time: form.estimatedTime,
      note: form.note.trim(),
      image: form.image.trim() || undefined,
    })
    emit('saved')
    close()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '保存失败'
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      reset()
      dialogRef.value?.showModal()
    } else {
      dialogRef.value?.close()
    }
  },
  { immediate: true }
)
</script>

<template>
  <dialog ref="dialogRef" @close="onClose">
    <div class="dlg-body">
      <header class="dlg-header">
        <h3>添加菜品</h3>
        <button class="btn btn-ghost" @click="close">取消</button>
      </header>

      <form @submit.prevent="submit" class="form">
        <label class="field">
          <span class="lbl">菜名 *</span>
          <input
            v-model="form.name"
            type="text"
            placeholder="例：番茄牛腩"
            maxlength="32"
            required
          />
        </label>

        <label class="field">
          <span class="lbl">图片链接（可选）</span>
          <input
            v-model="form.image"
            type="url"
            placeholder="https://example.com/food.jpg"
            maxlength="500"
          />
          <p class="hint">粘贴任意图片 URL,留空则在详情页不显示</p>
        </label>

        <fieldset class="field">
          <legend class="lbl">适用时段 *</legend>
          <div class="opt-row">
            <button
              v-for="s in ALL_SLOTS"
              :key="s"
              type="button"
              :class="['opt-btn', { on: form.slots.includes(s) }]"
              @click="toggleSlot(s)"
            >{{ SLOT_LABELS[s] }}</button>
          </div>
        </fieldset>

        <label class="field">
          <span class="lbl">预计用时（分钟）</span>
          <input
            v-model.number="form.estimatedTime"
            type="number"
            min="0"
            max="999"
            step="5"
            placeholder="30"
          />
          <p class="hint">大致做完这道菜要多久,留默认 30 也行</p>
        </label>

        <fieldset class="field">
          <legend class="lbl">食材（可选,Enter 或 + 添加）</legend>
          <div class="chip-input">
            <input
              v-model="ingredientInput"
              type="text"
              :maxlength="MAX_INGREDIENT_LEN"
              placeholder="例：番茄"
              :disabled="form.ingredients.length >= MAX_INGREDIENTS"
              @keydown.enter.prevent="addIngredient"
            />
            <button
              type="button"
              class="btn btn-primary chip-add-btn"
              :disabled="!ingredientInput.trim() || form.ingredients.length >= MAX_INGREDIENTS"
              @click="addIngredient"
            >+</button>
          </div>
          <div v-if="form.ingredients.length" class="chip-list">
            <span
              v-for="(ing, i) in form.ingredients"
              :key="ing + i"
              class="chip"
            >
              {{ ing }}
              <button
                type="button"
                class="chip-remove"
                :aria-label="`删除 ${ing}`"
                @click="removeIngredient(i)"
              >×</button>
            </span>
          </div>
          <p v-if="form.ingredients.length >= MAX_INGREDIENTS" class="hint">最多 {{ MAX_INGREDIENTS }} 个食材</p>
        </fieldset>

        <label class="field">
          <span class="lbl">做法 / 小贴士（可选）</span>
          <textarea
            v-model="form.note"
            rows="4"
            placeholder="比如：番茄炒出汁再下牛腩,小火慢炖 40 分钟…"
            maxlength="500"
          ></textarea>
        </label>

        <p v-if="error" class="error">{{ error }}</p>

        <div class="dlg-footer">
          <button type="button" class="btn btn-ghost" @click="close">取消</button>
          <button type="submit" class="btn btn-primary" :disabled="submitting">
            <span v-if="submitting" class="spinner" aria-hidden="true"></span>
            <span>保存到菜品库</span>
          </button>
        </div>
      </form>
    </div>
  </dialog>
</template>

<style scoped>
.dlg-body { padding: 20px; }
.dlg-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.dlg-header h3 { font-size: 19px; font-family: var(--font-display); }
.dlg-header .btn { padding: 6px 14px; font-size: 13px; }

.form { display: flex; flex-direction: column; gap: 16px; }

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border: none;
  padding: 0;
  margin: 0;
}
.lbl {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-ink);
}

input[type="text"], input[type="url"], input[type="number"], textarea {
  width: 100%;
  font-size: 15px;
}
textarea { resize: vertical; }
.hint {
  font-size: 12px;
  color: var(--color-muted);
  margin-top: 2px;
}

.opt-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.opt-btn {
  padding: 8px 16px;
  border-radius: var(--radius-pill);
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  font-size: 13px;
  font-weight: 600;
  font-family: var(--font-stack);
  min-height: 36px;
  transition: background 0.15s ease, color 0.15s ease, transform 0.15s var(--ease-spring);
}
.opt-btn:active { transform: scale(0.96); }
.opt-btn.on {
  background: var(--color-pink-400);
  color: #fff;
}

.chip-input {
  display: flex;
  gap: 6px;
  align-items: stretch;
}
.chip-input input {
  flex: 1 1 auto;
  min-width: 0;
}
.chip-add-btn {
  flex: 0 0 auto;
  padding: 0 16px;
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
  min-width: 44px;
  min-height: 44px;
}
.chip-add-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 4px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 6px 5px 12px;
  background: var(--color-pink-50);
  color: var(--color-pink-500);
  border-radius: var(--radius-pill);
  font-size: 13px;
  font-weight: 600;
  animation: fadeIn 0.18s ease;
}
.chip-remove {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.6);
  color: var(--color-pink-500);
  font-size: 14px;
  line-height: 1;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-stack);
  border: none;
  transition: background 0.15s ease, transform 0.15s var(--ease-spring);
}
.chip-remove:hover {
  background: var(--color-pink-400);
  color: #fff;
}
.chip-remove:active { transform: scale(0.85); }

.error {
  color: var(--color-danger);
  font-size: 13px;
  background: rgba(226, 109, 109, 0.08);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
}

.dlg-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>
