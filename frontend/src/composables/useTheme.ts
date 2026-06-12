import { useDark, useToggle } from '@vueuse/core'

/**
 * 暗色模式开关。包装 @vueuse/core 的 useDark,把状态写到
 * `<html data-theme="dark">` 上,然后 CSS 用 `[data-theme="dark"]` 选择器应用 token。
 * 第一次访问跟随系统,用户手动切换后写 localStorage 持久化。
 */
export function useTheme() {
  const isDark = useDark({
    selector: 'html',
    attribute: 'data-theme',
    valueDark: 'dark',
    valueLight: 'light',
    storageKey: 'dishes-menu-theme',
  })
  const toggle = useToggle(isDark)
  return { isDark, toggle }
}
