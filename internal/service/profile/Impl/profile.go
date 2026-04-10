package impl

import (
	"Project001/internal/dao"
	"Project001/internal/model"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type ProfileServiceImpl struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (s *ProfileServiceImpl) GetUserProfile(c *gin.Context, userID int64) (*model.UserProfileResponse, error) {
	user, err := dao.GetUserByID(s.DB, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 获取统计数据
	var followCount, fansCount, postCount int64

	s.DB.Model(&model.UserFollow{}).Where("user_id = ?", userID).Count(&followCount)
	s.DB.Model(&model.UserFollow{}).Where("follow_user_id = ?", userID).Count(&fansCount)
	s.DB.Model(&model.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&postCount)

	return &model.UserProfileResponse{
		ID:          int64(user.ID),
		OpenID:      user.OpenID,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Bio:         user.Bio,
		Background:  user.Background,
		Gender:      user.Gender,
		IsAdmin:     user.IsAdmin,
		FollowCount: followCount,
		FansCount:   fansCount,
		PostCount:   postCount,
	}, nil
}
func (s *ProfileServiceImpl) Follow(c *gin.Context, userID, targetUserID int64) error {
	if userID == targetUserID {
		return errors.New("不能关注自己")
	}

	// 检查用户是否存在
	user, err := dao.GetUserByID(s.DB, targetUserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 检查是否已关注
	exists, err := dao.CheckFollowExists(s.DB, userID, targetUserID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("已经关注过了")
	}

	return dao.CreateFollow(s.DB, userID, targetUserID)
}

func (s *ProfileServiceImpl) Unfollow(c *gin.Context, userID, targetUserID int64) error {
	if userID == targetUserID {
		return errors.New("不能取消关注自己")
	}

	exists, err := dao.CheckFollowExists(s.DB, userID, targetUserID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("尚未关注")
	}

	return dao.DeleteFollow(s.DB, userID, targetUserID)
}

func (s *ProfileServiceImpl) GetFollowStatus(c *gin.Context, userID, targetUserID int64) (bool, error) {
	return dao.CheckFollowExists(s.DB, userID, targetUserID)
}

func (s *ProfileServiceImpl) GetTopUsers(c *gin.Context, limit int) ([]model.RankUserResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rankList, err := dao.GetUserRankList(s.DB, limit)
	if err != nil {
		return nil, err
	}

	var result []model.RankUserResponse
	for i, item := range rankList {
		result = append(result, model.RankUserResponse{
			UserID:        item.UserID,
			Nickname:      item.Nickname,
			Avatar:        item.Avatar,
			PostCount:     item.PostCount,
			FeaturedCount: item.FeaturedCount,
			Rank:          int64(i + 1),
			Score:         item.Score,
		})
	}

	return result, nil
}

func (s *ProfileServiceImpl) GetUserRank(c *gin.Context, userID int64) (int64, error) {
	rank, err := dao.GetUserRankByID(s.DB, userID)
	if err != nil {
		return 0, err
	}
	return rank, nil
}
