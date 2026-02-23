package repository

import (
	"errors"
	"fmt"
	"strings"

	"golang-base/global/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ============================================================
// BaseRepository — Generic repository pattern dùng Go Generics
//
// T any: bất kỳ model nào (Product, Order, User...)
// đều dùng được chung các method CRUD, tìm kiếm, phân trang
//
// Tại sao dùng Generics?
// → Viết 1 lần, dùng cho TẤT CẢ model
// → Không cần copy-paste code cho mỗi model
// → Type-safe: compiler bắt lỗi nếu truyền sai type
// ============================================================
type BaseRepository[T any] struct {
	DB *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{DB: db}
}

// ============================================================
// PAGINATE — tự chọn strategy (offset vs keyset) dựa vào specs
// ============================================================
func (r *BaseRepository[T]) Paginate(specs common.Specs) (*common.PaginateResult[T], error) {
	if specs.UseKeyset {
		return r.keysetPaginate(specs)
	}
	return r.offsetPaginate(specs)
}

// ============================================================
// OFFSET PAGINATE
//
// SQL tạo ra:
//
//	SELECT * FROM products ORDER BY id DESC LIMIT 20 OFFSET 40
//
// Phù hợp: Admin panel, danh sách < 100k data
// Không phù hợp: Feed, infinite scroll, data > 100k
// ============================================================
func (r *BaseRepository[T]) offsetPaginate(specs common.Specs) (*common.PaginateResult[T], error) {
	var (
		data  []T
		total int64
	)

	query := r.buildBaseQuery(specs)

	// Đếm tổng trước khi apply limit/offset
	// COUNT(*) trên InnoDB = full table scan nếu không có WHERE clause
	// → Nếu bảng > 1 triệu rows và không filter, sẽ chậm
	// → Giải pháp scale: cache count bằng Redis, hoặc dùng approximate count
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count data failed: %w", err)
	}

	if specs.Sort != "" {
		query = query.Order(specs.Sort)
	}
	if specs.Limit > 0 {
		query = query.Limit(specs.Limit).Offset(specs.Offset)
	}

	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find data failed: %w", err)
	}

	hasMore := false
	if specs.Limit > 0 {
		hasMore = int64(specs.Offset+specs.Limit) < total
	}

	return &common.PaginateResult[T]{
		Data:    data,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

// ============================================================
// KEYSET PAGINATE (Cursor-based)
//
// Thay OFFSET bằng WHERE cursor_field < cursor_value
// → Luôn dùng INDEX → O(log n) thay vì O(n)
//
// Trang 1 (chưa có cursor):
//
//	SELECT * FROM products ORDER BY id DESC LIMIT 21
//
// Trang 2 (cursor = id cuối trang 1 = 9991):
//
//	SELECT * FROM products WHERE id < 9991 ORDER BY id DESC LIMIT 21
//
// Trick: Lấy limit+1 data để biết có trang tiếp không
// → Nếu lấy được 21 data (limit=20) → còn trang tiếp
// → Bỏ record thứ 21, chỉ trả 20
// → KHÔNG cần COUNT(*) → nhanh hơn nhiều
// ============================================================
func (r *BaseRepository[T]) keysetPaginate(specs common.Specs) (*common.PaginateResult[T], error) {
	var data []T

	query := r.buildBaseQuery(specs)

	// Validate cursor field — chống SQL injection
	if specs.CursorField != "" && !validFieldName(specs.CursorField) {
		return nil, fmt.Errorf("invalid cursor field: %s", specs.CursorField)
	}

	// Áp dụng cursor condition — đây là CORE của keyset pagination
	// "lt" (less than) → lấy data CŨ HƠN cursor (DESC order)
	// "gt" (greater than) → lấy data MỚI HƠN cursor (ASC order)
	if specs.CursorValue != nil && specs.CursorField != "" {
		operator := "<"
		if specs.CursorDirection == "gt" {
			operator = ">"
		}
		query = query.Where(
			fmt.Sprintf("%s %s ?", specs.CursorField, operator),
			specs.CursorValue,
		)
	}

	// Sort BẮT BUỘC theo cursor field
	// Nếu sort theo field khác → thứ tự không nhất quán → skip/duplicate data
	if specs.Sort != "" {
		query = query.Order(specs.Sort)
	} else {
		query = query.Order(specs.CursorField + " desc")
	}

	// Lấy limit+1 để detect hasMore KHÔNG cần COUNT(*)
	fetchLimit := specs.Limit + 1
	if err := query.Limit(fetchLimit).Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find data failed: %w", err)
	}

	hasMore := len(data) > specs.Limit
	if hasMore {
		data = data[:specs.Limit] // bỏ record thừa
	}

	// NextCursor = index trong slice, caller tự extract giá trị
	// VD: nextCursor = data[idx].ID
	var nextCursor any
	if hasMore && len(data) > 0 {
		nextCursor = len(data) - 1
	}

	return &common.PaginateResult[T]{
		Data:       data,
		Total:      -1,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// ============================================================
// buildBaseQuery — dựng query chung cho mọi pagination strategy
//
// Pipeline: SELECT fields → Preload → WHERE filters → Range → IN → Keyword
// ============================================================
func (r *BaseRepository[T]) buildBaseQuery(specs common.Specs) *gorm.DB {
	query := r.DB.Model(new(T)) // khởi tạo query từ model

	// Select fields — tránh SELECT *
	// Khi listing 1000 products, nếu mỗi product có description 5KB
	// → SELECT * = 5MB data thừa
	// → SELECT id, name, price, image = chỉ lấy cần thiết
	if len(specs.SelectFields) > 0 {
		query = query.Select(specs.SelectFields)
	}

	// Preload relations (Eager loading)
	// Giải quyết N+1 problem:
	// → Không preload: 1 query products + N query categories = N+1 queries
	// → Có preload: 1 query products + 1 query categories = 2 queries
	for _, relation := range specs.Relations {
		query = query.Preload(relation)
	}

	// Exact match filters — WHERE field = value
	for field, value := range specs.Filters {
		if !validFieldName(field) {
			continue // bỏ qua field không hợp lệ → chống SQL injection
		}
		query = query.Where(field+" = ?", value)
	}

	// Range filters — WHERE field >= min AND field <= max
	// Ứng dụng ecommerce:
	// → Lọc giá: {"price": {Min: 100000, Max: 500000}}
	// → Lọc ngày: {"created_at": {Min: "2024-01-01", Max: "2024-12-31"}}
	// → Lọc rating: {"avg_rating": {Min: 4, Max: nil}} (>=4 sao, không giới hạn trên)
	for field, rf := range specs.RangeFilters {
		if !validFieldName(field) {
			continue
		}
		if rf.Min != nil {
			query = query.Where(field+" >= ?", rf.Min)
		}
		if rf.Max != nil {
			query = query.Where(field+" <= ?", rf.Max)
		}
	}

	// IN filters — WHERE field IN (v1, v2, v3)
	// VD: lọc products thuộc nhiều categories cùng lúc
	// → {"category_id": [1, 2, 3]} → WHERE category_id IN (1, 2, 3)
	for field, values := range specs.InFilters {
		if !validFieldName(field) {
			continue
		}
		if len(values) > 0 {
			query = query.Where(field+" IN ?", values)
		}
	}

	// Keyword search — LIKE '%keyword%'
	//  Lưu ý: %keyword% KHÔNG dùng INDEX → full table scan
	// Scale lớn nên dùng: Full-Text Search (MySQL MATCH AGAINST, PostgreSQL tsvector)
	// hoặc search engine riêng (Elasticsearch, Meilisearch, Typesense)
	if specs.Keyword != "" && len(specs.KeywordFields) > 0 {
		conditions := make([]string, 0, len(specs.KeywordFields))
		args := make([]any, 0, len(specs.KeywordFields))

		for _, field := range specs.KeywordFields {
			if !validFieldName(field) {
				continue
			}
			conditions = append(conditions, field+" LIKE ?")
			args = append(args, "%"+specs.Keyword+"%")
		}

		if len(conditions) > 0 {
			query = query.Where(
				"("+strings.Join(conditions, " OR ")+")",
				args...,
			)
		}
	}

	return query
}

// ============================================================
// CRUD — Create, Read, Update, Delete
// ============================================================

// Create — tạo 1 record mới
func (r *BaseRepository[T]) Create(payload *T) error {
	if err := r.DB.Create(payload).Error; err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	return nil
}

// Insert — batch insert nhiều data cùng lúc
// GORM tự động chia thành chunks nếu quá nhiều
// VD: import 10000 sản phẩm từ CSV
func (r *BaseRepository[T]) Insert(payloads []T) error {
	if len(payloads) == 0 {
		return nil // không có gì để insert → skip
	}
	if err := r.DB.Create(&payloads).Error; err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}
	return nil
}

// InsertInBatches — batch insert với kích thước batch tùy chỉnh
// Khi insert SỐ LƯỢNG LỚN (>1000 data), nên chia batch để:
// → Tránh exceed max_allowed_packet (MySQL mặc định 64MB)
// → Tránh lock table quá lâu → block các query khác
// → Dễ retry nếu 1 batch fail
//
// VD: InsertInBatches(products, 500) → insert 500 data/lần
func (r *BaseRepository[T]) InsertInBatches(payloads []T, batchSize int) error {
	if len(payloads) == 0 {
		return nil
	}
	if err := r.DB.CreateInBatches(&payloads, batchSize).Error; err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}
	return nil
}

// Update — partial update, BỎ QUA zero values
// Dùng cho HTTP PATCH — chỉ update field có giá trị
//
//	Gotcha: Updates(struct) sẽ SKIP zero values!
//
// VD: payload.IsActive = false (bool zero value) → BỊ BỎ QUA
// → Dùng UpdateFields() với map nếu cần update zero values
func (r *BaseRepository[T]) Update(id uint, payload *T) error {
	result := r.DB.Model(new(T)).Where("id = ?", id).Updates(payload)
	if result.Error != nil {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found, ID: %d", id)
	}
	return nil
}

// Save — full update KỂ CẢ zero values
// Dùng cho HTTP PUT — ghi đè toàn bộ
// VD: set IsActive=false (bool zero value) vẫn được lưu
func (r *BaseRepository[T]) Save(payload *T) error {
	if err := r.DB.Save(payload).Error; err != nil {
		return fmt.Errorf("save failed: %w", err)
	}
	return nil
}

// UpdateFields — update chỉ các field chỉ định qua map
// Giải quyết vấn đề zero value của Updates(struct)
// VD: UpdateFields(1, map[string]any{"is_active": false, "stock": 0})
// → false và 0 đều được lưu đúng, không bị skip
func (r *BaseRepository[T]) UpdateFields(id uint, fields map[string]any) error {
	result := r.DB.Model(new(T)).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("update fields failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found, ID: %d", id)
	}
	return nil
}

// BulkUpdateFields — update nhiều data theo điều kiện
// VD: đổi trạng thái tất cả orders quá hạn → "cancelled"
//
//	BulkUpdateFields({"status":"pending"}, {"status":"cancelled"})
//
// conditions: WHERE clause (field → value)
// fields: SET clause (field → new value)
func (r *BaseRepository[T]) BulkUpdateFields(conditions map[string]any, fields map[string]any) (int64, error) {
	query := r.DB.Model(new(T))
	for field, value := range conditions {
		if !validFieldName(field) {
			continue
		}
		query = query.Where(field+" = ?", value)
	}
	result := query.Updates(fields)
	if result.Error != nil {
		return 0, fmt.Errorf("bulk update failed: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// Upsert — tạo mới nếu chưa tồn tại, update nếu đã có
// Dùng SQL: INSERT ... ON DUPLICATE KEY UPDATE (MySQL)
//
//	INSERT ... ON CONFLICT DO UPDATE (PostgreSQL)
//
// Ứng dụng ecommerce rất nhiều:
// → Cart: thêm sản phẩm vào giỏ, nếu đã có thì cập nhật quantity
// → Wishlist: toggle yêu thích
// → Inventory sync: đồng bộ tồn kho từ supplier
// → User preferences: lưu/cập nhật setting
//
// conflictColumns: field(s) xác định record đã tồn tại hay chưa
//
//	VD: []string{"user_id", "product_id"} → nếu đã có cart item của user+product này → update
//
// updateColumns: field(s) cần update nếu conflict
//
//	VD: []string{"quantity", "updated_at"} → chỉ update quantity, giữ nguyên các field khác
func (r *BaseRepository[T]) Upsert(payload *T, conflictColumns []string, updateColumns []string) error {
	columns := make([]clause.Column, len(conflictColumns))
	for i, col := range conflictColumns {
		columns[i] = clause.Column{Name: col}
	}

	doUpdates := clause.AssignmentColumns(updateColumns)

	if err := r.DB.Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: doUpdates,
	}).Create(payload).Error; err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}
	return nil
}

// Delete — xóa mềm hoặc cứng tùy model
// Nếu model có field gorm.DeletedAt → soft delete (đánh dấu deleted_at)
// Nếu không có → hard delete (xóa hẳn khỏi DB)
func (r *BaseRepository[T]) Delete(id uint) error {
	result := r.DB.Delete(new(T), id)
	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found, ID: %d", id)
	}
	return nil
}

// BulkDelete — xóa nhiều data cùng lúc theo mảng IDs
// VD: admin chọn 50 sản phẩm → xóa hàng loạt
func (r *BaseRepository[T]) BulkDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	result := r.DB.Where("id IN ?", ids).Delete(new(T))
	if result.Error != nil {
		return fmt.Errorf("bulk delete failed: %w", result.Error)
	}
	return nil
}

// DeleteByField — xóa theo điều kiện field
// VD: xóa tất cả cart items của user khi checkout xong
//
//	DeleteByField("user_id", 5)
func (r *BaseRepository[T]) DeleteByField(field string, value any) (int64, error) {
	if !validFieldName(field) {
		return 0, fmt.Errorf("invalid field name: %s", field)
	}
	result := r.DB.Where(field+" = ?", value).Delete(new(T))
	if result.Error != nil {
		return 0, fmt.Errorf("delete by field failed: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// RestoreById — khôi phục record đã soft delete
// Set deleted_at = NULL → record "sống lại"
//
// Unscoped() = bỏ qua default WHERE deleted_at IS NULL
// → Có thể thấy và update cả record đã bị soft delete
func (r *BaseRepository[T]) RestoreById(id uint) error {
	result := r.DB.Unscoped().Model(new(T)).
		Where("id = ?", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("restore failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found or not deleted, ID: %d", id)
	}
	return nil
}

// ============================================================
// FIND — Các method tìm kiếm
// ============================================================

// FindById — tìm 1 record theo ID
// Phân biệt rõ: không tìm thấy (ErrRecordNotFound) vs lỗi DB thật
// → Caller có thể xử lý khác nhau: 404 vs 500
func (r *BaseRepository[T]) FindById(id uint, relations []string) (*T, error) {
	var record T
	query := r.DB
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	err := query.First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // nil, nil = không tìm thấy (KHÔNG phải lỗi)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &record, nil
}

// FindByField — tìm 1 record theo 1 field bất kỳ
// VD: FindByField("email", "user@example.com", []string{"Profile"})
// VD: FindByField("slug", "iphone-15-pro", []string{"Category", "Brand"})
func (r *BaseRepository[T]) FindByField(field string, value any, relations []string) (*T, error) {
	if !validFieldName(field) {
		return nil, fmt.Errorf("invalid field name: %s", field)
	}
	var record T
	query := r.DB.Where(field+" = ?", value)
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	err := query.First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &record, nil
}

// FindManyByField — tìm NHIỀU data theo 1 field
// VD: tất cả orders của user_id = 5
// VD: tất cả products có category_id = 3
func (r *BaseRepository[T]) FindManyByField(field string, value any, relations []string, sort string) ([]T, error) {
	if !validFieldName(field) {
		return nil, fmt.Errorf("invalid field name: %s", field)
	}
	var data []T
	query := r.DB.Where(field+" = ?", value)
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	if sort != "" {
		query = query.Order(sort)
	}
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find data failed: %w", err)
	}
	return data, nil
}

// FindByFields — tìm 1 record theo NHIỀU điều kiện AND
// VD: FindByFields(map{"email":"a@b.com", "is_active": true}, []string{})
// → WHERE email = 'a@b.com' AND is_active = true
func (r *BaseRepository[T]) FindByFields(conditions map[string]any, relations []string) (*T, error) {
	var record T
	query := r.DB
	for field, value := range conditions {
		if !validFieldName(field) {
			continue
		}
		query = query.Where(field+" = ?", value)
	}
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	err := query.First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &record, nil
}

// FindWhereIn — lấy data theo mảng giá trị
// VD: FindWhereIn("id", []any{1,2,3}, []string{"Items"}, "created_at desc")
// → WHERE id IN (1, 2, 3) ORDER BY created_at desc
//
// Ứng dụng: lấy danh sách sản phẩm trong giỏ hàng từ mảng product IDs
func (r *BaseRepository[T]) FindWhereIn(field string, values []any, relations []string, sort string) ([]T, error) {
	if !validFieldName(field) {
		return nil, fmt.Errorf("invalid field name: %s", field)
	}
	var data []T
	query := r.DB.Where(field+" IN ?", values)
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	if sort != "" {
		query = query.Order(sort)
	}
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find where in failed: %w", err)
	}
	return data, nil
}

// FindLimit — lấy N data đầu tiên
// VD: "Top 10 sản phẩm bán chạy", "5 đơn hàng gần nhất"
func (r *BaseRepository[T]) FindLimit(limit int, sort string, relations []string) ([]T, error) {
	var data []T
	query := r.DB.Limit(limit)
	for _, rel := range relations {
		query = query.Preload(rel)
	}
	if sort != "" {
		query = query.Order(sort)
	}
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find limit failed: %w", err)
	}
	return data, nil
}

// ============================================================
// UTILITY — Các helper methods
// ============================================================

// ExistsById — check record có tồn tại không
// Dùng SELECT 1 LIMIT 1 thay vì COUNT(*) → nhanh hơn
// COUNT phải đếm TẤT CẢ rows match, EXISTS chỉ cần tìm 1 row
func (r *BaseRepository[T]) ExistsById(id uint) (bool, error) {
	var count int64
	err := r.DB.Model(new(T)).Where("id = ?", id).Limit(1).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("exists check failed: %w", err)
	}
	return count > 0, nil
}

// ExistsByField — check tồn tại theo field bất kỳ
// VD: ExistsByField("email", "user@example.com") → true/false
// Dùng khi: validate unique email, check slug trùng
func (r *BaseRepository[T]) ExistsByField(field string, value any) (bool, error) {
	if !validFieldName(field) {
		return false, fmt.Errorf("invalid field name: %s", field)
	}
	var count int64
	err := r.DB.Model(new(T)).Where(field+" = ?", value).Limit(1).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("exists check failed: %w", err)
	}
	return count > 0, nil
}

// Count — đếm data theo filter
func (r *BaseRepository[T]) Count(filters map[string]any) (int64, error) {
	var count int64
	query := r.DB.Model(new(T))
	for field, value := range filters {
		if !validFieldName(field) {
			continue
		}
		query = query.Where(field+" = ?", value)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}
	return count, nil
}

// ============================================================
// AGGREGATE — Hàm tổng hợp cho dashboard & báo cáo
//
// Ecommerce cần rất nhiều aggregation:
// → Tổng doanh thu hôm nay, tuần này, tháng này
// → Giá trị trung bình mỗi đơn hàng
// → Số lượng tồn kho thấp nhất
// → Tổng số sản phẩm theo category
// ============================================================

// Aggregate — thực hiện hàm tổng hợp SQL (SUM, AVG, COUNT, MIN, MAX)
//
// fn: tên hàm SQL ("SUM", "AVG", "COUNT", "MIN", "MAX")
// field: field cần aggregate
// filters: điều kiện WHERE
//
// VD: Aggregate("SUM", "total_amount", map{"status":"completed"})
//
//	→ SELECT SUM(total_amount) FROM orders WHERE status = 'completed'
//	→ Tổng doanh thu từ đơn hoàn thành
//
// VD: Aggregate("AVG", "price", map{"category_id": 5})
//
//	→ Giá trung bình sản phẩm trong category 5
func (r *BaseRepository[T]) Aggregate(fn string, field string, filters map[string]any) (float64, error) {
	// Validate inputs
	allowedFns := map[string]bool{"SUM": true, "AVG": true, "COUNT": true, "MIN": true, "MAX": true}
	if !allowedFns[strings.ToUpper(fn)] {
		return 0, fmt.Errorf("invalid aggregate function: %s", fn)
	}
	if !validFieldName(field) {
		return 0, fmt.Errorf("invalid field name: %s", field)
	}

	query := r.DB.Model(new(T))
	for f, v := range filters {
		if !validFieldName(f) {
			continue
		}
		query = query.Where(f+" = ?", v)
	}

	var result float64
	err := query.Select(fmt.Sprintf("%s(%s)", strings.ToUpper(fn), field)).Scan(&result).Error
	if err != nil {
		return 0, fmt.Errorf("aggregate failed: %w", err)
	}
	return result, nil
}

// ============================================================
// TRANSACTION — Chạy nhiều operations trong 1 atomic unit
//
// Cực kỳ quan trọng cho ecommerce:
// Khi tạo order phải đảm bảo TẤT CẢ đều thành công hoặc TẤT CẢ đều rollback:
//   1. Tạo order record
//   2. Tạo order_items
//   3. Trừ inventory
//   4. Tạo payment record
// Nếu bước 3 fail → rollback bước 1, 2 → data nhất quán
//
// ACID properties:
// A (Atomicity): tất cả hoặc không gì cả
// C (Consistency): DB luôn ở trạng thái hợp lệ
// I (Isolation): các transaction không ảnh hưởng nhau
// D (Durability): sau commit thì data được persist vĩnh viễn
// ============================================================

// Transaction — chạy callback function trong transaction
// Nếu callback return error → auto rollback
// Nếu callback return nil → auto commit
//
// VD:
//
//	repo.Transaction(func(tx *gorm.DB) error {
//	    if err := tx.Create(&order).Error; err != nil {
//	        return err // → rollback
//	    }
//	    if err := tx.Create(&orderItems).Error; err != nil {
//	        return err // → rollback cả order ở trên
//	    }
//	    return nil // → commit tất cả
//	})
func (r *BaseRepository[T]) Transaction(fn func(tx *gorm.DB) error) error {
	return r.DB.Transaction(fn) // mở transaction
}

// ============================================================
// LOCKING — Pessimistic locking cho concurrent operations
// Khi 2+ users cùng mua sản phẩm cuối cùng trong kho:
// Không có lock: cả 2 đều đọc stock=1 → cả 2 đều mua → stock = -1 ❌
//
// Có FOR UPDATE lock:
// User A: SELECT stock FROM products WHERE id=1 FOR UPDATE → lock row
// User B: SELECT stock FROM products WHERE id=1 FOR UPDATE → BỊ BLOCK, chờ
// User A: UPDATE stock = 0 → COMMIT → unlock
// User B: bây giờ mới đọc được, thấy stock=0 → trả lỗi "hết hàng" ✅
//
//  CHỈ dùng trong Transaction, không dùng ngoài transaction
// ============================================================

// FindByIdForUpdate — tìm record theo ID VÀ lock row
// Dùng trong transaction để đảm bảo không ai khác modify cùng record
//
// VD:
//
//	repo.Transaction(func(tx *gorm.DB) error {
//	    product, err := repo.FindByIdForUpdate(tx, productID)
//	    if product.Stock <= 0 { return errors.New("out of stock") }
//	    product.Stock -= quantity
//	    return tx.Save(product).Error
//	})
func (r *BaseRepository[T]) FindByIdForUpdate(tx *gorm.DB, id uint) (*T, error) {
	var record T
	// Clauses(clause.Locking{Strength: "UPDATE"}) → thêm FOR UPDATE vào query
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // return nil for not found
		}
		return nil, fmt.Errorf("find for update failed: %w", err)
	}
	return &record, nil
}

// FindByFieldForUpdate — tìm và lock theo field bất kỳ
// VD: lock inventory record theo product_id
func (r *BaseRepository[T]) FindByFieldForUpdate(tx *gorm.DB, field string, value any) (*T, error) {
	if !validFieldName(field) {
		return nil, fmt.Errorf("invalid field name: %s", field)
	}
	var record T
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(field+" = ?", value).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // return nil for not found
		}
		return nil, fmt.Errorf("find for update failed: %w", err)
	}
	return &record, nil
}
