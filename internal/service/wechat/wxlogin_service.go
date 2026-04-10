package wechat

import (
	"Project001/internal/model"
	impl "Project001/internal/service/wechat/Impl"
	"context"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

type (
	IWxloginService interface {
		MiniProgram(ctx context.Context) (*miniProgram.MiniProgram, error)
		Login(ctx context.Context, code string) (*model.User, error)
	}
)

func NewWxloginService() IWxloginService {
	return &impl.WxloginServiceImpl{}
}
