package controller

import (
	"Project001/global"

	"github.com/gin-gonic/gin"
)

type CommonController struct {
}

func NewCommonController() *CommonController {
	return &CommonController{}
}

func (s *CommonController) TestFun(c *gin.Context) {
	msg := "this is TestFun"
	global.Log.Debug("this is TestFun")
	println(msg)
	return
}
