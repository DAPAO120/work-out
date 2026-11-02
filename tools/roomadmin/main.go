package main

import (
"encoding/json"
"fmt"
"io"
"net/http"
"net/url"
"os"
"strconv"
"strings"
"time"

"Project001/common/enum"
"Project001/common/utils"
"Project001/config"
"Project001/global"

"gorm.io/driver/postgres"
"gorm.io/gorm"
"gorm.io/gorm/logger"
)

const apiBase = "http://127.0.0.1:8080"

var db *gorm.DB

func call(method, path, token string, form map[string]string) (map[string]interface{}, error) {
var body io.Reader
if form != nil {
v := url.Values{}
for k, val := range form {
v.Set(k, val)
}
body = strings.NewReader(v.Encode())
}
req, err := http.NewRequest(method, apiBase+path, body)
if err != nil {
return nil, err
}
if form != nil {
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}
if token != "" {
req.Header.Set("Authorization", token)
}
resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()
raw, _ := io.ReadAll(resp.Body)
var out map[string]interface{}
if err := json.Unmarshal(raw, &out); err != nil {
return nil, fmt.Errorf("HTTP %d 非 JSON: %s", resp.StatusCode, string(raw))
}
return out, nil
}

func roomByNo(no string) (int64, int64, bool) {
var id, owner int64
db.Raw(`SELECT id, owner_id FROM room WHERE room_no = ?`, no).Row().Scan(&id, &owner)
return id, owner, id > 0
}

func ownerToken(roomID int64) (string, int64) {
var owner int64
db.Raw(`SELECT owner_id FROM room WHERE id = ?`, roomID).Scan(&owner)
tok, _ := utils.GenerateToken(owner, false)
return tok, owner
}

func printMembers(roomNo string) {
_, owner, ok := roomByNo(roomNo)
if !ok {
fmt.Println("房间不存在:", roomNo)
return
}
tok, _ := utils.GenerateToken(owner, false)
r, err := call("GET", "/api/roomapi/detail?room_no="+roomNo, tok, nil)
if err != nil {
fmt.Println("查询失败:", err)
return
}
d, _ := r["data"].(map[string]interface{})
if d == nil {
fmt.Println("查询失败:", r["msg"])
return
}
ms, _ := d["members"].([]interface{})
fmt.Printf("房间 %v  轮次=%v  成员=%d\n", d["room_no"], d["round"], len(ms))
fmt.Printf("%-4s %-8s %-8s %-8s %s\n", "序", "user_id", "昵称", "分数", "身份")
for i, m := range ms {
mm := m.(map[string]interface{})
role := "成员"
if b, _ := mm["is_owner"].(bool); b {
role = "★ 房主"
}
fmt.Printf("%-4d %-8.0f %-8v %-8.0f %s\n", i+1, mm["user_id"], mm["nickname"], mm["total_score"], role)
}
}

func main() {
cfg := config.InitLoadConfig()
global.Config = cfg
var err error
db, err = gorm.Open(postgres.New(postgres.Config{
DSN:                  cfg.DataSource.Dsn(enum.PostgreSql),
PreferSimpleProtocol: true,
}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
if err != nil {
fmt.Println("连库失败:", err)
return
}

if len(os.Args) < 3 {
fmt.Println("用法: roomadmin <list|owner|kick|reset|dissolve|nextround|score> <room_no> [参数]")
return
}
cmd, roomNo := os.Args[1], os.Args[2]
roomID, _, ok := roomByNo(roomNo)
if !ok {
fmt.Println("房间不存在:", roomNo)
return
}

switch cmd {
case "list":
printMembers(roomNo)

case "owner":
// roomadmin owner <room_no> <newOwnerUserID>
if len(os.Args) < 4 {
fmt.Println("用法: roomadmin owner <room_no> <newOwnerUserID>")
return
}
newOwner, _ := strconv.ParseInt(os.Args[3], 10, 64)
var isMember int64
db.Raw(`SELECT count(*) FROM room_member WHERE room_id = ? AND user_id = ? AND kicked_at IS NULL`, roomID, newOwner).Scan(&isMember)
if isMember == 0 {
fmt.Println("该用户不是房间成员，无法设为房主")
return
}
db.Exec(`UPDATE room SET owner_id = ?, updated_at = NOW() WHERE id = ?`, newOwner, roomID)
db.Exec(`UPDATE room_member SET is_owner = (user_id = ?) WHERE room_id = ?`, newOwner, roomID)
fmt.Printf("房主已改为 user_id=%d，当前状态：\n\n", newOwner)
printMembers(roomNo)

case "kick":
// roomadmin kick <room_no> <targetUserID>
target, _ := strconv.ParseInt(os.Args[3], 10, 64)
tok, _ := ownerToken(roomID)
r, err := call("POST", "/api/roomapi/kick", tok, map[string]string{
"room_no": roomNo, "target_user_id": strconv.FormatInt(target, 10)})
if err != nil {
fmt.Println("失败:", err)
return
}
fmt.Println("结果:", r["code"], r["msg"])
printMembers(roomNo)

case "reset":
tok, _ := ownerToken(roomID)
r, err := call("POST", "/api/roomapi/reset", tok, map[string]string{"room_no": roomNo})
if err != nil {
fmt.Println("失败:", err)
return
}
fmt.Println("结果:", r["code"], r["msg"])
printMembers(roomNo)

case "nextround":
tok, _ := ownerToken(roomID)
r, err := call("POST", "/api/roomapi/nextRound", tok, map[string]string{"room_no": roomNo})
if err != nil {
fmt.Println("失败:", err)
return
}
fmt.Println("结果:", r["code"], r["msg"])
printMembers(roomNo)

case "dissolve":
tok, _ := ownerToken(roomID)
r, err := call("POST", "/api/roomapi/dissolve", tok, map[string]string{"room_no": roomNo})
if err != nil {
fmt.Println("失败:", err)
return
}
fmt.Println("结果:", r["code"], r["msg"])

case "score":
// roomadmin score <room_no> <userID> <delta>
uid, _ := strconv.ParseInt(os.Args[3], 10, 64)
delta := os.Args[4]
tok, _ := utils.GenerateToken(uid, false)
r, err := call("POST", "/api/roomapi/score", tok, map[string]string{"room_no": roomNo, "delta": delta})
if err != nil {
fmt.Println("失败:", err)
return
}
fmt.Println("结果:", r["code"], r["msg"])
printMembers(roomNo)

default:
fmt.Println("未知命令:", cmd)
}
}