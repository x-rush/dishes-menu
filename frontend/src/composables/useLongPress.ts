import type { Ref } from 'vue'
import { useEventListener } from '@vueuse/core'

export interface UseLongPressOptions {
  /** 达到阈值时触发(长按动作) */
  onLong: () => void
  /** 短按松开时触发(普通点击)。不传则静默 */
  onShort?: () => void
  /** 长按阈值,默认 500ms。太小易误触,太大没"手速感" */
  threshold?: number
  /** 移动容差,默认 30px。手指抖动超过此距离视为取消 */
  movement?: number
}

const DEFAULTS = {
  threshold: 500,
  movement: 30,
}

/**
 * 检测长按 + 短按手势。基于 Pointer Events,iOS 13+/Android 全版本。
 * 用法:
 *   const slotEl = ref<HTMLElement | null>(null)
 *   useLongPress(slotEl, {
 *     onLong: () => menu.shuffleDish(week, day, slot),
 *     onShort: () => menu.removeItem(...),
 *   })
 */
export function useLongPress(
  target: Ref<HTMLElement | null>,
  options: UseLongPressOptions
): void {
  const threshold = options.threshold ?? DEFAULTS.threshold
  const movement = options.movement ?? DEFAULTS.movement

  let timer: ReturnType<typeof setTimeout> | null = null
  let triggered = false
  let startX = 0
  let startY = 0
  let active = false

  function reset() {
    if (timer !== null) {
      clearTimeout(timer)
      timer = null
    }
    active = false
  }

  useEventListener(target, 'pointerdown', (e: PointerEvent) => {
    active = true
    triggered = false
    startX = e.clientX
    startY = e.clientY
    timer = setTimeout(() => {
      if (!active) return
      triggered = true
      timer = null
      try {
        navigator.vibrate?.(30)
      } catch {
        // iOS Safari 等不支持震动,静默
      }
      options.onLong()
    }, threshold)
  })

  useEventListener(target, 'pointermove', (e: PointerEvent) => {
    if (!active) return
    const dx = Math.abs(e.clientX - startX)
    const dy = Math.abs(e.clientY - startY)
    if (dx > movement || dy > movement) reset()
  })

  useEventListener(target, 'pointerup', () => {
    if (!active) return
    if (timer !== null) {
      clearTimeout(timer)
      timer = null
    }
    if (!triggered) options.onShort?.()
    active = false
  })

  useEventListener(target, 'pointercancel', reset)
  useEventListener(target, 'pointerleave', reset)
}
