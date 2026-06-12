import { createRouter, createWebHistory } from 'vue-router'
import Home from '../pages/Home.vue'
import { isoWeekKey, dayKeyForDate, dateToDayKey } from '../utils/isoWeek'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: () => `/${isoWeekKey()}/${todayDateString()}` },
    {
      // /:week/:date 双参路由,date 是 YYYY-MM-DD 格式(从 Monday +0..6 派生 day)
      path: '/:week(202\\d-W\\d{2})/:date(\\d{4}-\\d{2}-\\d{2})',
      component: Home,
    },
    {
      // /:week — 缺 date,跳到本周今天
      path: '/:week(202\\d-W\\d{2})',
      redirect: (to) => {
        const week = to.params.week as string
        // 用今天作为 day(无论周几),week 已经合法
        return `/${week}/${todayDateString()}`
      },
    },
    { path: '/:week', redirect: () => `/${isoWeekKey()}/${todayDateString()}` },
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
