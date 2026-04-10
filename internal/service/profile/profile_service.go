package profile

import (
	"Project001/global"
	"Project001/internal/model"
	impl "Project001/internal/service/profile/Impl"

	"github.com/gin-gonic/gin"
)

type (
	IProfileService interface {
		GetUserProfile(c *gin.Context, userID int64) (*model.UserProfileResponse, error)
		// Follow 关注用户
		Follow(c *gin.Context, userID, targetUserID int64) error
		// Unfollow 取消关注
		Unfollow(c *gin.Context, userID, targetUserID int64) error
		// GetFollowStatus 获取关注状态
		GetFollowStatus(c *gin.Context, userID, targetUserID int64) (bool, error)
		// GetTopUsers 获取排行榜前N名
		GetTopUsers(c *gin.Context, limit int) ([]model.RankUserResponse, error)
		// GetUserRank 获取用户排名
		GetUserRank(c *gin.Context, userID int64) (int64, error)
	}
)

func NewProfileService() IProfileService {
	return &impl.ProfileServiceImpl{
		DB:    global.DB,
		Redis: global.Redis,
	}
}
