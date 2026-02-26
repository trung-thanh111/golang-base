package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	response "golang-base/pkg/response"
)

// NewRouter — tạo gin.Engine KHÔNG có default middleware
// Middleware sẽ được add riêng ở initialize/run.go
func NewRouter(log *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New() // KHÔNG dùng gin.Default() → tránh log trùng
	return r
}

// RegisterRoutes — đăng ký tất cả routes theo module
func RegisterRoutes(r *gin.Engine, log *zap.Logger, db *gorm.DB) {
	// Health check — không cần auth
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status": "ok",
		})
	})

	// API v1
	v1 := r.Group("/v1")
	{
		v1.GET("/ping", Pong)
		v1.GET("/ping/:name", PongWithName)
	}

	// API v2 — placeholder cho tương lai
	// v2 := r.Group("/v2")
	// {
	// }
}

// Pong — demo endpoint
func Pong(c *gin.Context) {
	name := c.DefaultQuery("name", "world")
	id := c.Query("id")

	response.OK(c, gin.H{
		"message": "pong " + name,
		"id":      id,
	})
}

// PongWithName — demo endpoint with path param
func PongWithName(c *gin.Context) {
	name := c.Param("name")

	c.JSON(http.StatusOK, gin.H{
		"message": "pong " + name,
	})
}
