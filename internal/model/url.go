package model

import "time"

// type URL struct {
// 	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
// 	OriginalURL string    `gorm:"column:original_url;type:text;not null"`
// 	ShortCode   string    `gorm:"column:short_code;type:text;not null;size:100;uniqueIndex"`
// 	IsCustom    bool      `gorm:"column:is_custom;not null;default:false"`
// 	Views       int32     `json:"views"`
// 	ExpiredAt   time.Time `gorm:"column:expired_at;type:timestamp;not null"`
// 	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
// }

type URL struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      uint64    `gorm:"column:user_id;not null"` // 新增：关联用户ID
	OriginalURL string    `gorm:"column:original_url;type:text;not null"`
	ShortCode   string    `gorm:"column:short_code;type:varchar(100);not null;uniqueIndex"`
	IsCustom    bool      `gorm:"column:is_custom;not null;default:false"`
	Views       int32     `gorm:"column:views;not null;default:0"` // 修改：添加默认值0
	ExpiredAt   time.Time `gorm:"column:expired_at;type:timestamp;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
	User        *User     `gorm:"foreignKey:UserID"` // 新增：关联用户对象
}

func (u *URL) TableName() string {
	return "urls"
}

func (u *URL) IsExpired() bool {
	return u.ExpiredAt.Before(time.Now())
}

func (u *URL) Renew(duration time.Duration) {
	u.ExpiredAt = time.Now().Add(duration)
}
