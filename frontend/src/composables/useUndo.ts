import { ref, computed } from 'vue'

export interface UndoTask {
  id: number
  message: string
  undo: () => void | Promise<void>
  expiresAt: number
}

const tasks = ref<UndoTask[]>([])
let nextId = 0
let ticker: ReturnType<typeof setInterval> | null = null

function ensureTicker() {
  if (ticker !== null) return
  ticker = setInterval(() => {
    const now = Date.now()
    tasks.value = tasks.value.filter((t) => t.expiresAt > now)
    if (tasks.value.length === 0 && ticker !== null) {
      clearInterval(ticker)
      ticker = null
    }
  }, 250)
}

export function useUndo() {
  const currentTask = computed<UndoTask | null>(() => tasks.value[0] ?? null)

  /**
   * 推入一个撤销任务。**单槽**语义:同一时刻只显示一个 toast,
   * 新任务会替换旧任务(避免堆叠)。
   * 5 秒后自动过期,过期后 undo 不会再被调用。
   */
  function push(message: string, undo: () => void | Promise<void>, durationMs = 5000) {
    tasks.value = [
      {
        id: ++nextId,
        message,
        undo,
        expiresAt: Date.now() + durationMs,
      },
    ]
    ensureTicker()
  }

  async function performUndo() {
    const t = currentTask.value
    if (!t) return
    tasks.value = tasks.value.filter((x) => x.id !== t.id)
    await t.undo()
  }

  function dismiss() {
    tasks.value = []
  }

  return { currentTask, push, performUndo, dismiss }
}
