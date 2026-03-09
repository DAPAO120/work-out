package wechat

import (
	"context"
	"log"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/response"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/spf13/viper"
)

var MiniProgramApp *miniProgram.MiniProgram

type sWxlogin struct{}

// func init() {
// 	service.RegisterWechat(New())
// }

func New() *sWxlogin {
	s := &sWxlogin{}
	return s
}

func (s *sWxlogin) MiniProgram(ctx context.Context) (*miniProgram.MiniProgram, error) {
	if MiniProgramApp == nil {
		var cache kernel.CacheInterface

		viper.SetConfigName("config")   // 配置文件名称 (无扩展名)
		viper.SetConfigType("yaml")     // 如果配置文件没有扩展名，需要设置类型
		viper.AddConfigPath(".")        // 查找配置文件的路径
		viper.AddConfigPath("./config") // 可以添加多个路径
		viper.AutomaticEnv()            // 读取环境变量
		// 读取配置文件
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Warning: config file not found: %s", err)
		}

		// 获取 Redis 配置，支持从环境变量覆盖
		redisAddrs := viper.GetString("redis.default.address")
		redisDb := viper.GetInt("redis.default.db")
		redisPass := viper.GetString("redis.default.pass")
		// 也可以从环境变量读取
		// if envRedisAddrs := os.Getenv("REDIS_DEFAULT_ADDRESS"); envRedisAddrs != "" {
		// 	redisAddrs = envRedisAddrs
		// }
		// if envRedisDb := os.Getenv("REDIS_DEFAULT_DB"); envRedisDb != "" {
		// 	redisDb = envRedisDb
		// }
		// if envRedisPass := os.Getenv("REDIS_DEFAULT_PASS"); envRedisPass != "" {
		// 	redisPass = envRedisPass
		// }

		if redisAddrs != "" {
			cache = kernel.NewRedisClient(&kernel.UniversalOptions{
				Addrs: []string{redisAddrs},
				//Addrs: []string{
				//	"47.108.182.200:7000",
				//	"47.108.182.200:7001",
				//	"47.108.182.200:7002",
				//},
				DB:       redisDb,
				Password: redisPass,
			})
		}

		// WechatXcxAppId := service.ConfigBase().GetStr(ctx, "wechat_xcx_app_id", "")
		// WechatXcxAppSecret := service.ConfigBase().GetStr(ctx, "wechat_xcx_app_secret", "")
		WechatXcxAppId := "appid"
		WechatXcxAppSecret := "appsec"
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
				// Driver: &testLogDriver.SimpleLogger{},
				Level: "debug",
				File:  "./wechat.log",
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

func (s *sWxlogin) TestProg(ctx string) string {
	a := "1"
	return a
}
