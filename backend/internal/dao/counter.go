package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type CounterDAO struct{ db *sqlx.DB }

func NewCounterDAO(db *sqlx.DB) *CounterDAO { return &CounterDAO{db: db} }

var ErrCounterNotFound = errors.New("counter not found")

func (d *CounterDAO) Get(ctx context.Context, name string) (string, error) {
	var value string
	err := d.db.GetContext(ctx, &value, `SELECT value FROM counters WHERE name = ?`, name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrCounterNotFound
	}
	return value, err
}

func (d *CounterDAO) Set(ctx context.Context, name, value string) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO counters (name, value) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value)
	`, name, value)
	return err
}
