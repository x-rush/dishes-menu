-- 0002_multi_dish_per_slot.up.sql
-- 每个 (week_key, day_key, slot) 允许多道菜,加 seq 区分顺序

ALTER TABLE week_menus
  ADD COLUMN seq INT NOT NULL DEFAULT 0 AFTER slot;

-- 旧唯一键替换为包含 seq 的复合键
ALTER TABLE week_menus
  DROP INDEX uk_week_day_slot,
  ADD UNIQUE KEY uk_week_day_slot_seq (week_key, day_key, slot, seq);
