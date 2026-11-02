package model

import (
"time"
)

// Room 房间
type Room struct {
ID        int64      `gorm:"primaryKey" json:"id"`
RoomNo    string     `gorm:"type:varchar(8);uniqueIndex;not null;column:room_no" json:"room_no"`
OwnerID   int64      `gorm:"index;column:owner_id" json:"owner_id"`
Title     string     `gorm:"type:varchar(100);default:''" json:"title"`
Status    int8       `gorm:"default:1;column:status" json:"status"`
Round     int        `gorm:"default:1;column:round" json:"round"`
CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
ClosedAt  *time.Time `gorm:"column:closed_at" json:"closed_at"`
}

func (Room) TableName() string {
return "room"
}

// RoomMember 房间成员
type RoomMember struct {
ID         int64      `gorm:"primaryKey" json:"id"`
RoomID     int64      `gorm:"index;column:room_id" json:"room_id"`
UserID     int64      `gorm:"index;column:user_id" json:"user_id"`
Nickname   string     `gorm:"type:varchar(100);column:nickname" json:"nickname"`
Avatar     string     `gorm:"column:avatar" json:"avatar"`
TotalScore int        `gorm:"default:0;column:total_score" json:"total_score"`
IsOwner    bool       `gorm:"default:false;column:is_owner" json:"is_owner"`
JoinedAt   time.Time  `gorm:"column:joined_at" json:"joined_at"`
KickedAt   *time.Time `gorm:"column:kicked_at" json:"kicked_at"`
}

func (RoomMember) TableName() string {
return "room_member"
}

// RoomScore 计分记录
type RoomScore struct {
ID         int64     `gorm:"primaryKey" json:"id"`
RoomID     int64     `gorm:"index;column:room_id" json:"room_id"`
UserID     int64     `gorm:"index;column:user_id" json:"user_id"`
Round      int       `gorm:"column:round" json:"round"`
Delta      int       `gorm:"column:delta" json:"delta"`
ScoreAfter int       `gorm:"column:score_after" json:"score_after"`
CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RoomScore) TableName() string {
return "room_score"
}

// ==================== 请求 / 响应 DTO ====================

// RoomMemberVO 房间成员视图
type RoomMemberVO struct {
UserID     int64  `json:"user_id"`
Nickname   string `json:"nickname"`
Avatar     string `json:"avatar"`
TotalScore int    `json:"total_score"`
IsOwner    bool   `json:"is_owner"`
IsSelf     bool   `json:"is_self"`
JoinedAt   string `json:"joined_at"`
}

// RoomDetailVO 房间详情视图
type RoomDetailVO struct {
RoomNo string `json:"room_no"`
Title  string `json:"title"`
Round  int    `json:"round"`
OwnerID int64 `json:"owner_id"`
// 仅 REST 响应按「请求者」计算；WebSocket 广播中对所有人一致，客户端请用 owner_id 自行判断
IsOwner   bool           `json:"is_owner"`
Status    int8           `json:"status"`
Online    int            `json:"online"`
Members   []RoomMemberVO `json:"members"`
CreatedAt string         `json:"created_at"`
}

// RoomScoreLogVO 单条计分流水
type RoomScoreLogVO struct {
UserID     int64  `json:"user_id"`
Nickname   string `json:"nickname"`
Round      int    `json:"round"`
Delta      int    `json:"delta"`
ScoreAfter int    `json:"score_after"`
CreatedAt  string `json:"created_at"`
}

// RoomBriefVO 大厅「我的房间」列表项
type RoomBriefVO struct {
RoomNo      string `json:"room_no"`
Title       string `json:"title"`
Round       int    `json:"round"`
MemberCount int64  `json:"member_count"`
OnlineCount int    `json:"online_count"`
IsOwner     bool   `json:"is_owner"`
CreatedAt   string `json:"created_at"`
}

// LeaveRoomResult 退出房间的结果
type LeaveRoomResult struct {
// Dissolved = true 表示「房主独自一人退出」，房间已被自动解散
Dissolved bool          `json:"dissolved"`
Room      *RoomDetailVO `json:"room,omitempty"`
// NewOwnerID 房主退出时自动转让给了谁（0 表示没发生转让）
NewOwnerID int64 `json:"new_owner_id,omitempty"`
}