package initialize

import (
	"Project001/global"
	"Project001/internal/router"

	"github.com/gin-gonic/gin"
)

func routerInit() *gin.Engine {
	r := gin.Default()
	allRouter := router.AllRouter

	// 链路追踪日志中间件
	r.Use(global.Log.LogrusGinMiddleware())

	// api
	api := r.Group("/api")
	{
		allRouter.CommonRouter.InitApiRouter(api)
		allRouter.WechatRouter.InitApiRouter(api)
		allRouter.ArticleRouter.InitApiRouter(api)
		allRouter.ProfileRouter.InitApiRouter(api)
	}
	return r
}
