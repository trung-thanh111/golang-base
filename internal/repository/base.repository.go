package repository

import (
	// e "errors"
	// f "fmt"
	g "github.com/gin-gonic/gin"
)

type BaseRepository[T any] struct {
	// DB *gorm.DB //field lưu referance đên DB session manager của GORM để thao tác với DB
}

// func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
// 	return &BaseRepository[T]{DB: db}
// }

func (r *BaseRepository[T]) Paginate(c *g.Context) []T {
	return []T{}
}
