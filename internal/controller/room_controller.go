package controller

import (
"errors"
"net/http"
"strconv"
"strings"

"Project001/common/utils"
"Project001/global"
"Project001/internal/service/room"
"Project001/internal/ws"

"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
)

type RoomController struct {
roomService room.IRoomService
}

func NewRoomController(roomService room.IRoomService) *RoomController {
return &RoomController{
roomService: roomService,
}
}

// ---------------- 统一响应 ----------------

func roomOK(c *gin.Context, data interface{}) {
c.JSON(http.StatusOK, gin.H{
"code": 0,
"msg":  "success",
"data": data,
})
}

func roomFail(c *gin.Context, code int, err error) {
c.JSON(http.StatusOK, gin.H{
"code": code,
"msg":  err.Error(),
})
}

// currentUserID 从 JWT 中间件写入的上下文里取用户 ID
func currentUserID(c *gin.Context) (int64, bool) {
userID := c.GetInt64("user_id")
if userID <= 0 {
c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
return 0, false
}
return userID, true
}

// ---------------- 房间接口 ----------------

// CreateRoom 创建房间
func (ctl *RoomController) CreateRoom(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.CreateRoom(c, userID, c.PostForm("title"))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// JoinRoom 加入房间
func (ctl *RoomController) JoinRoom(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
roomNo := strings.TrimSpace(c.PostForm("room_no"))
if roomNo == "" {
roomFail(c, 400, errors.New("请输入房间号"))
return
}
data, err := ctl.roomService.JoinRoom(c, userID, roomNo)
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// GetRoomDetail 房间详情
func (ctl *RoomController) GetRoomDetail(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.GetRoomDetail(c, userID, strings.TrimSpace(c.Query("room_no")))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// ChangeScore 给自己加减分
func (ctl *RoomController) ChangeScore(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
delta, err := strconv.Atoi(strings.TrimSpace(c.PostForm("delta")))
if err != nil {
roomFail(c, 400, errors.New("分值必须是整数"))
return
}
data, err := ctl.roomService.ChangeScore(c, userID, strings.TrimSpace(c.PostForm("room_no")), delta)
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// NextRound 下一轮（房主）
func (ctl *RoomController) NextRound(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.NextRound(c, userID, strings.TrimSpace(c.PostForm("room_no")))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// ResetRoom 重置（房主）
func (ctl *RoomController) ResetRoom(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.ResetRoom(c, userID, strings.TrimSpace(c.PostForm("room_no")))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// KickMember 踢人（房主）
func (ctl *RoomController) KickMember(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
targetID, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("target_user_id")), 10, 64)
if err != nil || targetID <= 0 {
roomFail(c, 400, errors.New("目标用户不合法"))
return
}
data, err := ctl.roomService.KickMember(c, userID, strings.TrimSpace(c.PostForm("room_no")), targetID)
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// DissolveRoom 解散（房主）
func (ctl *RoomController) DissolveRoom(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
if err := ctl.roomService.DissolveRoom(c, userID, strings.TrimSpace(c.PostForm("room_no"))); err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, nil)
}

// LeaveRoom 退出房间
func (ctl *RoomController) LeaveRoom(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.LeaveRoom(c, userID, strings.TrimSpace(c.PostForm("room_no")))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// ScoreLogs 计分流水
func (ctl *RoomController) ScoreLogs(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.ListScoreLogs(c, userID, strings.TrimSpace(c.Query("room_no")))
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// TransferOwner 转让房主（房主）
func (ctl *RoomController) TransferOwner(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
targetID, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("target_user_id")), 10, 64)
if err != nil || targetID <= 0 {
roomFail(c, 400, errors.New("目标用户不合法"))
return
}
data, err := ctl.roomService.TransferOwner(c, userID, strings.TrimSpace(c.PostForm("room_no")), targetID)
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// MyRooms 我参与的房间列表（大厅「我的房间」）
func (ctl *RoomController) MyRooms(c *gin.Context) {
userID, ok := currentUserID(c)
if !ok {
return
}
data, err := ctl.roomService.MyRooms(c, userID)
if err != nil {
roomFail(c, 500, err)
return
}
roomOK(c, data)
}

// ---------------- WebSocket ----------------

var roomUpgrader = websocket.Upgrader{
ReadBufferSize:  1024,
WriteBufferSize: 1024,
// 小程序来源不固定，这里放开同源限制；鉴权靠 token
CheckOrigin: func(r *http.Request) bool { return true },
}

// parseUserIDFromToken 解析 token（允许带或不带 "Bearer " 前缀）
func parseUserIDFromToken(raw string) (int64, error) {
tokenStr := strings.TrimSpace(raw)
if tokenStr == "" {
return 0, errors.New("缺少 token")
}
if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
tokenStr = strings.TrimSpace(tokenStr[7:])
}
claims, err := utils.ParseToken(tokenStr)
if err != nil {
return 0, errors.New("token 无效")
}
return claims.UserID, nil
}

// RoomSocket WebSocket 长连接：实时接收房间状态
// 用法：ws://host/api/roomapi/ws?room_no=123456&token=<jwt>
func (ctl *RoomController) RoomSocket(c *gin.Context) {
roomNo := strings.TrimSpace(c.Query("room_no"))
rawToken := c.Query("token")
if rawToken == "" {
rawToken = c.GetHeader("Authorization")
}

userID, err := parseUserIDFromToken(rawToken)
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
return
}

roomObj, err := ctl.roomService.CheckMember(c, userID, roomNo)
if err != nil {
c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": err.Error()})
return
}

conn, err := roomUpgrader.Upgrade(c.Writer, c.Request, nil)
if err != nil {
global.Log.Error("WebSocket 升级失败: ", err)
return
}

client := ws.Default.Serve(roomObj.RoomNo, userID, conn)

// 连上先推一次当前状态
if detail, err := ctl.roomService.GetRoomDetail(c, userID, roomObj.RoomNo); err == nil {
client.Push(ws.Message{Type: ws.TypeRoomState, Data: detail})
}
}