package controller

import (
	"Project001/global"

	"github.com/gin-gonic/gin"
)

type CommonController struct {
}

func (s *CommonController) TestFun(c *gin.Context) {
	msg := "this is TestFun"
	global.Log.Debug("Debug123")
	global.Log.Warn("Warn123")
	global.Log.Info("Info123")

	println(msg)
	return
}
