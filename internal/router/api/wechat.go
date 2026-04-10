package api

import (
	"Project001/internal/controller"
	"Project001/internal/service/wechat"
	"Project001/middleware"

	"github.com/gin-gonic/gin"
)

type WechatRouter struct{}

func (dr *WechatRouter) InitApiRouter(parent *gin.RouterGroup) {
	//不使用jwt验证
	wxRouter := parent.Group("wechat")
	// 依赖注入
	wechatCtrl := controller.NewWechatController(wechat.NewWxloginService(), wechat.NewWxProfileService())
	{
		wxRouter.POST("login", wechatCtrl.WXLogin)
	}
	// 私有路由使用jwt验证
	privateRouter := parent.Group("wechatapi")
	privateRouter.Use(middleware.JWTAuth())
	{
		privateRouter.POST("myProfile", wechatCtrl.GetMyProfileHandler)
		privateRouter.POST("uploadImage", wechatCtrl.UploadImage)
		privateRouter.POST("updateProfile", wechatCtrl.UpdateProfile)

	}
}
