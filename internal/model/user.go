package model

import "time"

// type User struct {
// 	ID           int32     `json:"id"`
// 	Email        string    `json:"email"`
// 	PasswordHash string    `json:"password_hash"`
// 	CreatedAt    time.Time `json:"created_at"`
// 	UpdatedAt    time.Time `json:"updated_at"`
// }

type User struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Email        string    `gorm:"column:email;type:varchar(255);not null;uniqueIndex"`
	PasswordHash string    `gorm:"column:password_hash;type:text;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	URLs         []URL     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 用户创建的所有URL
}

func (u *User) TableName() string {
	return "users"
}
