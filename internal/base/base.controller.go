package base

import (
	"github.com/gin-gonic/gin"
)

// handle <=> controller
type BaseController[T any] struct {
}

func NewBaseController[T any](c *gin.Context) *BaseController[T] {
	return &BaseController[T]{}
}

// c === context
func (h *BaseController[T]) Paginate(c *gin.Context) {

}
