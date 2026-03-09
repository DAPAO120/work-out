package controller

import (
	"Project001/internal/impl/wechat"
	"Project001/internal/service"

	"github.com/gin-gonic/gin"
)

type WxloginController struct {
}

func NewWxloginController() *WxloginController {
	return &WxloginController{}
}
func (s *WxloginController) TestLogin(c *gin.Context) {
	keyword := c.Query("keyword")

	wechatNew := wechat.New()
	b := wechatNew.TestProg("1")
	aa := TestFun1()
	println(b)
	println(aa)
	a := service.IWxlogin.TestProg(wechatNew, keyword)
	println(a)
	return
}
func TestFun1() string {
	wechatNew := wechat.New()

	b := wechatNew.TestProg("1")
	println(b)

	return b
}
