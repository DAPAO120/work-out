package router

import "Project001/internal/router/api"

type RouterGroup struct {
	api.CommonRouter
	api.WechatRouter
	api.ArticleRouter
	api.ProfileRouter
	api.RoomRouter
}

var AllRouter = new(RouterGroup)
