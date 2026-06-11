package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError is the unified error envelope returned to clients.
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status"`
}

type apiErrorBody struct {
	Error APIError `json:"error"`
}

func abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, apiErrorBody{Error: APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}})
}

func badRequest(c *gin.Context, msg string) { abort(c, http.StatusBadRequest, "BAD_REQUEST", msg) }
func notFound(c *gin.Context, msg string)   { abort(c, http.StatusNotFound, "NOT_FOUND", msg) }
func serverErr(c *gin.Context, err error)   {
	abort(c, http.StatusInternalServerError, "INTERNAL", err.Error())
}
