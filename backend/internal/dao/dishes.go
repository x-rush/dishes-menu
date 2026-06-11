package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"dishes-menu/internal/model"
)

var ErrNotFound = errors.New("not found")

type dishesRow struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	Slots         JSON      `db:"slots"`
	Ingredients   JSON      `db:"ingredients"`
	EstimatedTime int       `db:"estimated_time"`
	Note          string    `db:"note"`
	Image         string    `db:"image"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (r dishesRow) toModel() (*model.Dish, error) {
	var slots []model.Slot
	if err := json.Unmarshal(r.Slots, &slots); err != nil {
		return nil, fmt.Errorf("unmarshal slots: %w", err)
	}
	var ingredients []string
	if len(r.Ingredients) > 0 {
		if err := json.Unmarshal(r.Ingredients, &ingredients); err != nil {
			return nil, fmt.Errorf("unmarshal ingredients: %w", err)
		}
	}
	if ingredients == nil {
		ingredients = []string{}
	}
	return &model.Dish{
		ID:            r.ID,
		Name:          r.Name,
		Slots:         slots,
		Ingredients:   ingredients,
		EstimatedTime: r.EstimatedTime,
		Note:          r.Note,
		Image:         r.Image,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}, nil
}

type DishRepo struct{ db *sqlx.DB }

func NewDishRepo(db *sqlx.DB) *DishRepo { return &DishRepo{db: db} }

func (r *DishRepo) List(ctx context.Context) ([]*model.Dish, error) {
	var rows []dishesRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, name, slots, ingredients, estimated_time, note, image, created_at, updated_at
		FROM dishes ORDER BY created_at ASC, id ASC
	`); err != nil {
		return nil, err
	}
	out := make([]*model.Dish, 0, len(rows))
	for i := range rows {
		d, err := rows[i].toModel()
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *DishRepo) Get(ctx context.Context, id string) (*model.Dish, error) {
	var row dishesRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, name, slots, ingredients, estimated_time, note, image, created_at, updated_at
		FROM dishes WHERE id = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toModel()
}

func (r *DishRepo) Create(ctx context.Context, d *model.Dish) error {
	if strings.TrimSpace(d.ID) == "" {
		d.ID = newDishID()
	}
	slots, err := json.Marshal(d.Slots)
	if err != nil {
		return fmt.Errorf("marshal slots: %w", err)
	}
	if d.Ingredients == nil {
		d.Ingredients = []string{}
	}
	ingredients, err := json.Marshal(d.Ingredients)
	if err != nil {
		return fmt.Errorf("marshal ingredients: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO dishes (id, name, slots, ingredients, estimated_time, note, image)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.Name, slots, ingredients, d.EstimatedTime, d.Note, d.Image)
	return err
}

// newDishID returns an 8-char short ID derived from a UUID v4.
// Compact enough for URLs and QR codes; collision-resistant for personal-scale use.
func newDishID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
}

func (r *DishRepo) Update(ctx context.Context, d *model.Dish) error {
	slots, err := json.Marshal(d.Slots)
	if err != nil {
		return fmt.Errorf("marshal slots: %w", err)
	}
	if d.Ingredients == nil {
		d.Ingredients = []string{}
	}
	ingredients, err := json.Marshal(d.Ingredients)
	if err != nil {
		return fmt.Errorf("marshal ingredients: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE dishes
		SET name = ?, slots = ?, ingredients = ?, estimated_time = ?, note = ?, image = ?
		WHERE id = ?
	`, d.Name, slots, ingredients, d.EstimatedTime, d.Note, d.Image, d.ID)
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

func (r *DishRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM dishes WHERE id = ?`, id)
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

func (r *DishRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, `SELECT COUNT(*) FROM dishes`)
	return n, err
}
