package impl

import (
	"Project001/global"
	"Project001/internal/model"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/ArtisanCloud/PowerWeChat/v3/test/testLogDriver"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/response"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

var OfficialAccountApp *officialAccount.OfficialAccount
var MiniProgramApp *miniProgram.MiniProgram

type WxloginServiceImpl struct{}

func (s *WxloginServiceImpl) MiniProgram(ctx context.Context) (*miniProgram.MiniProgram, error) {
	if MiniProgramApp == nil {
		var cache kernel.CacheInterface

		if global.Redis != nil {
			cache = kernel.NewRedisClient(&kernel.UniversalOptions{
				Addrs: []string{global.Config.Redis.Host + ":" + global.Config.Redis.Port},
				//Addrs: []string{
				//	"47.108.182.200:7000",
				//	"47.108.182.200:7001",
				//	"47.108.182.200:7002",
				//},
				Password: global.Config.Redis.Password,
			})
		}

		WechatXcxAppId := global.Config.Wechat.AppId
		WechatXcxAppSecret := global.Config.Wechat.Secret
		app, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
			AppID:        WechatXcxAppId,     // 小程序、公众号或者企业微信的appid
			Secret:       WechatXcxAppSecret, // 商户号 appID
			ResponseType: response.TYPE_MAP,
			//Token:        conf.MiniProgram.MessageToken,
			//AESKey:       conf.MiniProgram.MessageAesKey,
			//
			//AppKey:  conf.MiniProgram.VirtualPayAppKey,
			//OfferID: conf.MiniProgram.VirtualPayOfferID,
			Http: miniProgram.Http{},
			Log: miniProgram.Log{
				Driver: &testLogDriver.SimpleLogger{},
				Level:  "debug",
				File:   "./wechat.log",
			},
			//"sandbox": true,
			Cache:     cache,
			HttpDebug: true,
			Debug:     false,
		})

		MiniProgramApp = app
		return MiniProgramApp, err
	}

	return MiniProgramApp, nil
}
func (s *WxloginServiceImpl) Login(ctx context.Context, code string) (*model.User, error) {
	//调用微信接口换取 OpenID
	app, err := s.MiniProgram(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := app.Auth.Session(ctx, code)
	if err != nil {
		return nil, err
	}
	// 生成随机昵称
	randomStr := generateRandomString(6) // 6位随机字符串，可根据需要调整
	nickname := "新用户" + randomStr

	fileURL := fmt.Sprintf(
		"%s/work-out/%s",
		global.Config.Server.Domain,
		"defaultAvatar.jpg",
	)
	user := model.User{
		OpenID:   rs.OpenID,
		Avatar:   fileURL,
		Nickname: nickname,
	}
	//在数据库中查找用户，若不存在则创建 (Upsert 逻辑)
	result := global.DB.Where(model.User{OpenID: rs.OpenID}).FirstOrCreate(&user)
	if result.Error != nil {
		global.Log.Error("数据库查找用户失败" + result.Error.Error())
		return nil, result.Error
	}

	//更新 UnionID 和最后登录时间
	global.DB.Model(&user).Updates(model.User{
		UnionID:       rs.UnionID,
		LastLoginTime: time.Now(),
	})

	//返回用户信息（后续可在此生成 JWT）
	return &user, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
