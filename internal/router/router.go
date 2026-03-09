package router

import "Project001/internal/router/admin"

type RouterGroup struct {
	admin.CommonRouter
}

var AllRouter = new(RouterGroup)
