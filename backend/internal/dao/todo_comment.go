package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"dishes-menu/internal/model"
)

type TodoCommentDAO struct{ db *sqlx.DB }

func NewTodoCommentDAO(db *sqlx.DB) *TodoCommentDAO { return &TodoCommentDAO{db: db} }

// List 列出某 todo 的全部评论(按 created_at 升序,最早在前)。
func (d *TodoCommentDAO) List(ctx context.Context, todoID int64) ([]model.TodoComment, error) {
	var comments []model.TodoComment
	err := d.db.SelectContext(ctx, &comments, `
		SELECT id, todo_id, content, author_emoji, author_color, created_at
		FROM todo_comments
		WHERE todo_id = ?
		ORDER BY created_at ASC, id ASC
	`, todoID)
	return comments, err
}

// Add 添加一条评论,返回新评论(带 id 和 created_at)。
func (d *TodoCommentDAO) Add(ctx context.Context, todoID int64, content, emoji, color string) (model.TodoComment, error) {
	res, err := d.db.ExecContext(ctx, `
		INSERT INTO todo_comments (todo_id, content, author_emoji, author_color)
		VALUES (?, ?, ?, ?)
	`, todoID, content, emoji, color)
	if err != nil {
		return model.TodoComment{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.TodoComment{}, err
	}
	return d.Get(ctx, id)
}

// Get 单条评论(Add 内部用)。
func (d *TodoCommentDAO) Get(ctx context.Context, id int64) (model.TodoComment, error) {
	var c model.TodoComment
	err := d.db.GetContext(ctx, &c, `
		SELECT id, todo_id, content, author_emoji, author_color, created_at
		FROM todo_comments WHERE id = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.TodoComment{}, ErrNotFound
	}
	if err != nil {
		return model.TodoComment{}, err
	}
	return c, nil
}

// Delete 删除单条评论(目前 P2 不开放 API 给客户端,留作管理用)。
func (d *TodoCommentDAO) Delete(ctx context.Context, id int64) error {
	res, err := d.db.ExecContext(ctx, `DELETE FROM todo_comments WHERE id = ?`, id)
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
