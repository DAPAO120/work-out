package service

import (
	"context"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

type (
	IWxlogin interface {
		MiniProgram(ctx context.Context) (*miniProgram.MiniProgram, error)
		TestProg(ctx string) string
		// TestProg2(ctx string) string
	}
)

var localWechat IWxlogin

// func Wechat() IWxlogin {
// 	if localWechat == nil {
// 		panic("implement not found for interface IWechat, forgot register?")
// 	}
// 	return localWechat
// }
// func RegisterWechat(i IWxlogin) {
// 	localWechat = i
// }
