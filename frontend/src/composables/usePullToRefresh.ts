// 移动端下拉刷新手势
// 行为:
//   - 仅在 scrollTop=0 时响应(否则浏览器的原生滚动应该接管)
//   - 向下拉 70px+ 触发 onRefresh,松手时距离回弹到 0
//   - 阻力系数 0.4,模仿 iOS 的橡皮筋感(下拉 100px 只移动 40px)
//   - 与现有横向 useSwipeDay 手势无冲突(方向不同)
import { ref, onMounted, onUnmounted, watch, type Ref } from 'vue'

export interface PullToRefreshOptions {
  /** 监听的滚动容器,通常是 Home.vue 的根元素 */
  target: Ref<HTMLElement | null>
  /** 触发刷新时的回调(可异步) */
  onRefresh: () => void | Promise<void>
  /** 触发距离阈值(原始下拉 px),默认 70 */
  threshold?: number
  /** 阻力系数,默认 0.4 */
  resistance?: number
}

export function usePullToRefresh(options: PullToRefreshOptions) {
  const pulling = ref(false)
  const pullDistance = ref(0)
  const refreshing = ref(false)

  const threshold = options.threshold ?? 70
  const resistance = options.resistance ?? 0.4

  let startY = 0
  let isDragging = false
  let activeEl: HTMLElement | null = null

  function onTouchStart(e: TouchEvent) {
    if (refreshing.value) return
    const el = options.target.value
    if (!el || el.scrollTop > 0) return
    const t = e.touches[0]
    if (!t) return
    startY = t.clientY
    isDragging = true
    pulling.value = true
    pullDistance.value = 0
  }

  function onTouchMove(e: TouchEvent) {
    if (!isDragging) return
    const t = e.touches[0]
    if (!t) return
    const deltaY = t.clientY - startY
    if (deltaY <= 0) {
      // 向上滑:取消手势
      pullDistance.value = 0
      return
    }
    pullDistance.value = deltaY * resistance
  }

  async function onTouchEnd() {
    if (!isDragging) return
    isDragging = false

    if (pullDistance.value >= threshold) {
      // 触发刷新:把指示器固定在 threshold 位置,等回调完成再回弹
      refreshing.value = true
      pullDistance.value = threshold
      try {
        await options.onRefresh()
      } finally {
        refreshing.value = false
        pulling.value = false
        pullDistance.value = 0
      }
    } else {
      // 未达阈值:直接回弹
      pulling.value = false
      pullDistance.value = 0
    }
  }

  function bind() {
    const el = options.target.value
    if (!el || el === activeEl) return
    if (activeEl) unbind()
    activeEl = el
    el.addEventListener('touchstart', onTouchStart, { passive: true })
    el.addEventListener('touchmove', onTouchMove, { passive: true })
    el.addEventListener('touchend', onTouchEnd, { passive: true })
    el.addEventListener('touchcancel', onTouchEnd, { passive: true })
  }

  function unbind() {
    if (!activeEl) return
    activeEl.removeEventListener('touchstart', onTouchStart)
    activeEl.removeEventListener('touchmove', onTouchMove)
    activeEl.removeEventListener('touchend', onTouchEnd)
    activeEl.removeEventListener('touchcancel', onTouchEnd)
    activeEl = null
  }

  onMounted(() => bind())
  watch(options.target, () => bind())
  onUnmounted(() => unbind())

  return { pulling, pullDistance, refreshing, threshold }
}
