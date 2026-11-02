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
const avatarURL = "http://192.168.232.128:8080/work-out/avatar/1775762523_446.png"

var namePool = []string{"路人甲", "路人乙", "路人丙", "路人丁", "路人戊", "路人己", "路人庚", "路人辛", "路人壬", "路人癸"}

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
return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
}
return out, nil
}

func main() {
if len(os.Args) < 2 {
fmt.Println("用法: addmembers <room_no> [人数=3]")
return
}
roomNo := os.Args[1]
want := 3
if len(os.Args) >= 3 {
if n, err := strconv.Atoi(os.Args[2]); err == nil && n > 0 {
want = n
}
}

cfg := config.InitLoadConfig()
global.Config = cfg
db, err := gorm.Open(postgres.New(postgres.Config{
DSN:                  cfg.DataSource.Dsn(enum.PostgreSql),
PreferSimpleProtocol: true,
}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
if err != nil {
fmt.Println("连库失败:", err)
return
}

// 查房间
var roomID, ownerID int64
var status int
db.Raw(`SELECT id, owner_id, status FROM room WHERE room_no = ?`, roomNo).Row().Scan(&roomID, &ownerID, &status)
if roomID == 0 {
fmt.Printf("房间 %s 不存在\n", roomNo)
return
}
if status != 1 {
fmt.Printf("房间 %s 已解散，无法加入\n", roomNo)
return
}
fmt.Printf("房间 %s (id=%d) 房主 user_id=%d\n", roomNo, roomID, ownerID)

var before int64
db.Raw(`SELECT count(*) FROM room_member WHERE room_id = ? AND kicked_at IS NULL`, roomID).Scan(&before)
fmt.Printf("当前成员 %d 人，准备加入 %d 人\n\n", before, want)

ownerTok, _ := utils.GenerateToken(ownerID, false)
added := 0
for i := 0; i < want; i++ {
name := fmt.Sprintf("%s%d", namePool[i%len(namePool)], i+1)
openID := fmt.Sprintf("seed_join_%d_%d", time.Now().UnixNano(), i)
var uid int64
if err := db.Raw(`INSERT INTO users (open_id, nickname, avatar_url, bio, background, gender, created_time)
VALUES (?, ?, ?, '', '', 0, NOW()) RETURNING id`, openID, name, avatarURL).Scan(&uid).Error; err != nil {
fmt.Printf("  [X] %s 建用户失败: %v\n", name, err)
continue
}
tok, _ := utils.GenerateToken(uid, false)
jr, err := call("POST", "/api/roomapi/join", tok, map[string]string{"room_no": roomNo})
if err != nil {
fmt.Printf("  [X] %s 加入失败: %v\n", name, err)
continue
}
if code, _ := jr["code"].(float64); code != 0 {
fmt.Printf("  [X] %s 加入失败: %v\n", name, jr["msg"])
continue
}
fmt.Printf("  [√] %-8s (user_id=%d) 已加入\n", name, uid)
added++
}

// 打印最终状态
fmt.Println()
if r, err := call("GET", "/api/roomapi/detail?room_no="+roomNo, ownerTok, nil); err == nil {
if d, ok := r["data"].(map[string]interface{}); ok {
ms, _ := d["members"].([]interface{})
fmt.Printf("房间 %v · 第 %v 轮 · 共 %d 人\n", d["room_no"], d["round"], len(ms))
for i, m := range ms {
mm := m.(map[string]interface{})
role := "成员"
if b, _ := mm["is_owner"].(bool); b {
role = "★ 房主"
}
fmt.Printf("  %2d. user_id=%-4.0f %-10v %5.0f 分   %s\n", i+1, mm["user_id"], mm["nickname"], mm["total_score"], role)
}
} else {
fmt.Println("读取房间详情失败:", r["msg"])
}
}
fmt.Printf("\n本次实际加入 %d 人\n", added)
}