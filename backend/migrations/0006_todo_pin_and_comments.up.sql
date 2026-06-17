-- 0006_todo_pin_and_comments.up.sql
-- 置顶 + 一句话评论

SET NAMES utf8mb4;

ALTER TABLE todos
  ADD COLUMN pinned TINYINT(1) NOT NULL DEFAULT 0 AFTER author_color,
  ADD INDEX idx_pinned (pinned, due_date);

CREATE TABLE IF NOT EXISTS todo_comments (
  id            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
  todo_id       BIGINT        NOT NULL,
  content       VARCHAR(500)  NOT NULL,
  author_emoji  VARCHAR(8)    NOT NULL DEFAULT '',
  author_color  VARCHAR(16)   NOT NULL DEFAULT '',
  created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_todo_created (todo_id, created_at),
  CONSTRAINT fk_comment_todo FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
