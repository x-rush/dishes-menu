import { createRouter, createWebHistory } from 'vue-router'
import Home from '../pages/Home.vue'
import { isoWeekKey, dayKeyForDate, dateToDayKey } from '../utils/isoWeek'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // / → /menu/<current week>/<today>
    { path: '/', redirect: () => `/menu/${isoWeekKey()}/${todayDateString()}` },
    // 老的 /:week/:date → 重定向到 /menu/:week/:date(保留旧书签)
    {
      path: '/:week(202\\d-W\\d{2})/:date(\\d{4}-\\d{2}-\\d{2})',
      redirect: (to) => `/menu${to.fullPath}`,
    },
    // /menu/:week/:date → Home
    {
      path: '/menu/:week(202\\d-W\\d{2})/:date(\\d{4}-\\d{2}-\\d{2})',
      component: Home,
    },
    // /menu/:week — 缺 date,跳到本周今天(命名空间下)
    {
      path: '/menu/:week(202\\d-W\\d{2})',
      redirect: () => `/menu/${isoWeekKey()}/${todayDateString()}`,
    },
    // /todo → TodoPage(lazy)。D1 任务会创建 pages/TodoPage.vue
    {
      path: '/todo',
      component: () => import('../pages/TodoPage.vue'),
    },
    // 老的 /:week(约束) — 跳到 /menu/:week/today
    {
      path: '/:week(202\\d-W\\d{2})',
      redirect: (to) => {
        const week = to.params.week as string
        return `/menu/${week}/${todayDateString()}`
      },
    },
    // 老的 /:week(任意值) — 跳到 /menu/本周/today
    { path: '/:week', redirect: () => `/menu/${isoWeekKey()}/${todayDateString()}` },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

/** 今天 YYYY-MM-DD(本地时区) */
function todayDateString(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export default router
