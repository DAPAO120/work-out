package api

import (
"Project001/internal/controller"
"Project001/internal/service/room"
"Project001/middleware"

"github.com/gin-gonic/gin"
)

type RoomRouter struct{}

func (dr *RoomRouter) InitApiRouter(parent *gin.RouterGroup) {
roomCtrl := controller.NewRoomController(room.NewRoomService())

// WebSocket 长连接：自带 token 校验，不走 JWT 中间件
wsRouter := parent.Group("roomapi")
{
wsRouter.GET("/ws", roomCtrl.RoomSocket)
}

// 以下接口需要登录
privateRouter := parent.Group("roomapi")
privateRouter.Use(middleware.JWTAuth())
{
privateRouter.POST("/create", roomCtrl.CreateRoom)
privateRouter.POST("/join", roomCtrl.JoinRoom)
privateRouter.GET("/detail", roomCtrl.GetRoomDetail)
privateRouter.POST("/score", roomCtrl.ChangeScore)
privateRouter.POST("/nextRound", roomCtrl.NextRound)
privateRouter.POST("/reset", roomCtrl.ResetRoom)
privateRouter.POST("/kick", roomCtrl.KickMember)
privateRouter.POST("/dissolve", roomCtrl.DissolveRoom)
privateRouter.POST("/leave", roomCtrl.LeaveRoom)
privateRouter.POST("/transfer", roomCtrl.TransferOwner)
privateRouter.GET("/myRooms", roomCtrl.MyRooms)
privateRouter.GET("/logs", roomCtrl.ScoreLogs)
}
}