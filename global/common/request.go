package common

// ============================================================
// SPECS — Tham số tìm kiếm dùng chung cho mọi module
//
// Thay vì mỗi module tự define params riêng,
// dùng 1 struct chung → nhất quán, dễ mở rộng
// ============================================================
type Specs struct {
	// --- Tìm kiếm ---
	Keyword       string   // từ khóa search (VD: "iphone 15")
	KeywordFields []string // search trên những field nào (VD: ["name", "title"])

	// --- Quan hệ (Eager Loading) ---
	// Preload relations để tránh N+1 query
	// VD: lấy Product kèm Category, Brand → 1 query thay vì N query
	Relations []string

	// --- Lọc dữ liệu ---
	// Filters: điều kiện WHERE field = value (exact match)
	// VD: {"status": "active", "category_id": 5}
	Filters map[string]any

	// RangeFilters: điều kiện WHERE field BETWEEN min AND max
	// Cực kỳ quan trọng cho ecommerce:
	// → Lọc giá: price 100k-500k
	// → Lọc ngày: orders từ 01/01 đến 31/01
	// → Lọc rating: >= 4 sao
	// Key = tên field, Value = {Min, Max}
	RangeFilters map[string]RangeFilter

	// InFilters: điều kiện WHERE field IN (values)
	// VD: lọc products theo nhiều category cùng lúc
	// {"category_id": [1, 2, 3], "brand_id": [10, 20]}
	InFilters map[string][]any

	// --- Chọn field trả về (Projection) ---
	// Mặc định SELECT * → lãng phí nếu chỉ cần vài field
	// VD: listing chỉ cần ["id", "name", "price", "image"]
	// → Giảm bandwidth, giảm memory, query nhanh hơn
	SelectFields []string

	// --- Sắp xếp ---
	Sort string // VD: "price asc", "created_at desc"

	// --- Offset pagination ---
	// Dùng khi: data < 100k data, cần nhảy trang tự do (trang 1, 5, 10)
	// Nhược điểm: CHẬM DẦN khi offset lớn
	// VD: OFFSET 5000000 → DB scan 5 triệu data rồi bỏ → rất chậm
	Limit  int
	Offset int

	// --- Keyset pagination (Cursor-based) ---
	// Dùng khi: data lớn (>100k), infinite scroll, feed
	// Cách hoạt động: WHERE id < cursor thay vì OFFSET n
	// → Dùng INDEX trên cursor field → O(log n) thay vì O(n)
	// → Tốc độ KHÔNG ĐỔI dù 10 triệu data
	UseKeyset       bool   // true = keyset, false = offset
	CursorField     string // field làm cursor (thường "id" hoặc "created_at")
	CursorValue     any    // giá trị cursor của record cuối trang trước
	CursorDirection string // "lt" = DESC (cũ→mới), "gt" = ASC (mới→cũ)
}

// RangeFilter — định nghĩa khoảng giá trị cho range query
// Min/Max dùng pointer (*any) để phân biệt:
// → nil = không có giới hạn
// → VD: Min=100, Max=nil → WHERE price >= 100 (không giới hạn trên)
type RangeFilter struct {
	Min any // giá trị nhỏ nhất (nil = không giới hạn dưới)
	Max any // giá trị lớn nhất (nil = không giới hạn trên)
}

// DefaultSpecs — giá trị mặc định an toàn
// Mọi module nên bắt đầu từ đây rồi override field cần thiết
func DefaultSpecs() *Specs {
	return &Specs{
		Keyword:         "",
		KeywordFields:   []string{"name", "title"},
		Relations:       []string{},
		Filters:         map[string]any{},
		RangeFilters:    map[string]RangeFilter{},
		InFilters:       map[string][]any{},
		SelectFields:    []string{},
		Sort:            "id desc",
		Limit:           20,
		Offset:          0,
		UseKeyset:       false,
		CursorField:     "id",
		CursorDirection: "lt",
	}
}
