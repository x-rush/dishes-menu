-- 0001_init.up.sql
-- 建 dishes 菜品库 + week_menus 每周菜单

CREATE TABLE IF NOT EXISTS dishes (
    id          VARCHAR(32)  NOT NULL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL,
    slots       JSON         NOT NULL,
    taste       JSON         NOT NULL,
    ingredient  JSON         NOT NULL,
    difficulty  ENUM('quick','standard') NOT NULL,
    note        TEXT         NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dishes_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS week_menus (
    id           BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    week_key     VARCHAR(10)  NOT NULL,
    day_key      VARCHAR(8)   NOT NULL,
    slot         ENUM('breakfast','lunch','dinner','snack') NOT NULL,
    dish_id      VARCHAR(32)  NOT NULL,
    note         TEXT         NOT NULL,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_week_day_slot (week_key, day_key, slot),
    INDEX idx_week (week_key),
    CONSTRAINT fk_week_menus_dish FOREIGN KEY (dish_id) REFERENCES dishes(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
