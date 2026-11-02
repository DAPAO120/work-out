package impl

import (
"errors"
"fmt"
"math/rand"
"sort"
"time"

"Project001/internal/dao"
"Project001/internal/model"
"Project001/internal/ws"

"github.com/gin-gonic/gin"
"github.com/go-redis/redis"
"gorm.io/gorm"
)

type RoomServiceImpl struct {
DB    *gorm.DB
Redis *redis.Client
}

const (
roomStatusClosed int8 = 0 // 已解散
roomStatusActive int8 = 1 // 进行中

maxScoreDelta    = 100000 // 单次加减分上限（绝对值）
defaultScoreLogs = 50     // 默认返回的流水条数
)

// ---------------- 内部工具 ----------------

// genRoomNo 生成不重复的 6 位数字房间号
func genRoomNo(db *gorm.DB) (string, error) {
for i := 0; i < 30; i++ {
no := fmt.Sprintf("%06d", rand.Intn(1000000))
exists, err := dao.RoomNoExists(db, no)
if err != nil {
return "", err
}
if !exists {
return no, nil
}
}
return "", errors.New("房间号生成失败，请稍后重试")
}

func formatTime(t time.Time) string {
if t.IsZero() {
return ""
}
return t.Format("2006-01-02 15:04:05")
}

// mustRoom 取房间并校验状态
func (s *RoomServiceImpl) mustRoom(roomNo string) (*model.Room, error) {
if roomNo == "" {
return nil, errors.New("房间号不能为空")
}
room, err := dao.GetRoomByNo(s.DB, roomNo)
if err != nil {
return nil, err
}
if room == nil {
return nil, errors.New("房间不存在")
}
if room.Status != roomStatusActive {
return nil, errors.New("房间不存在或已解散")
}
return room, nil
}

// mustMember 校验用户是房间内的有效成员
func (s *RoomServiceImpl) mustMember(roomID, userID int64) (*model.RoomMember, error) {
member, err := dao.GetMember(s.DB, roomID, userID)
if err != nil {
return nil, err
}
if member == nil || member.KickedAt != nil {
return nil, errors.New("你不在这个房间里")
}
return member, nil
}

// mustOwner 校验用户是房主
func (s *RoomServiceImpl) mustOwner(room *model.Room, userID int64) error {
if room.OwnerID != userID {
return errors.New("只有房主可以进行该操作")
}
return nil
}

// buildDetail 组装房间详情（viewerID 用于标记 is_self）
func (s *RoomServiceImpl) buildDetail(room *model.Room, viewerID int64) (*model.RoomDetailVO, error) {
members, err := dao.ListMembers(s.DB, room.ID)
if err != nil {
return nil, err
}

// 昵称/头像以 users 表为准：用户改过资料后，房间内立即生效
ids := make([]int64, 0, len(members))
for _, m := range members {
ids = append(ids, m.UserID)
}
userMap := make(map[uint]model.User, len(ids))
if users, uerr := dao.GetUsersByIDs(s.DB, ids); uerr == nil {
for _, u := range users {
userMap[u.ID] = u
}
}

vos := make([]model.RoomMemberVO, 0, len(members))
for _, m := range members {
nickname, avatar := m.Nickname, m.Avatar
if u, ok := userMap[uint(m.UserID)]; ok {
if u.Nickname != "" {
nickname = u.Nickname
}
if u.Avatar != "" {
avatar = u.Avatar
}
}
vos = append(vos, model.RoomMemberVO{
UserID:     m.UserID,
Nickname:   nickname,
Avatar:     avatar,
TotalScore: m.TotalScore,
IsOwner:    m.IsOwner,
IsSelf:     m.UserID == viewerID,
JoinedAt:   formatTime(m.JoinedAt),
})
}

// 按总分倒序；同分按加入时间正序，保证顺序稳定
sort.SliceStable(vos, func(i, j int) bool {
if vos[i].TotalScore != vos[j].TotalScore {
return vos[i].TotalScore > vos[j].TotalScore
}
return vos[i].JoinedAt < vos[j].JoinedAt
})

return &model.RoomDetailVO{
RoomNo:    room.RoomNo,
Title:     room.Title,
Round:     room.Round,
OwnerID:   room.OwnerID,
IsOwner:   room.OwnerID == viewerID,
Status:    room.Status,
Online:    ws.Default.OnlineCount(room.RoomNo),
Members:   vos,
CreatedAt: formatTime(room.CreatedAt),
}, nil
}
// pushState 广播房间最新状态给房间内所有在线连接
// 注意：广播内容对所有接收者一致，is_owner 在这里没有意义，客户端用 owner_id 自行判断
func (s *RoomServiceImpl) pushState(roomNo string) {
room, err := dao.GetRoomByNo(s.DB, roomNo)
if err != nil || room == nil {
return
}
detail, err := s.buildDetail(room, 0)
if err != nil {
return
}
ws.Default.Broadcast(roomNo, ws.Message{Type: ws.TypeRoomState, Data: detail})
}

// ---------------- 对外方法 ----------------

// CreateRoom 创建房间
func (s *RoomServiceImpl) CreateRoom(c *gin.Context, userID int64, title string) (*model.RoomDetailVO, error) {
user, err := dao.GetUserByID(s.DB, userID)
if err != nil {
return nil, err
}
if user == nil {
return nil, errors.New("用户不存在")
}

roomNo, err := genRoomNo(s.DB)
if err != nil {
return nil, err
}

now := time.Now()
room := &model.Room{
RoomNo:    roomNo,
OwnerID:   userID,
Title:     title,
Status:    roomStatusActive,
Round:     1,
CreatedAt: now,
UpdatedAt: now,
}
if err := dao.CreateRoom(s.DB, room); err != nil {
return nil, err
}

member := &model.RoomMember{
RoomID:     room.ID,
UserID:     userID,
Nickname:   user.Nickname,
Avatar:     user.Avatar,
TotalScore: 0,
IsOwner:    true,
JoinedAt:   now,
}
if err := dao.AddMember(s.DB, member); err != nil {
return nil, err
}

return s.buildDetail(room, userID)
}

// JoinRoom 加入房间
func (s *RoomServiceImpl) JoinRoom(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}

user, err := dao.GetUserByID(s.DB, userID)
if err != nil {
return nil, err
}
if user == nil {
return nil, errors.New("用户不存在")
}

member, err := dao.GetMember(s.DB, room.ID, userID)
if err != nil {
return nil, err
}

if member != nil {
if member.KickedAt != nil {
return nil, errors.New("你已被房主移出该房间")
}
// 已在房间内：刷新一次昵称头像
if err := dao.UpdateMemberProfile(s.DB, member.ID, user.Nickname, user.Avatar); err != nil {
return nil, err
}
} else {
newMember := &model.RoomMember{
RoomID:     room.ID,
UserID:     userID,
Nickname:   user.Nickname,
Avatar:     user.Avatar,
TotalScore: 0,
IsOwner:    false,
JoinedAt:   time.Now(),
}
if err := dao.AddMember(s.DB, newMember); err != nil {
return nil, err
}
}

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// GetRoomDetail 房间详情
func (s *RoomServiceImpl) GetRoomDetail(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if _, err := s.mustMember(room.ID, userID); err != nil {
return nil, err
}
return s.buildDetail(room, userID)
}

// ChangeScore 给自己加减分
func (s *RoomServiceImpl) ChangeScore(c *gin.Context, userID int64, roomNo string, delta int) (*model.RoomDetailVO, error) {
if delta == 0 {
return nil, errors.New("分数不能为 0")
}
if delta > maxScoreDelta || delta < -maxScoreDelta {
return nil, fmt.Errorf("单次加减分不能超过 %d", maxScoreDelta)
}

room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
member, err := s.mustMember(room.ID, userID)
if err != nil {
return nil, err
}

newScore := member.TotalScore + delta
if err := dao.UpdateMemberScore(s.DB, member.ID, newScore); err != nil {
return nil, err
}

record := &model.RoomScore{
RoomID:     room.ID,
UserID:     userID,
Round:      room.Round,
Delta:      delta,
ScoreAfter: newScore,
CreatedAt:  time.Now(),
}
if err := dao.InsertScoreRecord(s.DB, record); err != nil {
return nil, err
}
_ = dao.TouchRoom(s.DB, room.ID)

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// NextRound 开启下一轮（房主）
func (s *RoomServiceImpl) NextRound(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if err := s.mustOwner(room, userID); err != nil {
return nil, err
}
if err := dao.UpdateRoomRound(s.DB, room.ID, room.Round+1); err != nil {
return nil, err
}
room.Round = room.Round + 1

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// ResetRoom 重置分数与轮次（房主）
func (s *RoomServiceImpl) ResetRoom(c *gin.Context, userID int64, roomNo string) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if err := s.mustOwner(room, userID); err != nil {
return nil, err
}
if err := dao.ResetRoomScores(s.DB, room.ID); err != nil {
return nil, err
}
room.Round = 1

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// KickMember 踢人（房主）
func (s *RoomServiceImpl) KickMember(c *gin.Context, userID int64, roomNo string, targetUserID int64) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if err := s.mustOwner(room, userID); err != nil {
return nil, err
}
if targetUserID == userID {
return nil, errors.New("不能踢出自己")
}

target, err := dao.GetMember(s.DB, room.ID, targetUserID)
if err != nil {
return nil, err
}
if target == nil || target.KickedAt != nil {
return nil, errors.New("该用户不在房间内")
}
if err := dao.KickMember(s.DB, target.ID); err != nil {
return nil, err
}

// 先通知被踢的人，再广播给剩下的人
ws.Default.KickUser(roomNo, targetUserID, ws.Message{
Type: ws.TypeKicked,
Msg:  "你已被房主移出房间",
})

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// DissolveRoom 解散房间（房主）—— 直接从数据库删除
func (s *RoomServiceImpl) DissolveRoom(c *gin.Context, userID int64, roomNo string) error {
room, err := s.mustRoom(roomNo)
if err != nil {
return err
}
if err := s.mustOwner(room, userID); err != nil {
return err
}
if err := dao.DeleteRoom(s.DB, room.ID); err != nil {
return err
}
ws.Default.CloseRoom(roomNo, ws.Message{
Type: ws.TypeDissolved,
Msg:  "房间已被房主解散",
})
return nil
}

// LeaveRoom 退出房间
//   - 普通成员：直接退出
//   - 房主 + 房里还有别人：自动把房主转让给最早加入的成员，然后退出
//   - 房主 + 只剩自己：自动解散房间
func (s *RoomServiceImpl) LeaveRoom(c *gin.Context, userID int64, roomNo string) (*model.LeaveRoomResult, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
member, err := s.mustMember(room.ID, userID)
if err != nil {
return nil, err
}

result := &model.LeaveRoomResult{}

if member.IsOwner {
count, err := dao.CountMembers(s.DB, room.ID)
if err != nil {
return nil, err
}

if count <= 1 {
// 房里只剩房主一个人 -> 直接删除房间
if err := dao.DeleteRoom(s.DB, room.ID); err != nil {
return nil, err
}
ws.Default.CloseRoom(roomNo, ws.Message{
Type: ws.TypeDissolved,
Msg:  "房主已退出，房间自动解散",
})
result.Dissolved = true
return result, nil
}

// 还有别人 -> 房主自动转让给最早加入的那位
next, err := dao.GetEarliestMember(s.DB, room.ID, userID)
if err != nil {
return nil, err
}
if next == nil {
return nil, errors.New("转让房主失败，请重试")
}
if err := dao.TransferOwner(s.DB, room.ID, next.UserID); err != nil {
return nil, err
}
result.NewOwnerID = next.UserID
}

if err := dao.DeleteMember(s.DB, member.ID); err != nil {
return nil, err
}

s.pushState(roomNo)
return result, nil
}

// TransferOwner 转让房主给房间内其他成员（房主）
func (s *RoomServiceImpl) TransferOwner(c *gin.Context, userID int64, roomNo string, targetUserID int64) (*model.RoomDetailVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if err := s.mustOwner(room, userID); err != nil {
return nil, err
}
if targetUserID == userID {
return nil, errors.New("你已经是房主了")
}

target, err := dao.GetMember(s.DB, room.ID, targetUserID)
if err != nil {
return nil, err
}
if target == nil || target.KickedAt != nil {
return nil, errors.New("该用户不在房间内")
}

if err := dao.TransferOwner(s.DB, room.ID, targetUserID); err != nil {
return nil, err
}
room.OwnerID = targetUserID

detail, err := s.buildDetail(room, userID)
if err != nil {
return nil, err
}
s.pushState(roomNo)
return detail, nil
}

// MyRooms 我参与的所有「进行中」房间（大厅「我的房间」列表）
func (s *RoomServiceImpl) MyRooms(c *gin.Context, userID int64) ([]model.RoomBriefVO, error) {
list, err := dao.ListRoomsByUser(s.DB, userID)
if err != nil {
return nil, err
}
result := make([]model.RoomBriefVO, 0, len(list))
for _, r := range list {
result = append(result, model.RoomBriefVO{
RoomNo:      r.RoomNo,
Title:       r.Title,
Round:       r.Round,
MemberCount: r.MemberCount,
OnlineCount: ws.Default.OnlineCount(r.RoomNo),
IsOwner:     r.OwnerID == userID,
CreatedAt:   formatTime(r.CreatedAt),
})
}
return result, nil
}
// ListScoreLogs 计分流水
func (s *RoomServiceImpl) ListScoreLogs(c *gin.Context, userID int64, roomNo string) ([]model.RoomScoreLogVO, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if _, err := s.mustMember(room.ID, userID); err != nil {
return nil, err
}

logs, err := dao.ListScoreLogs(s.DB, room.ID, defaultScoreLogs)
if err != nil {
return nil, err
}

// 一次性取出成员昵称，避免 N+1
members, err := dao.ListMembers(s.DB, room.ID)
if err != nil {
return nil, err
}
nameMap := make(map[int64]string, len(members))
for _, m := range members {
nameMap[m.UserID] = m.Nickname
}

result := make([]model.RoomScoreLogVO, 0, len(logs))
for _, l := range logs {
result = append(result, model.RoomScoreLogVO{
UserID:     l.UserID,
Nickname:   nameMap[l.UserID],
Round:      l.Round,
Delta:      l.Delta,
ScoreAfter: l.ScoreAfter,
CreatedAt:  formatTime(l.CreatedAt),
})
}
return result, nil
}

// CheckMember 校验成员身份（WebSocket 握手用）
func (s *RoomServiceImpl) CheckMember(c *gin.Context, userID int64, roomNo string) (*model.Room, error) {
room, err := s.mustRoom(roomNo)
if err != nil {
return nil, err
}
if _, err := s.mustMember(room.ID, userID); err != nil {
return nil, err
}
return room, nil
}