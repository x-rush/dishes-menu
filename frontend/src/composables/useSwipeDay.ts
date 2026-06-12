import type { Ref } from 'vue'
import { useSwipe } from '@vueuse/core'

export interface UseSwipeDayOptions {
  /** 排除容器:内部滑动不触发切 day(默认 '.week-tabs',否则 tab 横滑会同时切 day) */
  excludeSelector?: string
  /** 触发切 day 的最小水平距离 px(默认 60) */
  threshold?: number
  /** 左滑触发(下一天) */
  onNext?: () => void
  /** 右滑触发(上一天) */
  onPrev?: () => void
}

/**
 * 横向滑动切 day:左滑 = 下一天,右滑 = 上一天。
 * 绑定到一个容器 ref(比如 .home),容器内任意位置起手都生效。
 * excludeSelector 内的元素(默认 .week-tabs)滑动不触发。
 *
 * 用回调 API 而不是 emit,因为 composable 不是组件,不能 defineEmits。
 * caller(Home.vue)负责把 day 变化同步到 URL / store。
 */
export function useSwipeDay(
  target: Ref<HTMLElement | null>,
  options: UseSwipeDayOptions = {}
) {
  const exclude = options.excludeSelector ?? '.week-tabs'
  const threshold = options.threshold ?? 60

  useSwipe(target, {
    threshold,
    onSwipeEnd(e, direction) {
      if (direction !== 'left' && direction !== 'right') return
      const t = e.target as HTMLElement | null
      if (t?.closest(exclude)) return
      if (direction === 'left') options.onNext?.()
      else options.onPrev?.()
    },
  })
}
