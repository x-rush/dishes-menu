// ISO Week helpers — matches backend's fmt.Sprintf("%d-W%02d", y, w)

export function isoWeekKey(d: Date = new Date()): string {
  const target = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()))
  const dayNr = (target.getUTCDay() + 6) % 7
  target.setUTCDate(target.getUTCDate() - dayNr + 3)
  const firstThursday = new Date(Date.UTC(target.getUTCFullYear(), 0, 4))
  const firstDayNr = (firstThursday.getUTCDay() + 6) % 7
  firstThursday.setUTCDate(firstThursday.getUTCDate() - firstDayNr + 3)
  const week =
    1 +
    Math.round(
      (target.getTime() - firstThursday.getTime()) / (7 * 24 * 3600 * 1000)
    )
  return `${target.getUTCFullYear()}-W${String(week).padStart(2, '0')}`
}

export function dayKeyForDate(d: Date = new Date()): 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat' | 'sun' {
  const map = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'] as const
  return map[d.getDay()]
}

/** ISO 周 key → 那个周的 Monday(本地时区午夜) */
export function weekKeyToMonday(weekKey: string): Date {
  const m = /^(\d{4})-W(\d{2})$/.exec(weekKey)
  if (!m) return new Date()
  const year = Number(m[1])
  const week = Number(m[2])
  // Jan 4 总在 week 1,本地时区
  const jan4 = new Date(year, 0, 4)
  const jan4Day = jan4.getDay() // 0=Sun..6=Sat
  // 推到 week 1 的 Monday:Mon(1)→0, Tue(2)→-1, ..., Sat(6)→-5, Sun(0)→-6
  const daysFromJan4ToMon = jan4Day === 0 ? -6 : 1 - jan4Day
  const mon = new Date(jan4)
  mon.setDate(jan4.getDate() + daysFromJan4ToMon + (week - 1) * 7)
  return mon
}

/** ISO 周 key 偏移 N 天(常用于 +7/-7 切周) */
export function addDaysToWeek(weekKey: string, days: number): string {
  const mon = weekKeyToMonday(weekKey)
  mon.setDate(mon.getDate() + days)
  return isoWeekKey(mon)
}

// ─── date ↔ day 互转(URL 同步用) ──────────────────────────────────────────

/** YYYY-MM-DD 字符串 → Date(本地时区午夜) */
export function parseDateString(s: string): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s)
  if (!m) return null
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return isNaN(d.getTime()) ? null : d
}

/** Date → YYYY-MM-DD(本地时区) */
export function formatDateString(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** YYYY-MM-DD → day key('mon'..'sun') */
export function dateToDayKey(s: string): 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat' | 'sun' | null {
  const d = parseDateString(s)
  if (!d) return null
  return dayKeyForDate(d)
}

/** week + day → 该 day 对应的 YYYY-MM-DD(本地时区) */
export function dayToDate(weekKey: string, day: 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat' | 'sun'): string {
  const mon = weekKeyToMonday(weekKey)
  const order: Record<typeof day, number> = { mon: 0, tue: 1, wed: 2, thu: 3, fri: 4, sat: 5, sun: 6 }
  mon.setDate(mon.getDate() + order[day])
  return formatDateString(mon)
}
