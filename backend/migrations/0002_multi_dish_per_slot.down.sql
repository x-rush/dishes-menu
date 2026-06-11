-- 0002_multi_dish_per_slot.down.sql

-- 反向:恢复旧唯一键(每个 (week, day, slot) 只留 seq=0)
-- 假设 seq > 0 的行已被业务清理;否则此处直接报错
DELETE FROM week_menus WHERE seq > 0;

ALTER TABLE week_menus
  DROP INDEX uk_week_day_slot_seq,
  ADD UNIQUE KEY uk_week_day_slot (week_key, day_key, slot);

ALTER TABLE week_menus DROP COLUMN seq;
