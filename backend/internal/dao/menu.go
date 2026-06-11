package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"dishes-menu/internal/model"
)

type menuRow struct {
	ID      int64  `db:"id"`
	WeekKey string `db:"week_key"`
	DayKey  string `db:"day_key"`
	Slot    string `db:"slot"`
	Seq     int    `db:"seq"`
	DishID  string `db:"dish_id"`
	Note    string `db:"note"`
}

type MenuRepo struct{ db *sqlx.DB }

func NewMenuRepo(db *sqlx.DB) *MenuRepo { return &MenuRepo{db: db} }

// GetWeek returns all menu items for a given week, aggregated into per-slot slices.
// Rows are ordered by seq ASC so insertion order is preserved end-to-end.
func (r *MenuRepo) GetWeek(ctx context.Context, weekKey string) (*model.WeekMenu, error) {
	var rows []menuRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, week_key, day_key, slot, seq, dish_id, note
		FROM week_menus WHERE week_key = ?
		ORDER BY seq ASC
	`, weekKey); err != nil {
		return nil, err
	}
	w := &model.WeekMenu{}
	for _, row := range rows {
		day := w.Day(row.DayKey)
		if day == nil {
			continue
		}
		day.AddItem(model.Slot(row.Slot), &model.MenuItem{
			Seq:    row.Seq,
			DishID: row.DishID,
			Note:   row.Note,
		})
	}
	w.Normalize()
	return w, nil
}

// AppendItem inserts a new menu item at the end of the slot list (max seq + 1).
// Uses SELECT ... FOR UPDATE to prevent concurrent inserts from picking the same seq.
func (r *MenuRepo) AppendItem(ctx context.Context, weekKey, dayKey string, slot model.Slot, item *model.MenuItem) error {
	if !model.ValidDay(dayKey) {
		return errors.New("invalid day_key")
	}
	if !slot.Valid() {
		return errors.New("invalid slot")
	}
	if item == nil {
		return errors.New("item is nil")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var maxSeq sql.NullInt64
	if err := tx.GetContext(ctx, &maxSeq, `
		SELECT MAX(seq) FROM week_menus
		WHERE week_key = ? AND day_key = ? AND slot = ?
		FOR UPDATE
	`, weekKey, dayKey, string(slot)); err != nil {
		return err
	}
	nextSeq := 0
	if maxSeq.Valid {
		nextSeq = int(maxSeq.Int64) + 1
	}
	item.Seq = nextSeq

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO week_menus (week_key, day_key, slot, seq, dish_id, note)
		VALUES (?, ?, ?, ?, ?, ?)
	`, weekKey, dayKey, string(slot), nextSeq, item.DishID, item.Note); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteItem removes a menu item by seq and renumbers subsequent items in the same slot.
// Transactional so the list stays contiguous.
func (r *MenuRepo) DeleteItem(ctx context.Context, weekKey, dayKey string, slot model.Slot, seq int) error {
	if !model.ValidDay(dayKey) {
		return errors.New("invalid day_key")
	}
	if !slot.Valid() {
		return errors.New("invalid slot")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		DELETE FROM week_menus
		WHERE week_key = ? AND day_key = ? AND slot = ? AND seq = ?
	`, weekKey, dayKey, string(slot), seq)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE week_menus SET seq = seq - 1
		WHERE week_key = ? AND day_key = ? AND slot = ? AND seq > ?
	`, weekKey, dayKey, string(slot), seq); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateItemNote updates only the note field of a specific item (seq-targeted).
func (r *MenuRepo) UpdateItemNote(ctx context.Context, weekKey, dayKey string, slot model.Slot, seq int, note string) error {
	if !model.ValidDay(dayKey) {
		return errors.New("invalid day_key")
	}
	if !slot.Valid() {
		return errors.New("invalid slot")
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE week_menus SET note = ?
		WHERE week_key = ? AND day_key = ? AND slot = ? AND seq = ?
	`, note, weekKey, dayKey, string(slot), seq)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MenuRepo) ListRecentDishIDs(ctx context.Context, weekKey string, weeksBack int) (map[string]int, error) {
	if weeksBack <= 0 {
		weeksBack = 1
	}
	var rows []struct {
		DishID string `db:"dish_id"`
		Cnt    int    `db:"cnt"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT dish_id, COUNT(*) AS cnt FROM week_menus
		WHERE STR_TO_DATE(CONCAT(week_key, '-1'), '%X-W%V-%w') >=
		      DATE_SUB(STR_TO_DATE(CONCAT(?, '-1'), '%X-W%V-%w'), INTERVAL ? WEEK)
		GROUP BY dish_id
	`, weekKey, weeksBack)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]int{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, r := range rows {
		out[r.DishID] = r.Cnt
	}
	return out, nil
}
