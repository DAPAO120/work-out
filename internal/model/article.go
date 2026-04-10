package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID            int64 `gorm:"primaryKey"`
	UserID        int64
	Content       string
	IsFeatured    bool
	FavoriteCount int
	CommentCount  int
	ImageCount    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Title         string
	User          User        `gorm:"foreignKey:UserID"`
	Images        []PostImage `gorm:"foreignKey:PostID"`

	IsFavorite bool `gorm:"-"`
	// 格式化后的时间字段
	FormattedCreatedAt string `gorm:"-"`
	FormattedUpdatedAt string `gorm:"-"`
}

func (Post) TableName() string {
	return "user_post"
}

// AfterFind GORM 钩子，查询后自动格式化时间
func (p *Post) AfterFind(tx *gorm.DB) error {
	p.FormattedCreatedAt = FormatFriendlyTime(p.CreatedAt)
	p.FormattedUpdatedAt = FormatFriendlyTime(p.UpdatedAt)
	return nil
}

type PostImage struct {
	ID       int64 `gorm:"primaryKey"`
	PostID   int64
	ImageURL string
	Sort     int
}

func (PostImage) TableName() string {
	return "post_images"
}

type PostComment struct {
	ID        int64 `gorm:"primaryKey"`
	PostID    int64
	UserID    int64
	ParentID  int64
	Content   string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt

	User   User           `gorm:"foreignKey:UserID"`
	Images []CommentImage `gorm:"foreignKey:CommentID"`
	// 格式化后的时间字段
	FormattedCreatedAt string `gorm:"-"`
}

// AfterFind GORM 钩子，查询后自动格式化时间
func (c *PostComment) AfterFind(tx *gorm.DB) error {
	c.FormattedCreatedAt = FormatFriendlyTime(c.CreatedAt)
	return nil
}

func (PostComment) TableName() string {
	return "post_comment"
}

type CommentImage struct {
	ID        int64 `gorm:"primaryKey"`
	CommentID int64
	ImageURL  string
}

func (CommentImage) TableName() string {
	return "comment_images"
}

type PostFavorite struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64
	PostID    int64
	CreatedAt time.Time
	// 格式化后的时间字段
	FormattedCreatedAt string `gorm:"-"`
}

func (f *PostFavorite) AfterFind(tx *gorm.DB) error {
	f.FormattedCreatedAt = FormatFriendlyTime(f.CreatedAt)
	return nil
}

func (PostFavorite) TableName() string {
	return "post_favorite"
}

type SearchPostsModel struct {
	SearchType string    // "all", "featured", "favorite"
	Keyword    string    // 搜索关键词
	UserID     int64     // 用户ID（收藏搜索时需要）
	CursorTime time.Time // 游标时间
	CursorID   int64     // 游标ID
	PageSize   int       // 每页大小
}
