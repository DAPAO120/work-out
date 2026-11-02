package main

import (
"encoding/json"
"fmt"
"io"
"net/http"
"net/url"
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

var seedNames = []string{"小明", "阿强", "老王", "翠花", "大熊", "二狗", "铁蛋", "花花"}
var seedScores = []int{8, 15, 23, 5, 31, 12, -6, 19}

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
return nil, fmt.Errorf("非 JSON 响应(HTTP %d): %s", resp.StatusCode, string(raw))
}
return out, nil
}

func codeOf(m map[string]interface{}) float64 {
if m == nil {
return -1
}
if c, ok := m["code"].(float64); ok {
return c
}
return -1
}

func msgOf(m map[string]interface{}) string {
if m == nil {
return "nil"
}
if s, ok := m["msg"].(string); ok {
return s
}
return fmt.Sprint(m)
}

func main() {
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

ownerTok, _ := utils.GenerateToken(1, true)

// ---------- 1. 老大1 建房 ----------
r, err := call("POST", "/api/roomapi/create", ownerTok, map[string]string{"title": "8人测试房"})
if err != nil {
fmt.Println("建房请求失败:", err)
return
}
if codeOf(r) != 0 {
fmt.Println("建房失败:", msgOf(r))
return
}
data, _ := r["data"].(map[string]interface{})
roomNo, _ := data["room_no"].(string)
var roomID int64
db.Raw(`SELECT id FROM room WHERE room_no = ?`, roomNo).Scan(&roomID)
fmt.Printf("房间已创建：%s\n", roomNo)

// ---------- 2. 造 8 个用户并加入 ----------
fmt.Println("\n成员加入情况：")
okCount := 0
for i, name := range seedNames {
openID := fmt.Sprintf("seed_room_%d_%d", time.Now().UnixNano(), i)
var uid int64
if err := db.Raw(`INSERT INTO users (open_id, nickname, avatar_url, bio, background, gender, created_time)
VALUES (?, ?, ?, '', '', 0, NOW()) RETURNING id`, openID, name, avatarURL).Scan(&uid).Error; err != nil {
fmt.Printf("  [X] %-4s 建用户失败: %v\n", name, err)
continue
}
tok, _ := utils.GenerateToken(uid, false)
jr, err := call("POST", "/api/roomapi/join", tok, map[string]string{"room_no": roomNo})
if err != nil || codeOf(jr) != 0 {
fmt.Printf("  [X] %-4s 加入失败: %v %s\n", name, err, msgOf(jr))
continue
}
if seedScores[i] != 0 {
_, _ = call("POST", "/api/roomapi/score", tok, map[string]string{
"room_no": roomNo,
"delta":   fmt.Sprint(seedScores[i]),
})
}
fmt.Printf("  [√] %-4s (user_id=%-3d) 已加入，初始分 %d\n", name, uid, seedScores[i])
okCount++
}

// ---------- 3. 老大1 也记 12 分 ----------
_, _ = call("POST", "/api/roomapi/score", ownerTok, map[string]string{"room_no": roomNo, "delta": "12"})

// ---------- 4. 拉一次房间详情确认 ----------
dr, err := call("GET", "/api/roomapi/detail?room_no="+roomNo, ownerTok, nil)
if err == nil && codeOf(dr) == 0 {
d, _ := dr["data"].(map[string]interface{})
ms, _ := d["members"].([]interface{})
fmt.Printf("\n===== 房间当前状态（轮次 %v，共 %d 人）=====\n", d["round"], len(ms))
for i, m := range ms {
mm := m.(map[string]interface{})
tag := ""
if b, _ := mm["is_owner"].(bool); b {
tag = "  ← 房主"
}
fmt.Printf("  %d. %-6v %4v 分%s\n", i+1, mm["nickname"], mm["total_score"], tag)
}
}

fmt.Printf("\n==========================================================\n")
fmt.Printf("  房间号： %s      （记下来，等会儿在小程序里输入）\n", roomNo)
fmt.Printf("  房主：   老大1 (user_id=1)\n")
fmt.Printf("  成员：   %d 人\n", okCount+1)
fmt.Printf("==========================================================\n")
fmt.Printf("\n【清理测试数据用】\n")
fmt.Printf("DELETE FROM room_score  WHERE room_id = %d;\n", roomID)
fmt.Printf("DELETE FROM room_member WHERE room_id = %d;\n", roomID)
fmt.Printf("DELETE FROM room        WHERE id = %d;\n", roomID)
fmt.Printf("DELETE FROM users WHERE open_id LIKE 'seed_room_%%';\n")
}