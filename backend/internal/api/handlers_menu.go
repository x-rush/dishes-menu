package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/model"
	"dishes-menu/internal/service"
)

type MenuHandler struct {
	menus   *dao.MenuRepo
	dishes  *dao.DishRepo
	shuffle *service.ShuffleService
}

func NewMenuHandler(menus *dao.MenuRepo, dishes *dao.DishRepo, sh *service.ShuffleService) *MenuHandler {
	return &MenuHandler{menus: menus, dishes: dishes, shuffle: sh}
}

func (h *MenuHandler) GetWeek(c *gin.Context) {
	week := c.Query("week")
	if week == "" {
		week = model.ISOWeekKey(time.Now())
	}
	w, err := h.menus.GetWeek(c.Request.Context(), week)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"week": week, "menu": w})
}

type appendItemReq struct {
	DishID string `json:"dish_id"`
	Note   string `json:"note"`
}

// AppendItem POST /api/menu/:day/:slot — add a new dish to the slot's list.
func (h *MenuHandler) AppendItem(c *gin.Context) {
	day := c.Param("day")
	slot := model.Slot(c.Param("slot"))
	if !model.ValidDay(day) {
		badRequest(c, "invalid day")
		return
	}
	if !slot.Valid() {
		badRequest(c, "invalid slot")
		return
	}
	week := c.Query("week")
	if week == "" {
		week = model.ISOWeekKey(time.Now())
	}
	var req appendItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if req.DishID == "" {
		badRequest(c, "dish_id required")
		return
	}
	if _, err := h.dishes.Get(c.Request.Context(), req.DishID); err != nil {
		if err == dao.ErrNotFound {
			badRequest(c, "dish_id not found")
			return
		}
		serverErr(c, err)
		return
	}
	item := &model.MenuItem{DishID: req.DishID, Note: req.Note}
	if err := h.menus.AppendItem(c.Request.Context(), week, day, slot, item); err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"week": week, "day": day, "slot": slot, "item": item})
}

type updateNoteReq struct {
	Note string `json:"note"`
}

// UpdateItemNote PUT /api/menu/:day/:slot/:seq — update note of one item.
func (h *MenuHandler) UpdateItemNote(c *gin.Context) {
	day := c.Param("day")
	slot := model.Slot(c.Param("slot"))
	if !model.ValidDay(day) {
		badRequest(c, "invalid day")
		return
	}
	if !slot.Valid() {
		badRequest(c, "invalid slot")
		return
	}
	seq, err := strconv.Atoi(c.Param("seq"))
	if err != nil || seq < 0 {
		badRequest(c, "invalid seq")
		return
	}
	week := c.Query("week")
	if week == "" {
		week = model.ISOWeekKey(time.Now())
	}
	var req updateNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if err := h.menus.UpdateItemNote(c.Request.Context(), week, day, slot, seq, req.Note); err != nil {
		if err == dao.ErrNotFound {
			abort(c, http.StatusNotFound, "NOT_FOUND", "item not found")
			return
		}
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"week": week, "day": day, "slot": slot, "seq": seq, "note": req.Note})
}

// DeleteItem DELETE /api/menu/:day/:slot/:seq — remove one item, renumber subsequent.
func (h *MenuHandler) DeleteItem(c *gin.Context) {
	day := c.Param("day")
	slot := model.Slot(c.Param("slot"))
	if !model.ValidDay(day) {
		badRequest(c, "invalid day")
		return
	}
	if !slot.Valid() {
		badRequest(c, "invalid slot")
		return
	}
	seq, err := strconv.Atoi(c.Param("seq"))
	if err != nil || seq < 0 {
		badRequest(c, "invalid seq")
		return
	}
	week := c.Query("week")
	if week == "" {
		week = model.ISOWeekKey(time.Now())
	}
	if err := h.menus.DeleteItem(c.Request.Context(), week, day, slot, seq); err != nil {
		if err == dao.ErrNotFound {
			abort(c, http.StatusNotFound, "NOT_FOUND", "item not found")
			return
		}
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"week": week, "day": day, "slot": slot, "seq": seq})
}

func (h *MenuHandler) Shuffle(c *gin.Context) {
	day := c.Query("day")
	slot := model.Slot(c.Query("slot"))
	if !model.ValidDay(day) {
		badRequest(c, "missing or invalid day")
		return
	}
	if !slot.Valid() {
		badRequest(c, "missing or invalid slot")
		return
	}
	week := c.Query("week")
	if week == "" {
		week = model.ISOWeekKey(time.Now())
	}
	dish, err := h.shuffle.Shuffle(c.Request.Context(), day, slot, week)
	if err != nil {
		if strings.HasPrefix(err.Error(), "no matching dish") {
			abort(c, http.StatusUnprocessableEntity, "NO_MATCH", err.Error())
			return
		}
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"dish": dish, "week": week, "day": day, "slot": slot})
}
