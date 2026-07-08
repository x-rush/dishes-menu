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

const subText = computed(() => {
  if (!loaded.value) return ''
  if (!state.value.together_since) return '点我设置'
  return '在一起'
})

function onClick() {
  if (!state.value.together_since) {
    dialogRef.value?.showModal()
    return
  }
  surpriseVisible.value = true
  currentIdx.value = 0
  if (phraseTimer !== null) {
    window.clearInterval(phraseTimer)
    phraseTimer = null
  }
  // 5 句轮播,每句 700ms,总 3.5s 后收起
  phraseTimer = window.setInterval(() => {
    currentIdx.value = (currentIdx.value + 1) % phrases.length
  }, 700)
  window.setTimeout(() => {
    surpriseVisible.value = false
    if (phraseTimer !== null) {
      window.clearInterval(phraseTimer)
      phraseTimer = null
    }
  }, phrases.length * 700)
}

// 5 句轮播的深情话 — 按从含蓄到外放排列
const phrases = [
  '💗 今天的你,也想见 💗',
  '🌸 想你想到心都软了',
  '☀️ 想到你就会笑',
  '🍚 一起吃饭的日子最暖',
  '✨ 永远最爱你',
]
const currentIdx = ref(0)
let phraseTimer: number | null = null

async function save() {
  await setDate(inputDate.value)
  dialogRef.value?.close()
}
</script>

<template>
  <button class="heart-tab" :class="{ pulse: surpriseVisible }" @click="onClick" aria-label="在一起的天数">
    <span class="heart-disc">
      <span class="heart-num">{{ display }}</span>
    </span>
    <span class="heart-sub">{{ subText }}</span>
  </button>

  <!-- Teleport to body 避开 TabBar 的 backdrop-filter 形成的 containing block,
       否则 position:fixed 会以 TabBar 自身为参照居中,导致文案落在底部 bar 区域 -->
  <Teleport to="body">
    <Transition name="surprise">
      <div v-if="surpriseVisible" class="surprise-overlay" aria-hidden="true">
        <div v-for="i in 14" :key="i" :class="['sparkle', `s${i}`]">✨</div>
        <Transition name="phrase" mode="out-in">
          <p :key="currentIdx" class="surprise-msg">{{ phrases[currentIdx] }}</p>
        </Transition>
        <div class="surprise-dots" aria-hidden="true">
          <span
            v-for="(_, i) in phrases"
            :key="i"
            :class="['dot', { active: i === currentIdx }]"
          />
        </div>
      </div>
    </Transition>
  </Teleport>

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
.heart-tab {
  background: transparent;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  width: 100%;
  padding: 0;
  margin-top: -10px;
  position: relative;
  z-index: 1;
}
.heart-disc {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ffb0ce 0%, #ec7da6 60%, #d65a8a 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 6px 18px rgba(236, 125, 166, 0.4),
    inset 0 -2px 6px rgba(214, 90, 138, 0.4),
    inset 0 2px 4px rgba(255, 255, 255, 0.3);
  border: 3px solid #fff8f3;
  animation: heart-bob 2.6s ease-in-out infinite;
  transition: transform 0.2s var(--ease-spring);
}
:root[data-theme="dark"] .heart-disc {
  border-color: #2d2226;
}
.heart-tab:active .heart-disc { transform: scale(0.92); }

.heart-num {
  font-size: 19px;
  font-weight: 700;
  font-family: var(--font-stack);
  letter-spacing: -0.02em;
  line-height: 1;
  text-shadow: 0 1px 2px rgba(214, 90, 138, 0.4);
}

.heart-sub {
  font-size: 10px;
  color: var(--color-pink-500);
  font-weight: 600;
  letter-spacing: 0.04em;
  white-space: nowrap;
  line-height: 1;
  margin-top: 4px;
}

.heart-tab.pulse .heart-disc {
  animation: heart-pop 0.6s var(--ease-spring);
}

@keyframes heart-bob {
  0%, 100% { transform: translateY(0) scale(1); }
  50%      { transform: translateY(-3px) scale(1.04); }
}
@keyframes heart-pop {
  0%   { transform: scale(1); }
  40%  { transform: scale(1.22); }
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
  font-size: 26px;
  animation: sparkle-fly 1.6s ease-out forwards;
}
.s1  { left: 50%; top: 50%; }
.s2  { left: 22%; top: 38%; }
.s3  { left: 78%; top: 32%; }
.s4  { left: 30%; top: 70%; }
.s5  { left: 70%; top: 72%; }
.s6  { left: 18%; top: 52%; }
.s7  { left: 82%; top: 52%; }
.s8  { left: 45%; top: 25%; }
.s9  { left: 55%; top: 78%; }
.s10 { left: 12%; top: 30%; }
.s11 { left: 88%; top: 30%; }
.s12 { left: 38%; top: 80%; }
.s13 { left: 62%; top: 80%; }
.s14 { left: 50%; top: 20%; }
@keyframes sparkle-fly {
  0%   { transform: translate(0, 0) scale(0.4); opacity: 0; }
  30%  { opacity: 1; }
  100% { transform: translate(var(--tx, 40px), var(--ty, -60px)) scale(1.2); opacity: 0; }
}
.s1  { --tx:   0px; --ty:  -90px; }
.s2  { --tx: -80px; --ty:  -40px; }
.s3  { --tx:  80px; --ty:  -50px; }
.s4  { --tx: -60px; --ty:  60px; }
.s5  { --tx:  60px; --ty:  60px; }
.s6  { --tx:-100px; --ty:   0px; }
.s7  { --tx: 100px; --ty:   0px; }
.s8  { --tx: -30px; --ty: -100px; }
.s9  { --tx:  30px; --ty:  90px; }
.s10 { --tx:-120px; --ty: -30px; }
.s11 { --tx: 120px; --ty: -30px; }
.s12 { --tx: -50px; --ty: 110px; }
.s13 { --tx:  50px; --ty: 110px; }
.s14 { --tx:   0px; --ty:-130px; }

.surprise-msg {
  /* overlay 已是 flex 居中,这里只控制自身排版,不要 absolute 拉走 */
  font-size: 26px;
  font-weight: 700;
  color: var(--color-pink-600);
  font-family: var(--font-display);
  text-shadow: 0 2px 12px rgba(255, 240, 245, 0.9), 0 0 2px #fff;
  letter-spacing: 0.02em;
  white-space: nowrap;
  text-align: center;
  padding: 0 24px;
  max-width: 90vw;
}

.surprise-dots {
  position: absolute;
  bottom: max(20px, env(safe-area-inset-bottom));
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 6px;
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-pink-200);
  transition: background 0.2s ease, transform 0.2s ease;
}
.dot.active {
  background: var(--color-pink-500);
  transform: scale(1.4);
}

.phrase-enter-active, .phrase-leave-active {
  transition: opacity 0.22s ease, transform 0.22s var(--ease-spring);
}
.phrase-enter-from {
  opacity: 0;
  transform: translateY(8px) scale(0.92);
}
.phrase-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.92);
}

.surprise-enter-active, .surprise-leave-active { transition: opacity 0.2s ease; }
.surprise-enter-from, .surprise-leave-to { opacity: 0; }

.dlg-body { padding: 20px; min-width: min(320px, 90vw); }
.dlg-body h3 { margin: 0 0 8px; font-size: 17px; }
.hint { font-size: 13px; color: var(--color-muted); margin: 0 0 12px; }
.dlg-body input[type=date] { width: 100%; font-size: 15px; }
.dlg-footer { display: flex; gap: 8px; justify-content: flex-end; margin-top: 14px; }
</style>
