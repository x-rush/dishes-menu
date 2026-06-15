import { ref } from 'vue'
import { api } from '../api/client'
import type { Together } from '../types'

// Module-level state shared across all components using this composable,
// so the heart FAB on every page reads the same value without re-fetching.
const state = ref<Together>({ together_since: null, days: 0 })
const loaded = ref(false)

/**
 * 共享在一起纪念日。模块级 state 让所有组件 (心形 FAB、设置页等) 看到同一份数据。
 * 失败时静默 (心形保持默认 0),不向用户报错 — 这个数据是装饰性的,不是阻塞性的。
 */
export function useTogether() {
  async function refresh(): Promise<void> {
    try {
      state.value = await api.getTogether()
    } catch {
      // 静默失败,心形显示默认 0
    } finally {
      loaded.value = true
    }
  }

  async function setDate(date: string): Promise<void> {
    await api.setTogether(date)
    await refresh()
  }

  return { state, loaded, refresh, setDate }
}
