package middlewares

import (
	"github.com/gin-gonic/gin"
	"time"
)

const RequestStartTimeKey = "request_start_time"

// RequestTimer lưu thời điểm bắt đầu request vào context
// Dùng làm middleware trước tất cả các route
func RequestTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lưu thời điểm nhận request vào context
		c.Set(RequestStartTimeKey, time.Now())
		c.Next()
	}
}
