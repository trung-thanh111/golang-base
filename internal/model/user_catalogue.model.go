package model

import "time" // dùng cho các field created_at, updated_at

// định nghĩa các field trong bảng user_catalogues
type UserCatalogue struct {

	// go sử dụng   type   struct tag sẽ trả về theo field trên DB thay vì go. gorm mapping struct -> table
	ID          uint      `json:"id"               gorm:"primarykey,autoIncrement"`
	Name        string    `json:"name"             gorm:"not null"`
	Slug        string    `json:"slug"             gorm:"not null,unique"`
	Description string    `json:"description"      gorm:"null"`
	Role        string    `json:"role"             gorm:"not null,default:user"`
	Publish     uint      `json:"publish"          gorm:"not null,default:2"`
	CreatedAt   time.Time `json:"created_at"       gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at"       gorm:"autoUpdateTime"`
}

// khai báo tên bảng trong DB
func (UC *UserCatalogue) TableName() string {
	return "user_catalogues"
}
