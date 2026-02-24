package pkg

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse là cấu trúc chuẩn cho tất cả HTTP response
type APIResponse struct {
	Status        bool   `json:"status"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Data          any    `json:"data,omitempty"`
	Errors        any    `json:"errors,omitempty"`
	ExecutionTime string `json:"execution_time"` // Thời gian xử lý request (ms)
}

// getExecutionTime tính thời gian xử lý request từ context
// Trả về chuỗi dạng "12.34ms", nếu không có thì trả "N/A"
func getExecutionTime(c *gin.Context) string {
	startTime, exists := c.Get("request_start_time")
	if !exists {
		return "N/A"
	}

	start, ok := startTime.(time.Time)
	if !ok {
		return "N/A"
	}

	duration := time.Since(start)
	return fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6)
}

// Success trả về response thành công kèm data
func Success(c *gin.Context, code int, message string, data any) {
	c.JSON(code, APIResponse{
		Status:        true,
		Code:          code,
		Message:       message,
		Data:          data,
		ExecutionTime: getExecutionTime(c),
	})
}

// Error trả về response lỗi kèm errors
func Error(c *gin.Context, code int, message string, errors any) {
	c.JSON(code, APIResponse{
		Status:        false,
		Code:          code,
		Message:       message,
		Errors:        errors,
		ExecutionTime: getExecutionTime(c),
	})
}

// OK - 200
func OK(c *gin.Context, data any) {
	Success(c, HTTP_OK, "success", data)
}

// BadRequest - 400
func BadRequest(c *gin.Context, errors any) {
	Error(c, HTTP_BAD_REQUEST, "bad request", errors)
}

// InternalServerError - 500
func InternalServerError(c *gin.Context, errors any) {
	Error(c, HTTP_INTERNAL_SERVER_ERROR, "internal server error", errors)
}
