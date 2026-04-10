package controller

import (
	"Project001/internal/service/profile"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileService profile.IProfileService
}

func NewProfileController(profileService profile.IProfileService) *ProfileController {
	return &ProfileController{
		profileService: profileService,
	}
}

// GetUserProfile 获取用户个人信息
func (c *ProfileController) GetUserProfile(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "无效的用户ID"})
		return
	}

	profile, err := c.profileService.GetUserProfile(ctx, userID)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "success", "data": profile})
}

// Follow 关注用户
func (c *ProfileController) Follow(ctx *gin.Context) {
	// 从上下文中获取当前用户ID（假设已经通过中间件设置）
	userID := ctx.GetInt64("user_id")

	targetUserIDStr := ctx.Query("target_user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "无效的用户ID"})
		return
	}

	err = c.profileService.Follow(ctx, userID, targetUserID)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "关注成功"})
}

// Unfollow 取消关注
func (c *ProfileController) Unfollow(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	targetUserIDStr := ctx.Query("target_user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "无效的用户ID"})
		return
	}

	err = c.profileService.Unfollow(ctx, userID, targetUserID)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "取消关注成功"})
}

// GetRankList 获取排行榜
func (c *ProfileController) GetRankList(ctx *gin.Context) {
	limit := 10
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	rankList, err := c.profileService.GetTopUsers(ctx, limit)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "success", "data": rankList})
}

// GetUserRank 获取当前用户排名
func (c *ProfileController) GetUserRank(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	rank, err := c.profileService.GetUserRank(ctx, userID)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "success", "data": gin.H{"rank": rank}})
}

// GetFollowStatus 获取关注状态
func (c *ProfileController) GetFollowStatus(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	targetUserIDStr := ctx.Query("target_user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "无效的用户ID"})
		return
	}

	isFollowing, err := c.profileService.GetFollowStatus(ctx, userID, targetUserID)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"code": 0, "msg": "success", "data": gin.H{"is_following": isFollowing}})
}
