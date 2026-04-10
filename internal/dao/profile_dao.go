package dao

import (
	"Project001/internal/model"
	"errors"

	"gorm.io/gorm"
)

func GetUserByID(db *gorm.DB, userID int64) (*model.User, error) {
	var user model.User
	err := db.Where("id = ? AND deleted_time IS NULL", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByOpenID(db *gorm.DB, openID string) (*model.User, error) {
	var user model.User
	err := db.Where("open_id = ? AND deleted_time IS NULL", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, user *model.User) error {
	return db.Save(user).Error
}

func CheckFollowExists(db *gorm.DB, userID, followUserID int64) (bool, error) {
	var count int64
	err := db.Model(&model.UserFollow{}).
		Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
		Count(&count).Error
	return count > 0, err
}

// CreateFollow 创建关注
func CreateFollow(db *gorm.DB, userID, followUserID int64) error {
	follow := &model.UserFollow{
		UserID:       userID,
		FollowUserID: followUserID,
	}
	return db.Create(follow).Error
}

// DeleteFollow 取消关注
func DeleteFollow(db *gorm.DB, userID, followUserID int64) error {
	return db.Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
		Delete(&model.UserFollow{}).Error
}

// GetFollowCount 获取关注数
func GetFollowCount(db *gorm.DB, userID int64) (int64, error) {
	var count int64
	err := db.Model(&model.UserFollow{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetFansCount 获取粉丝数
func GetFansCount(db *gorm.DB, userID int64) (int64, error) {
	var count int64
	err := db.Model(&model.UserFollow{}).Where("follow_user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetFollowings 获取用户的关注列表
func GetFollowings(db *gorm.DB, userID int64, cursorID int64, pageSize int) ([]model.UserFollow, error) {
	var follows []model.UserFollow
	query := db.Where("user_id = ?", userID).Order("id DESC").Limit(pageSize)
	if cursorID > 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Find(&follows).Error
	return follows, err
}

// GetFans 获取用户的粉丝列表
func GetFans(db *gorm.DB, userID int64, cursorID int64, pageSize int) ([]model.UserFollow, error) {
	var fans []model.UserFollow
	query := db.Where("follow_user_id = ?", userID).Order("id DESC").Limit(pageSize)
	if cursorID > 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Find(&fans).Error
	return fans, err
}

// 排序规则：先按精品数降序，再按帖子数降序，只返回post_count > 0的用户
func GetUserRankList(db *gorm.DB, limit int) ([]model.RankItem, error) {
	var ranks []model.RankItem

	// 使用子查询分别统计总帖数和精品帖数
	err := db.Table("users").
		Select(`
			users.id as user_id,
			users.nickname,
			users.avatar_url as avatar,
			COALESCE(post_stats.post_count, 0) as post_count,
			COALESCE(featured_stats.featured_count, 0) as featured_count
		`).
		Joins(`
			LEFT JOIN (
				SELECT user_id, COUNT(*) as post_count 
				FROM user_post 
				WHERE deleted_at IS NULL 
				GROUP BY user_id
			) as post_stats ON users.id = post_stats.user_id
		`).
		Joins(`
			LEFT JOIN (
				SELECT user_id, COUNT(*) as featured_count 
				FROM user_post 
				WHERE is_featured = true AND deleted_at IS NULL 
				GROUP BY user_id
			) as featured_stats ON users.id = featured_stats.user_id
		`).
		Where("users.deleted_time IS NULL").
		Where("COALESCE(post_stats.post_count, 0) > 0"). // 只返回有帖子的用户
		Order("featured_count DESC, post_count DESC").
		Limit(limit).
		Scan(&ranks).Error

	if err != nil {
		return nil, err
	}

	// 计算综合分数（可选）
	for i := range ranks {
		// 综合分数 = 精品数*10 + 总帖数
		ranks[i].Score = ranks[i].FeaturedCount*10 + ranks[i].PostCount
	}

	return ranks, nil
}

// GetUserRankListWithWeight 带权重的排行榜
// postWeight: 帖子权重，featuredWeight: 精品权重
func GetUserRankListWithWeight(db *gorm.DB, limit int, postWeight, featuredWeight int) ([]model.RankItem, error) {
	var ranks []model.RankItem

	err := db.Table("users").
		Select(`
			users.id as user_id,
			users.nickname,
			users.avatar_url as avatar,
			COALESCE(post_stats.post_count, 0) as post_count,
			COALESCE(featured_stats.featured_count, 0) as featured_count,
			(COALESCE(featured_stats.featured_count, 0) * ? + COALESCE(post_stats.post_count, 0) * ?) as score
		`, featuredWeight, postWeight).
		Joins(`
			LEFT JOIN (
				SELECT user_id, COUNT(*) as post_count 
				FROM user_post 
				WHERE deleted_at IS NULL 
				GROUP BY user_id
			) as post_stats ON users.id = post_stats.user_id
		`).
		Joins(`
			LEFT JOIN (
				SELECT user_id, COUNT(*) as featured_count 
				FROM user_post 
				WHERE is_featured = true AND deleted_at IS NULL 
				GROUP BY user_id
			) as featured_stats ON users.id = featured_stats.user_id
		`).
		Where("users.deleted_time IS NULL").
		Where("COALESCE(post_stats.post_count, 0) > 0").
		Order("score DESC").
		Limit(limit).
		Scan(&ranks).Error

	return ranks, err
}

// GetUserRankByID 获取指定用户的排名
func GetUserRankByID(db *gorm.DB, userID int64) (int64, error) {
	var rank int64

	// 使用子查询计算排名
	err := db.Raw(`
		SELECT ranking FROM (
			SELECT 
				user_id,
				ROW_NUMBER() OVER (ORDER BY featured_count DESC, post_count DESC) as ranking
			FROM (
				SELECT 
					users.id as user_id,
					COALESCE(post_stats.post_count, 0) as post_count,
					COALESCE(featured_stats.featured_count, 0) as featured_count
				FROM users
				LEFT JOIN (
					SELECT user_id, COUNT(*) as post_count 
					FROM user_post 
					WHERE deleted_at IS NULL 
					GROUP BY user_id
				) as post_stats ON users.id = post_stats.user_id
				LEFT JOIN (
					SELECT user_id, COUNT(*) as featured_count 
					FROM user_post 
					WHERE is_featured = true AND deleted_at IS NULL 
					GROUP BY user_id
				) as featured_stats ON users.id = featured_stats.user_id
				WHERE users.deleted_time IS NULL
					AND COALESCE(post_stats.post_count, 0) > 0
			) as stats
		) as ranked
		WHERE user_id = ?
	`, userID).Scan(&rank).Error

	return rank, err
}
