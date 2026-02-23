package user

import (
	"golang-base/internal/model"
	"golang-base/internal/repository"

	"gorm.io/gorm"
)

// CatalogueRepository — struct quản lý thao tác DB với UserCatalogue
type CatalogueRepository struct {
	*repository.BaseRepository[model.UserCatalogue]
}

// NewCatalogueRepository — factory function, map DB connection vào base repo
// Trả về pointer của struct, bên trong đã tự động có sẵn toàn bộ hàm Base
func NewCatalogueRepository(db *gorm.DB) *CatalogueRepository {
	return &CatalogueRepository{
		BaseRepository: repository.NewBaseRepository[model.UserCatalogue](db),
	}
}

// Dưới đây là cách bạn định nghĩa hàm ĐẶC THÙ chỉ có ở Catalogue mà Base chưa có.
// Lưu ý cú pháp receiver: (r *catalogueRepository)
// func (r *catalogueRepository) FindActiveCatalogues() ([]model.UserCatalogue, error) {
// 	  return r.FindManyByField("publish", 2, nil, "")
// }
