-- 0004_add_dish_image.up.sql
-- 加 image 列(URL 字符串,可选,默认空串)
-- 之前 model.Dish.Image 字段已经存在但迁移漏了

ALTER TABLE dishes
  ADD COLUMN image VARCHAR(500) NOT NULL DEFAULT '' AFTER estimated_time;
