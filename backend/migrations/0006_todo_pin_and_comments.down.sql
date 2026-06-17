-- 0006_todo_pin_and_comments.down.sql
-- 回滚:删 todo_comments + 移除 todos.pinned

DROP TABLE IF EXISTS todo_comments;

ALTER TABLE todos
  DROP INDEX idx_pinned,
  DROP COLUMN pinned;
