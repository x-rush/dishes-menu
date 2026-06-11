-- 0003_simplify_dish.down.sql

-- 旧字段回填为空数据(用户数据已无法恢复)
ALTER TABLE dishes
  DROP COLUMN ingredients,
  DROP COLUMN estimated_time,
  ADD COLUMN taste       JSON NOT NULL AFTER slots,
  ADD COLUMN ingredient  JSON NOT NULL AFTER taste,
  ADD COLUMN difficulty ENUM('quick','standard') NOT NULL AFTER ingredient;
