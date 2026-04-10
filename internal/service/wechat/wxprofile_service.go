package wechat

import (
	"Project001/internal/model"
	impl "Project001/internal/service/wechat/Impl"

	"github.com/gin-gonic/gin"
)

type (
	IWxProfileService interface {
		//查看别人主页（用不上暂时）
		GetUserProfile(currentUserID, targetUserID int64) (*model.UserProfileResp, error)
		//查看个人信息
		GetMyProfile(userID int64) (*model.UserMeResp, error)
		//上传头像
		UploadImage(c *gin.Context)
		//更新个人信息
		UpdateUserProfile(c *gin.Context)
	}
)

func NewWxProfileService() IWxProfileService {
	return &impl.WxProfileServiceImpl{}
}
