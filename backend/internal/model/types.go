package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Slot string

const (
	SlotBreakfast Slot = "breakfast"
	SlotLunch     Slot = "lunch"
	SlotDinner    Slot = "dinner"
	SlotSnack     Slot = "snack"
)

func (s Slot) Valid() bool {
	switch s {
	case SlotBreakfast, SlotLunch, SlotDinner, SlotSnack:
		return true
	}
	return false
}

// SlotsForWeekday returns the slots available on a given day.
// Mon-Fri: breakfast + snack; Sat-Sun: breakfast + lunch + dinner.
func SlotsForWeekday(weekday string) []Slot {
	switch weekday {
	case "sat", "sun":
		return []Slot{SlotBreakfast, SlotLunch, SlotDinner}
	default:
		return []Slot{SlotBreakfast, SlotSnack}
	}
}

type Dish struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slots         []Slot    `json:"slots"`
	Ingredients   []string  `json:"ingredients"`
	EstimatedTime int       `json:"estimated_time"`
	Note          string    `json:"note"`
	Image         string    `json:"image,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MenuItem struct {
	Seq    int    `json:"seq"`
	DishID string `json:"dish_id"`
	Note   string `json:"note"`
}

type DayMenu struct {
	Breakfast []*MenuItem `json:"breakfast"`
	Lunch     []*MenuItem `json:"lunch"`
	Dinner    []*MenuItem `json:"dinner"`
	Snack     []*MenuItem `json:"snack"`
}

// Items returns the slice of menu items for the given slot (empty slice, never nil).
func (d DayMenu) Items(slot Slot) []*MenuItem {
	switch slot {
	case SlotBreakfast:
		return d.Breakfast
	case SlotLunch:
		return d.Lunch
	case SlotDinner:
		return d.Dinner
	case SlotSnack:
		return d.Snack
	}
	return nil
}

// AddItem appends an item to the given slot. Caller is responsible for assigning Seq.
func (d *DayMenu) AddItem(slot Slot, item *MenuItem) {
	switch slot {
	case SlotBreakfast:
		d.Breakfast = append(d.Breakfast, item)
	case SlotLunch:
		d.Lunch = append(d.Lunch, item)
	case SlotDinner:
		d.Dinner = append(d.Dinner, item)
	case SlotSnack:
		d.Snack = append(d.Snack, item)
	}
}

// Normalize replaces nil slot slices with empty slices so JSON output is []
// instead of null (Go's nil slice → JSON null, empty slice → JSON []).
// Frontend reads .length directly; null.length throws TypeError.
func (d *DayMenu) Normalize() {
	if d.Breakfast == nil {
		d.Breakfast = []*MenuItem{}
	}
	if d.Lunch == nil {
		d.Lunch = []*MenuItem{}
	}
	if d.Dinner == nil {
		d.Dinner = []*MenuItem{}
	}
	if d.Snack == nil {
		d.Snack = []*MenuItem{}
	}
}

// Normalize applies Normalize to all 7 days so a freshly-aggregated WeekMenu
// always serializes with non-null slot arrays.
func (w *WeekMenu) Normalize() {
	w.Mon.Normalize()
	w.Tue.Normalize()
	w.Wed.Normalize()
	w.Thu.Normalize()
	w.Fri.Normalize()
	w.Sat.Normalize()
	w.Sun.Normalize()
}

type WeekMenu struct {
	Mon DayMenu `json:"mon"`
	Tue DayMenu `json:"tue"`
	Wed DayMenu `json:"wed"`
	Thu DayMenu `json:"thu"`
	Fri DayMenu `json:"fri"`
	Sat DayMenu `json:"sat"`
	Sun DayMenu `json:"sun"`
}

func (w *WeekMenu) Day(key string) *DayMenu {
	switch key {
	case "mon":
		return &w.Mon
	case "tue":
		return &w.Tue
	case "wed":
		return &w.Wed
	case "thu":
		return &w.Thu
	case "fri":
		return &w.Fri
	case "sat":
		return &w.Sat
	case "sun":
		return &w.Sun
	}
	return nil
}

const (
	DayMon = "mon"
	DayTue = "tue"
	DayWed = "wed"
	DayThu = "thu"
	DayFri = "fri"
	DaySat = "sat"
	DaySun = "sun"
)

func ValidDay(s string) bool {
	switch s {
	case DayMon, DayTue, DayWed, DayThu, DayFri, DaySat, DaySun:
		return true
	}
	return false
}

var AllDays = []string{DayMon, DayTue, DayWed, DayThu, DayFri, DaySat, DaySun}

// ISOWeekKey returns "YYYY-Www" like "2026-W24".
func ISOWeekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}

// Todo mirrors the todos table. Used directly by sqlx (db tags) — no DAO row type by design
// (Todo has no JSON/joined columns). MarshalJSON formats DueDate as YYYY-MM-DD.
type Todo struct {
	ID          int64      `db:"id" json:"id"`
	Content     string     `db:"content" json:"content"`
	DueDate     *time.Time `db:"due_date" json:"due_date,omitempty"`
	AuthorEmoji string     `db:"author_emoji" json:"author_emoji"`
	AuthorColor string     `db:"author_color" json:"author_color"`
	Pinned      bool       `db:"pinned" json:"pinned"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

// TodoComment mirrors the todo_comments table.
type TodoComment struct {
	ID          int64     `db:"id" json:"id"`
	TodoID      int64     `db:"todo_id" json:"todo_id"`
	Content     string    `db:"content" json:"content"`
	AuthorEmoji string    `db:"author_emoji" json:"author_emoji"`
	AuthorColor string    `db:"author_color" json:"author_color"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// MarshalJSON renders DueDate as "YYYY-MM-DD" so the FE's formatDue (which
// does d.slice(5)) renders correctly. Other fields use default encoding.
func (t Todo) MarshalJSON() ([]byte, error) {
	type alias Todo
	aux := struct {
		DueDate *string `json:"due_date"`
		*alias
	}{
		alias: (*alias)(&t),
	}
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		aux.DueDate = &s
	}
	return json.Marshal(aux)
}

type Counter struct {
	Name  string `db:"name" json:"name"`
	Value string `db:"value" json:"value"`
}
