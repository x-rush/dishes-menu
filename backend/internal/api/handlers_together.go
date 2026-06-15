package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"dishes-menu/internal/dao"
)

const togetherSinceKey = "together_since"

type TogetherHandler struct {
	repo *dao.CounterDAO
}

func NewTogetherHandler(repo *dao.CounterDAO) *TogetherHandler { return &TogetherHandler{repo: repo} }

type togetherSetReq struct {
	Date string `json:"date"`
}

// Get GET /api/together — 返回 together_since 日期与距今天数。
// 未设置时返回 {together_since: null, days: 0},不当作 404。
func (h *TogetherHandler) Get(c *gin.Context) {
	v, err := h.repo.Get(c.Request.Context(), togetherSinceKey)
	if err == nil {
		t, perr := time.Parse("2006-01-02", v)
		if perr != nil {
			// 服务端数据损坏 (FE 没传 v),应返回 500
			serverErr(c, fmt.Errorf("corrupt counter %q value: %w", togetherSinceKey, perr))
			return
		}
		days := int(time.Since(t).Hours() / 24)
		if days < 0 {
			days = 0
		}
		c.JSON(200, gin.H{"together_since": v, "days": days})
		return
	}
	if errors.Is(err, dao.ErrCounterNotFound) {
		c.JSON(200, gin.H{"together_since": nil, "days": 0})
		return
	}
	serverErr(c, err)
}

// Set POST /api/together — 设置 / 更新 together_since 日期。
func (h *TogetherHandler) Set(c *gin.Context) {
	var body togetherSetReq
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if body.Date == "" {
		badRequest(c, "date 不能为空")
		return
	}
	if _, err := time.Parse("2006-01-02", body.Date); err != nil {
		badRequest(c, "date 格式应为 YYYY-MM-DD")
		return
	}
	if err := h.repo.Set(c.Request.Context(), togetherSinceKey, body.Date); err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true, "together_since": body.Date})
}
