package repository

import (
	"regexp" // dùng regex
)

// validFieldName — whitelist ký tự hợp lệ trong tên field
// Chỉ cho phép: chữ, số, underscore, dấu chấm (cho relation fields)
// → Chặn SQL injection qua field name
//
// Tại sao cần?
// Nếu user gửi field = "id; DROP TABLE users--"
// → Nối vào query = xong, mất table
// Regex này đảm bảo chỉ field name hợp lệ mới được dùng
var fieldNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

func validFieldName(field string) bool {
	return fieldNameRegex.MatchString(field)
}
