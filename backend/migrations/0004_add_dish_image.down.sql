-- 0004_add_dish_image.down.sql
-- 回滚 image 列

ALTER TABLE dishes
  DROP COLUMN image;
