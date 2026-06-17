package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/model"
)

type TodoHandler struct {
	repo *dao.TodoDAO
}

func NewTodoHandler(repo *dao.TodoDAO) *TodoHandler { return &TodoHandler{repo: repo} }

type todoCreateReq struct {
	Content     string  `json:"content"`
	DueDate     *string `json:"due_date"` // "YYYY-MM-DD" 或 null
	AuthorEmoji string  `json:"author_emoji"`
	AuthorColor string  `json:"author_color"`
}

type todoPatchReq struct {
	Content   *string         `json:"content"`
	Completed *bool           `json:"completed"`
	DueDate   json.RawMessage `json:"due_date"` // 区分缺失 vs null(清除)
	Pinned    *bool           `json:"pinned"`
}

// List GET /api/todos — 列出所有待办(未完成优先 + 截止日期升序)。
func (h *TodoHandler) List(c *gin.Context) {
	todos, err := h.repo.List(c.Request.Context())
	if err != nil {
		serverErr(c, err)
		return
	}
	if todos == nil {
		todos = []model.Todo{} // 让前端拿到 [] 而不是 null
	}
	c.JSON(200, todos)
}

// Create POST /api/todos — 新建一条待办。due_date 可选 (YYYY-MM-DD 或 null)。
func (h *TodoHandler) Create(c *gin.Context) {
	var body todoCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if body.Content == "" || body.AuthorEmoji == "" || body.AuthorColor == "" {
		badRequest(c, "content / author_emoji / author_color 不能为空")
		return
	}
	var due *time.Time
	if body.DueDate != nil && *body.DueDate != "" {
		t, err := time.Parse("2006-01-02", *body.DueDate)
		if err != nil {
			badRequest(c, "due_date 格式应为 YYYY-MM-DD")
			return
		}
		due = &t
	}
	todo, err := h.repo.Create(c.Request.Context(), body.Content, due, body.AuthorEmoji, body.AuthorColor)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(201, todo)
}

// Patch PATCH /api/todos/:id — 支持同时改 content / completed / due_date / pinned;至少一个字段。
// 注意:Completed 与 Pinned 都互斥(都是切换语义),不与其他字段共存。
func (h *TodoHandler) Patch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		badRequest(c, "invalid id")
		return
	}
	var body todoPatchReq
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if body.Completed != nil {
		// ToggleComplete 互斥 — 不与 content/due_date/pinned 混用
		if body.Content != nil || len(body.DueDate) > 0 || body.Pinned != nil {
			badRequest(c, "completed 不能与其他字段同时修改")
			return
		}
		todo, err := h.repo.ToggleComplete(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, dao.ErrNotFound) {
				notFound(c, "todo not found")
				return
			}
			serverErr(c, err)
			return
		}
		c.JSON(200, todo)
		return
	}
	if body.Pinned != nil {
		// TogglePin 互斥
		if body.Content != nil || len(body.DueDate) > 0 {
			badRequest(c, "pinned 不能与其他字段同时修改")
			return
		}
		todo, err := h.repo.TogglePin(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, dao.ErrNotFound) {
				notFound(c, "todo not found")
				return
			}
			serverErr(c, err)
			return
		}
		c.JSON(200, todo)
		return
	}
	if body.Content == nil && len(body.DueDate) == 0 {
		badRequest(c, "未指定要更新的字段")
		return
	}
	if body.Content != nil {
		if *body.Content == "" {
			badRequest(c, "content 不能为空")
			return
		}
		if err := h.repo.UpdateContent(c.Request.Context(), id, *body.Content); err != nil {
			if errors.Is(err, dao.ErrNotFound) {
				notFound(c, "todo not found")
				return
			}
			serverErr(c, err)
			return
		}
	}
	if len(body.DueDate) > 0 {
		var due *time.Time
		if !bytes.Equal(body.DueDate, []byte("null")) {
			var s string
			if err := json.Unmarshal(body.DueDate, &s); err != nil {
				badRequest(c, "due_date 格式应为 YYYY-MM-DD 或 null")
				return
			}
			if s != "" {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					badRequest(c, "due_date 格式应为 YYYY-MM-DD")
					return
				}
				due = &t
			}
		}
		if err := h.repo.UpdateDueDate(c.Request.Context(), id, due); err != nil {
			if errors.Is(err, dao.ErrNotFound) {
				notFound(c, "todo not found")
				return
			}
			serverErr(c, err)
			return
		}
	}
	todo, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(200, todo)
}

// Delete DELETE /api/todos/:id — 删除一条待办。
func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		badRequest(c, "invalid id")
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			notFound(c, "todo not found")
			return
		}
		serverErr(c, err)
		return
	}
	c.Status(204)
}
