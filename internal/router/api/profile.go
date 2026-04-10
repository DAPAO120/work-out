package api

import (
	"Project001/internal/controller"
	"Project001/internal/service/profile"
	"Project001/middleware"

	"github.com/gin-gonic/gin"
)

type ProfileRouter struct{}

func (dr *ProfileRouter) InitApiRouter(parent *gin.RouterGroup) {
	// 依赖注入
	profileCtrl := controller.NewProfileController(profile.NewProfileService())
	// 私有路由使用jwt验证
	privateRouter := parent.Group("profileApi")
	privateRouter.Use(middleware.JWTAuth())
	{
		// 个人信息
		privateRouter.GET("/userProfile", profileCtrl.GetUserProfile)

		// 关注相关
		privateRouter.GET("/follow", profileCtrl.Follow)
		privateRouter.GET("/unfollow", profileCtrl.Unfollow)
		privateRouter.GET("/followStatus", profileCtrl.GetFollowStatus)

		// 排行榜
		privateRouter.GET("/rankList", profileCtrl.GetRankList)
		privateRouter.GET("/userRank", profileCtrl.GetUserRank)
	}

}
