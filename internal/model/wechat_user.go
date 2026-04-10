package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint   `gorm:"primaryKey"`
	OpenID        string `gorm:"type:varchar(128);uniqueIndex;not null;column:open_id"`
	UnionID       string `gorm:"type:varchar(128);index;column:union_id"`
	Nickname      string `gorm:"type:varchar(100);column:nickname"`
	Avatar        string `gorm:"column:avatar_url"`
	IsAdmin       bool   `gorm:"column:is_admin;default:false"`
	Bio           string
	Background    string
	Gender        int8           `gorm:"default:0;column:gender"`
	LastLoginTime time.Time      `gorm:"column:last_login_time"`
	CreatedTime   time.Time      `gorm:"column:created_time"`
	UpdatedTime   time.Time      `gorm:"column:updated_time"`
	DeletedTime   gorm.DeletedAt `gorm:"index;column:deleted_time"`
}

func (User) TableName() string {
	return "users"
}

type UserFollow struct {
	ID           uint `gorm:"primaryKey"`
	UserID       int64
	FollowUserID int64
	CreatedAt    time.Time
}

func (UserFollow) TableName() string {
	return "user_follow"
}
