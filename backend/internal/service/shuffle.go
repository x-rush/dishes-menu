package service

import (
	"context"
	"fmt"
	"math/rand"
	"slices"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/model"
)

type ShuffleService struct {
	dishes *dao.DishRepo
	menus  *dao.MenuRepo
}

func NewShuffleService(d *dao.DishRepo, m *dao.MenuRepo) *ShuffleService {
	return &ShuffleService{dishes: d, menus: m}
}

func (s *ShuffleService) Shuffle(ctx context.Context, day string, slot model.Slot, weekKey string) (*model.Dish, error) {
	if !model.ValidDay(day) {
		return nil, fmt.Errorf("invalid day: %s", day)
	}
	if !slot.Valid() {
		return nil, fmt.Errorf("invalid slot: %s", slot)
	}

	candidates, err := s.candidates(ctx, slot)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no matching dish in library")
	}

	history, err := s.menus.ListRecentDishIDs(ctx, weekKey, 4)
	if err != nil {
		return nil, err
	}

	week, err := s.menus.GetWeek(ctx, weekKey)
	if err != nil {
		return nil, err
	}
	dayMenu := week.Day(day)
	sameDayDishIDs := map[string]struct{}{}
	if dayMenu != nil {
		// 同一天其他 slot 选过的菜降权 ×0.1,避免一天三顿都是番茄
		// 同一 slot 已选的菜不降权 — 允许「随便来」重复推同一道,这是设计决定
		for _, s2 := range model.SlotsForWeekday(day) {
			if s2 == slot {
				continue
			}
			for _, it := range dayMenu.Items(s2) {
				sameDayDishIDs[it.DishID] = struct{}{}
			}
		}
	}

	pool := make([]weighted, 0, len(candidates))
	for _, d := range candidates {
		w := 1.0
		if cnt, ok := history[d.ID]; ok {
			w = 1.0 / float64(1+cnt)
		}
		if _, ok := sameDayDishIDs[d.ID]; ok {
			w *= 0.1
		}
		pool = append(pool, weighted{dish: d, weight: w})
	}
	return weightedPick(pool).dish, nil
}

type weighted struct {
	dish   *model.Dish
	weight float64
}

func (s *ShuffleService) candidates(ctx context.Context, slot model.Slot) ([]*model.Dish, error) {
	all, err := s.dishes.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*model.Dish, 0, len(all))
	for _, d := range all {
		if slices.Contains(d.Slots, slot) {
			out = append(out, d)
		}
	}
	return out, nil
}

func weightedPick(pool []weighted) weighted {
	var total float64
	for _, p := range pool {
		total += p.weight
	}
	if total <= 0 {
		return pool[rand.Intn(len(pool))]
	}
	r := rand.Float64() * total
	var acc float64
	for _, p := range pool {
		acc += p.weight
		if r <= acc {
			return p
		}
	}
	return pool[len(pool)-1]
}
