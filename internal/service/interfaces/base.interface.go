// ============================================================
// Tầng Controller (HTTP Layer) sẽ nhận HTTP Request → parse ra DTO / Struct → Gửi vào Service.
// Do đó, Service chỉ nhận Go Structs (T hoặc DTO) và trả về (Data, error).
//
// Lợi ích:
// 1. Dễ test: Bạn KHÔNG CẦN phải mock nguyên cái HTTP Request / framework Gin để test business logic.
// 2. Tái sử dụng: Nếu mai mốt gọi Service này từ Cronjob (không có HTTP), gRPC, hoặc CLI → Vẫn chạy 100%.
// ============================================================
package interfaces

import (
	"golang-base/global/common"
)

// IHook — Định nghĩa các điểm neo (hooks) để module con can thiệp vào lồng xử lý
// Nếu module con không implement, mặc định sẽ chạy logic rỗng (trả về nil).
type IHook[T any] interface {
	BeforeCreate(payload *T) error
	AfterCreate(payload *T) error

	BeforeUpdate(id uint, payload *T) error
	AfterUpdate(id uint, payload *T) error

	BeforeDelete(id uint) error
	AfterDelete(id uint) error
}

// IBaseService — Định nghĩa các hành vi nghiệp vụ dùng chung
type IBaseService[T any] interface {
	IHook[T] // Kế thừa toàn bộ hook

	Create(payload *T) error
	BulkCreate(payloads []T) error

	Update(id uint, payload *T) error
	BulkUpdate(conditions map[string]any, payload map[string]any) (int64, error)

	Delete(id uint) error

	FindById(id uint) (*T, error)
	Paginate(specs common.Specs) (*common.PaginateResult[T], error)
}
