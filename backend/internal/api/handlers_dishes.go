package api

import (
	"errors"

	"github.com/gin-gonic/gin"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/model"
)

type DishHandler struct {
	repo *dao.DishRepo
}

func NewDishHandler(repo *dao.DishRepo) *DishHandler { return &DishHandler{repo: repo} }

func (h *DishHandler) List(c *gin.Context) {
	dishes, err := h.repo.List(c.Request.Context())
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"dishes": dishes})
}

func (h *DishHandler) Create(c *gin.Context) {
	var d model.Dish
	if err := c.ShouldBindJSON(&d); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if err := validateDish(&d); err != nil {
		badRequest(c, err.Error())
		return
	}
	d.ID = ""
	if err := h.repo.Create(c.Request.Context(), &d); err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(201, d)
}

func (h *DishHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		badRequest(c, "missing id")
		return
	}
	var d model.Dish
	if err := c.ShouldBindJSON(&d); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if err := validateDish(&d); err != nil {
		badRequest(c, err.Error())
		return
	}
	d.ID = id
	if err := h.repo.Update(c.Request.Context(), &d); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			notFound(c, "dish not found")
			return
		}
		serverErr(c, err)
		return
	}
	got, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, got)
}

func (h *DishHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			notFound(c, "dish not found")
			return
		}
		serverErr(c, err)
		return
	}
	c.Status(204)
}

func validateDish(d *model.Dish) error {
	if d.Name == "" {
		return errors.New("name is required")
	}
	if len(d.Slots) == 0 {
		return errors.New("at least one slot is required")
	}
	for _, s := range d.Slots {
		if !s.Valid() {
			return errors.New("invalid slot: " + string(s))
		}
	}
	if d.EstimatedTime < 0 || d.EstimatedTime > 999 {
		return errors.New("estimated_time must be between 0 and 999 minutes")
	}
	return nil
}
