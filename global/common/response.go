package common

// ============================================================
// PAGINATE RESULT — Kết quả phân trang trả về cho client
// Dùng chung cho mọi data models trong dự án
// ============================================================
type PaginateResult[T any] struct {
	Data       []T   `json:"data"`        // mảng data
	Total      int64 `json:"total"`       // tổng số (offset), -1 (keyset)
	NextCursor any   `json:"next_cursor"` // cursor cho trang tiếp (keyset)
	HasMore    bool  `json:"has_more"`    // còn trang tiếp không?
}
