package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"dishes-menu/internal/model"
)

type TodoDAO struct{ db *sqlx.DB }

func NewTodoDAO(db *sqlx.DB) *TodoDAO { return &TodoDAO{db: db} }

func (d *TodoDAO) List(ctx context.Context) ([]model.Todo, error) {
	var todos []model.Todo
	err := d.db.SelectContext(ctx, &todos, `
		SELECT id, content, due_date, author_emoji, author_color, created_at, completed_at
		FROM todos
		ORDER BY (completed_at IS NULL) DESC, due_date IS NULL, due_date, created_at DESC
	`)
	return todos, err
}

func (d *TodoDAO) Create(ctx context.Context, content string, dueDate *time.Time, emoji, color string) (model.Todo, error) {
	res, err := d.db.ExecContext(ctx,
		`INSERT INTO todos (content, due_date, author_emoji, author_color) VALUES (?, ?, ?, ?)`,
		content, dueDate, emoji, color,
	)
	if err != nil {
		return model.Todo{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Todo{}, err
	}
	return d.Get(ctx, id)
}

func (d *TodoDAO) Get(ctx context.Context, id int64) (model.Todo, error) {
	var t model.Todo
	err := d.db.GetContext(ctx, &t, `
		SELECT id, content, due_date, author_emoji, author_color, created_at, completed_at
		FROM todos WHERE id = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, ErrNotFound
	}
	if err != nil {
		return model.Todo{}, err
	}
	return t, nil
}

func (d *TodoDAO) ToggleComplete(ctx context.Context, id int64) (model.Todo, error) {
	now := time.Now()
	// 已完成 → 取消完成(completed_at = NULL);未完成 → 标记完成
	res, err := d.db.ExecContext(ctx, `
		UPDATE todos
		SET completed_at = CASE WHEN completed_at IS NULL THEN ? ELSE NULL END
		WHERE id = ?
	`, now, id)
	if err != nil {
		return model.Todo{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return model.Todo{}, err
	}
	if n == 0 {
		return model.Todo{}, ErrNotFound
	}
	return d.Get(ctx, id)
}

func (d *TodoDAO) UpdateContent(ctx context.Context, id int64, content string) error {
	res, err := d.db.ExecContext(ctx, `UPDATE todos SET content = ? WHERE id = ?`, content, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *TodoDAO) Delete(ctx context.Context, id int64) error {
	res, err := d.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
