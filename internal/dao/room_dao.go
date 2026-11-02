package dao

import (
"Project001/internal/model"
"errors"
"time"

"gorm.io/gorm"
)

// ==================== 房间 ====================

// CreateRoom 创建房间
func CreateRoom(db *gorm.DB, room *model.Room) error {
return db.Create(room).Error
}

// GetRoomByNo 按房间号查询
func GetRoomByNo(db *gorm.DB, roomNo string) (*model.Room, error) {
var room model.Room
err := db.Where("room_no = ?", roomNo).First(&room).Error
if err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, nil
}
return nil, err
}
return &room, nil
}

// GetRoomByID 按主键查询
func GetRoomByID(db *gorm.DB, id int64) (*model.Room, error) {
var room model.Room
err := db.Where("id = ?", id).First(&room).Error
if err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, nil
}
return nil, err
}
return &room, nil
}

// RoomNoExists 房间号是否已存在
func RoomNoExists(db *gorm.DB, roomNo string) (bool, error) {
var count int64
err := db.Model(&model.Room{}).Where("room_no = ?", roomNo).Count(&count).Error
return count > 0, err
}

// UpdateRoom 更新房间
func UpdateRoom(db *gorm.DB, room *model.Room) error {
return db.Save(room).Error
}

// DeleteRoom 解散房间：直接从数据库删除该房间及其成员、计分记录
func DeleteRoom(db *gorm.DB, roomID int64) error {
return db.Transaction(func(tx *gorm.DB) error {
if err := tx.Where("room_id = ?", roomID).Delete(&model.RoomScore{}).Error; err != nil {
return err
}
if err := tx.Where("room_id = ?", roomID).Delete(&model.RoomMember{}).Error; err != nil {
return err
}
return tx.Where("id = ?", roomID).Delete(&model.Room{}).Error
})
}
// UpdateRoomRound 更新轮次
func UpdateRoomRound(db *gorm.DB, roomID int64, round int) error {
return db.Model(&model.Room{}).
Where("id = ?", roomID).
Updates(map[string]interface{}{
"round":      round,
"updated_at": time.Now(),
}).Error
}

// TouchRoom 刷新房间更新时间
func TouchRoom(db *gorm.DB, roomID int64) error {
return db.Model(&model.Room{}).
Where("id = ?", roomID).
Update("updated_at", time.Now()).Error
}

// ==================== 房间成员 ====================

// AddMember 加入房间
func AddMember(db *gorm.DB, member *model.RoomMember) error {
return db.Create(member).Error
}

// GetMember 查询成员（包含已被踢出的记录）
func GetMember(db *gorm.DB, roomID, userID int64) (*model.RoomMember, error) {
var member model.RoomMember
err := db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error
if err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, nil
}
return nil, err
}
return &member, nil
}

// GetMemberByID 按主键查询成员
func GetMemberByID(db *gorm.DB, id int64) (*model.RoomMember, error) {
var member model.RoomMember
err := db.Where("id = ?", id).First(&member).Error
if err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, nil
}
return nil, err
}
return &member, nil
}

// ListMembers 列出房间内所有未被踢出的成员（按加入时间）
func ListMembers(db *gorm.DB, roomID int64) ([]model.RoomMember, error) {
var members []model.RoomMember
err := db.Where("room_id = ? AND kicked_at IS NULL", roomID).
Order("joined_at ASC, id ASC").
Find(&members).Error
return members, err
}

// CountMembers 统计房间有效成员数
func CountMembers(db *gorm.DB, roomID int64) (int64, error) {
var count int64
err := db.Model(&model.RoomMember{}).
Where("room_id = ? AND kicked_at IS NULL", roomID).
Count(&count).Error
return count, err
}

// UpdateMemberScore 更新成员总分
func UpdateMemberScore(db *gorm.DB, memberID int64, score int) error {
return db.Model(&model.RoomMember{}).
Where("id = ?", memberID).
Update("total_score", score).Error
}

// UpdateMemberProfile 重新加入时刷新昵称头像
func UpdateMemberProfile(db *gorm.DB, memberID int64, nickname, avatar string) error {
return db.Model(&model.RoomMember{}).
Where("id = ?", memberID).
Updates(map[string]interface{}{
"nickname":  nickname,
"avatar":    avatar,
"kicked_at": nil,
}).Error
}

// KickMember 踢出成员
func KickMember(db *gorm.DB, memberID int64) error {
now := time.Now()
return db.Model(&model.RoomMember{}).
Where("id = ?", memberID).
Update("kicked_at", now).Error
}

// DeleteMember 成员主动退出
func DeleteMember(db *gorm.DB, memberID int64) error {
return db.Where("id = ?", memberID).Delete(&model.RoomMember{}).Error
}

// ResetRoomScores 重置房间分数与轮次
func ResetRoomScores(db *gorm.DB, roomID int64) error {
return db.Transaction(func(tx *gorm.DB) error {
if err := tx.Model(&model.RoomMember{}).
Where("room_id = ?", roomID).
Update("total_score", 0).Error; err != nil {
return err
}
if err := tx.Where("room_id = ?", roomID).Delete(&model.RoomScore{}).Error; err != nil {
return err
}
return tx.Model(&model.Room{}).
Where("id = ?", roomID).
Updates(map[string]interface{}{
"round":      1,
"updated_at": time.Now(),
}).Error
})
}

// ==================== 计分流水 ====================

// InsertScoreRecord 写入一条计分记录
func InsertScoreRecord(db *gorm.DB, record *model.RoomScore) error {
return db.Create(record).Error
}

// ListScoreLogs 查询房间计分流水（倒序，最多 limit 条）
func ListScoreLogs(db *gorm.DB, roomID int64, limit int) ([]model.RoomScore, error) {
var logs []model.RoomScore
err := db.Where("room_id = ?", roomID).
Order("id DESC").
Limit(limit).
Find(&logs).Error
return logs, err
}

// RoomWithStat 房间 + 成员数统计（大厅列表用）
type RoomWithStat struct {
ID          int64
RoomNo      string
Title       string
Round       int
OwnerID     int64
CreatedAt   time.Time
MemberCount int64
}

// ListRoomsByUser 查询某用户参与的所有「进行中」房间（含有效成员数），按最近活跃排序
func ListRoomsByUser(db *gorm.DB, userID int64) ([]RoomWithStat, error) {
var list []RoomWithStat
err := db.Table("room AS r").
Select(`r.id, r.room_no, r.title, r.round, r.owner_id, r.created_at,
(SELECT COUNT(*) FROM room_member m2 WHERE m2.room_id = r.id AND m2.kicked_at IS NULL) AS member_count`).
Joins("JOIN room_member m ON m.room_id = r.id AND m.user_id = ? AND m.kicked_at IS NULL", userID).
Where("r.status = ?", 1).
Order("r.updated_at DESC, r.id DESC").
Limit(50).
Scan(&list).Error
return list, err
}

// GetEarliestMember 取房间内最早加入的成员（可排除指定用户）
func GetEarliestMember(db *gorm.DB, roomID, excludeUserID int64) (*model.RoomMember, error) {
var m model.RoomMember
err := db.Where("room_id = ? AND kicked_at IS NULL AND user_id <> ?", roomID, excludeUserID).
Order("joined_at ASC, id ASC").
First(&m).Error
if err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, nil
}
return nil, err
}
return &m, nil
}

// TransferOwner 转让房主（同时维护 room_member.is_owner）
func TransferOwner(db *gorm.DB, roomID, newOwnerID int64) error {
return db.Transaction(func(tx *gorm.DB) error {
if err := tx.Model(&model.Room{}).Where("id = ?", roomID).
Updates(map[string]interface{}{
"owner_id":   newOwnerID,
"updated_at": time.Now(),
}).Error; err != nil {
return err
}
if err := tx.Model(&model.RoomMember{}).Where("room_id = ?", roomID).
Update("is_owner", false).Error; err != nil {
return err
}
return tx.Model(&model.RoomMember{}).
Where("room_id = ? AND user_id = ?", roomID, newOwnerID).
Update("is_owner", true).Error
})
}

// GetUsersByIDs 批量查询用户（用于刷新房间内昵称/头像）
func GetUsersByIDs(db *gorm.DB, ids []int64) ([]model.User, error) {
if len(ids) == 0 {
return nil, nil
}
var users []model.User
err := db.Where("id IN ? AND deleted_time IS NULL", ids).Find(&users).Error
return users, err
}