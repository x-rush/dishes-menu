<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTogether } from '../composables/useTogether'
import { formatDateString } from '../utils/isoWeek'

const { state, loaded, refresh, setDate } = useTogether()
const dialogRef = ref<HTMLDialogElement | null>(null)
const surpriseVisible = ref(false)
const inputDate = ref<string>(
  state.value.together_since ?? formatDateString(new Date())
)

onMounted(() => {
  refresh()
})

const display = computed(() => {
  if (!loaded.value || !state.value.together_since) return '♡'
  return String(state.value.days)
})

const subtitle = computed(() => {
  if (!loaded.value) return '加载中…'
  if (!state.value.together_since) return '点我设置纪念日'
  return `在一起 ${state.value.days} 天`
})

function onClick() {
  if (!state.value.together_since) {
    // 首次:打开设置弹窗
    dialogRef.value?.showModal()
    return
  }
  // 已有:触发惊喜动画
  surpriseVisible.value = true
  setTimeout(() => {
    surpriseVisible.value = false
  }, 2400)
}

async function save() {
  await setDate(inputDate.value)
  dialogRef.value?.close()
}
</script>

<template>
  <button class="heart-fab" :class="{ pulse: surpriseVisible }" @click="onClick">
    <span class="heart-icon">{{ display }}</span>
    <span class="heart-sub">{{ subtitle }}</span>
  </button>

  <!-- 惊喜粒子 -->
  <Transition name="surprise">
    <div v-if="surpriseVisible" class="surprise-overlay" aria-hidden="true">
      <div v-for="i in 12" :key="i" :class="['sparkle', `s${i}`]">✨</div>
      <p class="surprise-msg">💗 想你 💗</p>
    </div>
  </Transition>

  <dialog ref="dialogRef" @click="(e) => e.target === dialogRef && dialogRef?.close()">
    <div class="dlg-body">
      <h3>设置纪念日</h3>
      <p class="hint">输入你们在一起的日子,以后每天都会算在一起的天数。</p>
      <input v-model="inputDate" type="date" />
      <div class="dlg-footer">
        <button class="btn btn-ghost" @click="dialogRef?.close()">取消</button>
        <button class="btn btn-primary" @click="save">保存</button>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.heart-fab {
  position: fixed;
  bottom: calc(72px + env(safe-area-inset-bottom));  /* 浮在 TabBar 上 */
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  width: 76px;
  background: transparent;
  z-index: 30;
  pointer-events: auto;
}
.heart-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff8fb1, #ec7da6);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 18px rgba(236, 125, 166, 0.4);
  animation: heart-bob 2.4s ease-in-out infinite;
}
.heart-sub {
  font-size: 10px;
  color: var(--color-pink-500);
  font-weight: 600;
  background: var(--color-cream);
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  box-shadow: var(--shadow-sm);
  white-space: nowrap;
}
.heart-fab.pulse .heart-icon {
  animation: heart-pop 0.6s var(--ease-spring);
}
@keyframes heart-bob {
  0%, 100% { transform: translateY(0) scale(1); }
  50%      { transform: translateY(-3px) scale(1.04); }
}
@keyframes heart-pop {
  0%   { transform: scale(1); }
  40%  { transform: scale(1.25); }
  100% { transform: scale(1); }
}

.surprise-overlay {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}
.sparkle {
  position: absolute;
  font-size: 28px;
  animation: sparkle-fly 1.6s ease-out forwards;
}
.s1  { left: 20%; top: 30%; animation-delay: 0.0s; }
.s2  { left: 30%; top: 60%; animation-delay: 0.1s; }
.s3  { left: 70%; top: 25%; animation-delay: 0.15s; }
.s4  { left: 80%; top: 55%; animation-delay: 0.05s; }
.s5  { left: 50%; top: 20%; animation-delay: 0.2s; }
.s6  { left: 25%; top: 70%; animation-delay: 0.12s; }
.s7  { left: 75%; top: 70%; animation-delay: 0.18s; }
.s8  { left: 45%; top: 80%; animation-delay: 0.08s; }
.s9  { left: 55%; top: 35%; animation-delay: 0.22s; }
.s10 { left: 15%; top: 45%; animation-delay: 0.13s; }
.s11 { left: 85%; top: 40%; animation-delay: 0.17s; }
.s12 { left: 40%; top: 50%; animation-delay: 0.07s; }
@keyframes sparkle-fly {
  0%   { transform: translate(0, 0) scale(0.5); opacity: 0; }
  30%  { opacity: 1; }
  100% { transform: translate(var(--tx, 40px), var(--ty, -60px)) scale(1.2); opacity: 0; }
}
.s1  { --tx: -30px; --ty: -50px; }
.s2  { --tx:  20px; --ty: -70px; }
.s3  { --tx: -40px; --ty: -30px; }
.s4  { --tx:  35px; --ty: -80px; }
.s5  { --tx: -10px; --ty: -90px; }
.s6  { --tx:  50px; --ty: -40px; }
.s7  { --tx: -45px; --ty: -20px; }
.s8  { --tx:  25px; --ty: -100px; }
.s9  { --tx: -25px; --ty: -60px; }
.s10 { --tx:  60px; --ty: -30px; }
.s11 { --tx: -55px; --ty: -45px; }
.s12 { --tx:  10px; --ty: -110px; }

.surprise-msg {
  position: absolute;
  top: 38%;
  font-size: 28px;
  font-weight: 700;
  color: var(--color-pink-600);
  font-family: var(--font-display);
  animation: msg-pop 2.4s ease-out forwards;
}
@keyframes msg-pop {
  0%   { transform: scale(0.6); opacity: 0; }
  20%  { transform: scale(1.1); opacity: 1; }
  80%  { transform: scale(1); opacity: 1; }
  100% { transform: scale(1.1); opacity: 0; }
}

.surprise-enter-active, .surprise-leave-active { transition: opacity 0.2s ease; }
.surprise-enter-from, .surprise-leave-to { opacity: 0; }

.dlg-body { padding: 20px; min-width: min(320px, 90vw); }
.dlg-body h3 { margin: 0 0 8px; font-size: 17px; }
.hint { font-size: 13px; color: var(--color-muted); margin: 0 0 12px; }
.dlg-body input[type=date] { width: 100%; font-size: 15px; }
.dlg-footer { display: flex; gap: 8px; justify-content: flex-end; margin-top: 14px; }
</style>
