package model

import "time"

type User struct {
	ID        uint      `json:"id"           gorm:"primarykey,autoIncrement"`
	Name      string    `json:"name"         gorm:"not null"`
	Email     string    `json:"email"        gorm:"not null"`
	Password  string    `json:"password"     gorm:"not null"`
	CreatedAt time.Time `json:"created_at"   gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at"   gorm:"autoUpdateTime"`
}

// khai báo tên bảng trong DB
func (U *User) TableName() string {
	return "users"
}
