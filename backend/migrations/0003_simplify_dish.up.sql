-- 0003_simplify_dish.up.sql
-- 简化菜品模型:DROP taste/ingredient/difficulty,ADD ingredients/estimated_time
-- 旧数据丢弃(用户接受重录)

ALTER TABLE dishes
  DROP COLUMN taste,
  DROP COLUMN ingredient,
  DROP COLUMN difficulty,
  ADD COLUMN ingredients   JSON   NOT NULL AFTER note,
  ADD COLUMN estimated_time INT   NOT NULL DEFAULT 0 AFTER ingredients;
