package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/model"
)

type TodoCommentHandler struct {
	repo *dao.TodoCommentDAO
	todo *dao.TodoDAO // 校验 todo 存在
}

func NewTodoCommentHandler(commentRepo *dao.TodoCommentDAO, todoRepo *dao.TodoDAO) *TodoCommentHandler {
	return &TodoCommentHandler{repo: commentRepo, todo: todoRepo}
}

type commentCreateReq struct {
	Content     string `json:"content"`
	AuthorEmoji string `json:"author_emoji"`
	AuthorColor string `json:"author_color"`
}

// List GET /api/todos/:id/comments — 列出某 todo 的全部评论。
func (h *TodoCommentHandler) List(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		badRequest(c, "invalid id")
		return
	}
	// 校验 todo 存在(否则 FK 错误信息对用户不友好)
	if _, err := h.todo.Get(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			notFound(c, "todo not found")
			return
		}
		serverErr(c, err)
		return
	}
	comments, err := h.repo.List(c.Request.Context(), id)
	if err != nil {
		serverErr(c, err)
		return
	}
	if comments == nil {
		comments = []model.TodoComment{} // 让前端拿到 [] 而不是 null
	}
	c.JSON(200, comments)
}

// Create POST /api/todos/:id/comments — 新增一条评论。
func (h *TodoCommentHandler) Create(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		badRequest(c, "invalid id")
		return
	}
	var body commentCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "invalid json: "+err.Error())
		return
	}
	if body.Content == "" {
		badRequest(c, "content 不能为空")
		return
	}
	if body.AuthorEmoji == "" || body.AuthorColor == "" {
		badRequest(c, "author_emoji / author_color 不能为空")
		return
	}
	// 校验 todo 存在
	if _, err := h.todo.Get(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			notFound(c, "todo not found")
			return
		}
		serverErr(c, err)
		return
	}
	comment, err := h.repo.Add(c.Request.Context(), id, body.Content, body.AuthorEmoji, body.AuthorColor)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(201, comment)
}
