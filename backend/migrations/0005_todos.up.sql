-- 0005_todos.up.sql
-- 待办 + 一同吃饭计数器

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS todos (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  content       VARCHAR(500) NOT NULL,
  due_date      DATE         NULL,
  author_emoji  VARCHAR(8)   NOT NULL,
  author_color  VARCHAR(16)  NOT NULL,
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at  DATETIME     NULL,
  INDEX idx_todos_open (completed_at, due_date, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS counters (
  name   VARCHAR(64)  NOT NULL PRIMARY KEY,
  value  VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
