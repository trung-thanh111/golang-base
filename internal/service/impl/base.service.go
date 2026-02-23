// BASE SERVICE — business logic dùng chung
// Nhiệm vụ:
// - Gọi repository để lấy/ghi Data.
// - KHÔNG biết Database là MySQL hay Mongo (đã có repo lo).
// - KHÔNG biết HTTP Request/Response (đã có Controller lo).
package impl

import (
	"golang-base/global/common"
	r "golang-base/internal/repository"
	"golang-base/internal/service/interfaces"
	si "golang-base/internal/service/interfaces"
)

// BaseService — implement interfaces.IBaseService[T]
type BaseService[T any] struct {
	br   *r.BaseRepository[T]
	hook si.IHook[T] // Instance chứa các hàm hook của module con
}

// NewBaseService — inject cả Repository lẫn Hook
// Hook thường chính là con trỏ của module service con.
// VD: &UserCatalogueService{}
func NewBaseService[T any](br *r.BaseRepository[T], hook interfaces.IHook[T]) *BaseService[T] {
	return &BaseService[T]{
		br:   br,
		hook: hook,
	}
}

// ============================================================
// SINGLE ACTIONS — Có áp dụng Hook Pipeline
// ============================================================

func (s *BaseService[T]) Create(payload *T) error {
	// 1. Hook Before
	if s.hook != nil {
		if err := s.hook.BeforeCreate(payload); err != nil {
			return err // Dừng sớm nều validate/logic trước khi tạo fail
		}
	}

	// 2. Action chính
	if err := s.br.Create(payload); err != nil {
		return err
	}

	// 3. Hook After
	if s.hook != nil {
		if err := s.hook.AfterCreate(payload); err != nil {
			return err // Báo lỗi nhưng record đã được insert (trừ khi quấn trong Transaction)
		}
	}
	return nil
}

func (s *BaseService[T]) Update(id uint, payload *T) error {
	if s.hook != nil {
		if err := s.hook.BeforeUpdate(id, payload); err != nil {
			return err
		}
	}

	if err := s.br.Update(id, payload); err != nil {
		return err
	}

	if s.hook != nil {
		return s.hook.AfterUpdate(id, payload)
	}
	return nil
}

func (s *BaseService[T]) Delete(id uint) error {
	if s.hook != nil {
		if err := s.hook.BeforeDelete(id); err != nil {
			return err
		}
	}

	if err := s.br.Delete(id); err != nil {
		return err
	}

	if s.hook != nil {
		return s.hook.AfterDelete(id)
	}
	return nil
}

// ============================================================
// BULK ACTIONS — Dành riêng cho hiệu suất
// Không nên chạy bulk qua Single Hook vì sẽ lặp vòng for rất chậm.
// ============================================================

func (s *BaseService[T]) BulkCreate(payloads []T) error {
	// Delegate việc batch size (vd: 500) xuống cho DB xử lý an toàn
	return s.br.InsertInBatches(payloads, 500)
}

func (s *BaseService[T]) BulkUpdate(conditions map[string]any, payload map[string]any) (int64, error) {
	return s.br.BulkUpdateFields(conditions, payload)
}

func (s *BaseService[T]) FindById(id uint) (*T, error) {
	return s.br.FindById(id, nil)
}

func (s *BaseService[T]) Paginate(specs common.Specs) (*common.PaginateResult[T], error) {
	return s.br.Paginate(specs)
}
