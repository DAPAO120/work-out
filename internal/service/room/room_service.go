package room

import (
"Project001/global"
"Project001/internal/model"
impl "Project001/internal/service/room/Impl"

"github.com/gin-gonic/gin"
)

type (
IRoomService interface {
// CreateRoom 创建房间（创建者即房主）
CreateRoom(c *gin.Context, userID int64, title string) (*model.RoomDetailVO, error)
// JoinRoom 加入房间
JoinRoom(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error)
// GetRoomDetail 获取房间详情（需是房间成员）
GetRoomDetail(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error)
// MyRooms 我参与的所有进行中房间（大厅「我的房间」列表）
MyRooms(c *gin.Context, userID int64) ([]model.RoomBriefVO, error)
// ChangeScore 给自己加减分
ChangeScore(c *gin.Context, userID int64, roomNo string, delta int) (*model.RoomDetailVO, error)
// NextRound 开启下一轮（房主）
NextRound(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error)
// ResetRoom 重置房间分数与轮次（房主）
ResetRoom(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error)
// TransferOwner 转让房主给房间内其他成员（房主）
TransferOwner(c *gin.Context, userID int64, roomNo string, targetUserID int64) (*model.RoomDetailVO, error)
// KickMember 踢人（房主）
KickMember(c *gin.Context, userID int64, roomNo string, targetUserID int64) (*model.RoomDetailVO, error)
// DissolveRoom 解散房间（房主）
DissolveRoom(c *gin.Context, userID int64, roomNo string) error
// LeaveRoom 退出房间：
//   普通成员     -> 直接退出
//   房主 + 还有别人 -> 自动转让给最早加入的成员后退出
//   房主 + 只剩自己 -> 自动解散房间
LeaveRoom(c *gin.Context, userID int64, roomNo string) (*model.LeaveRoomResult, error)
// ListScoreLogs 计分流水
ListScoreLogs(c *gin.Context, userID int64, roomNo string) ([]model.RoomScoreLogVO, error)
// CheckMember 校验某人是否是房间有效成员（WebSocket 握手用）
CheckMember(c *gin.Context, userID int64, roomNo string) (*model.Room, error)
}
)

func NewRoomService() IRoomService {
return &impl.RoomServiceImpl{
DB:    global.DB,
Redis: global.Redis,
}
}