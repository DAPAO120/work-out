package controller

import (
	"Project001/common/utils"
	"Project001/global"
	"Project001/internal/model"
	"Project001/internal/service/wechat"
	"bytes"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WechatController struct {
	wechatService        wechat.IWxloginService
	wechatProfileService wechat.IWxProfileService
}

func NewWechatController(wechatService wechat.IWxloginService, wechatProfileService wechat.IWxProfileService) *WechatController {
	return &WechatController{
		wechatService:        wechatService,
		wechatProfileService: wechatProfileService}
}

func (s *WechatController) WXLogin(c *gin.Context) {
	// global.Log.Debug("Headers: %v\n", c.Request.Header)
	// global.Log.Debug("Method: %s, URL: %s\n", c.Request.Method, c.Request.URL.String())
	// if c.Request.Body != nil {
	// 	bodyBytes, _ := io.ReadAll(c.Request.Body)
	// 	// 把读取出的内容重新塞回 Body，供后续 app.Auth.Session 或 c.ShouldBind 使用
	// 	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	// 	global.Log.Debug("Body: %s\n", string(bodyBytes))
	// }

	//微信code
	var code string
	code = c.PostForm("code")
	if code == "" {
		c.JSON(400, gin.H{"error": "未获取到code,请联系管理员"})
		return
	}

	user, err := s.wechatService.Login(c, code)
	if err != nil {
		global.Log.Error("微信根据code获取openid失败,code:" + code)
		return
	}
	//jwt获取token
	token, err := utils.GenerateToken(int64(user.ID), user.IsAdmin)

	result := model.Result{
		Code: 200,
		Msg:  "login success",
		Data: model.LoginData{
			Token:   token,
			UserId:  user.ID,
			Openid:  user.OpenID,
			IsAdmin: user.IsAdmin,
		},
	}
	c.JSON(200, result)
}

func (s *WechatController) GetUserProfileHandler(c *gin.Context) {
	targetIDStr := c.Param("id")
	targetID, _ := strconv.ParseInt(targetIDStr, 10, 64)

	// 当前登录用户（JWT中间件设置）
	currentUserIDVal, _ := c.Get("user_id")
	currentUserID, _ := currentUserIDVal.(int64)

	data, err := s.wechatProfileService.GetUserProfile(currentUserID, targetID)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "user not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": data,
	})
}

func (s *WechatController) GetMyProfileHandler(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"msg": "unauthorized",
		})
		return
	}

	userID := userIDVal.(int64)

	data, err := s.wechatProfileService.GetMyProfile(userID)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "user not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": data,
	})
}

// 上传头像
func (s *WechatController) UploadImage(c *gin.Context) {
	s.wechatProfileService.UploadImage(c)
}

// 更新个人信息
func (s *WechatController) UpdateProfile(c *gin.Context) {
	global.Log.Debug("Headers: %v\n", c.Request.Header)
	global.Log.Debug("Method: %s, URL: %s\n", c.Request.Method, c.Request.URL.String())
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		// 把读取出的内容重新塞回 Body，供后续 app.Auth.Session 或 c.ShouldBind 使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		global.Log.Debug("Body: %s\n", string(bodyBytes))
	}
	s.wechatProfileService.UpdateUserProfile(c)
}
