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

var db *gorm.DB
var failed int

func check(name string, ok bool, detail string) {
if ok {
fmt.Printf("  [PASS] %s\n", name)
} else {
failed++
fmt.Printf("  [FAIL] %s  -> %s\n", name, detail)
}
}

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

func data(m map[string]interface{}) map[string]interface{} {
if d, ok := m["data"].(map[string]interface{}); ok {
return d
}
return nil
}

// 查我的房间列表里某个房间的 is_owner
func myRoomFlag(token, roomNo string) (bool, bool, int64) {
r, err := call("GET", "/api/roomapi/myRooms", token, nil)
if err != nil {
return false, false, 0
}
arr, _ := r["data"].([]interface{})
for _, it := range arr {
mm, _ := it.(map[string]interface{})
if mm == nil {
continue
}
if fmt.Sprint(mm["room_no"]) == roomNo {
owner, _ := mm["is_owner"].(bool)
cnt := int64(0)
if c, ok := mm["member_count"].(float64); ok {
cnt = int64(c)
}
return true, owner, cnt
}
}
return false, false, 0
}

func main() {
cfg := config.InitLoadConfig()
global.Config = cfg
db, _ = gorm.Open(postgres.New(postgres.Config{
DSN:                  cfg.DataSource.Dsn(enum.PostgreSql),
PreferSimpleProtocol: true,
}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

// 两个临时用户
mkUser := func(name string) int64 {
var id int64
db.Raw(`INSERT INTO users (open_id, nickname, avatar_url, bio, background, gender, created_time)
VALUES (?, ?, '', '', '', 0, NOW()) RETURNING id`, fmt.Sprintf("e2e2_%d_%s", time.Now().UnixNano(), name), name).Scan(&id)
return id
}
uidA := mkUser("测试A")
uidB := mkUser("测试B")
var roomID int64
defer func() {
if roomID > 0 {
db.Exec(`DELETE FROM room_score WHERE room_id = ?`, roomID)
db.Exec(`DELETE FROM room_member WHERE room_id = ?`, roomID)
db.Exec(`DELETE FROM room WHERE id = ?`, roomID)
}
db.Exec(`DELETE FROM users WHERE id IN (?, ?)`, uidA, uidB)
fmt.Println("\n(已清理测试数据)")
}()

tok1, _ := utils.GenerateToken(1, true)
tokA, _ := utils.GenerateToken(uidA, false)
tokB, _ := utils.GenerateToken(uidB, false)

// 1) 老大1 建房
r, _ := call("POST", "/api/roomapi/create", tok1, map[string]string{})
if data(r) == nil {
		fmt.Printf("创建房间失败，原始响应: %v`n", r)
		return
	}
	d := data(r)
roomNo := d["room_no"].(string)
db.Raw(`SELECT id FROM room WHERE room_no = ?`, roomNo).Scan(&roomID)
fmt.Printf("房间 %s (id=%d)\n\n", roomNo, roomID)

// 2) A B 加入
fmt.Println("===== 1. 加入 =====")
call("POST", "/api/roomapi/join", tokA, map[string]string{"room_no": roomNo})
call("POST", "/api/roomapi/join", tokB, map[string]string{"room_no": roomNo})
found, isOwner, cnt := myRoomFlag(tok1, roomNo)
check("myRooms 能查到房间", found, "没查到")
check("老大1 显示为房主", isOwner, "is_owner=false")
check("成员数=3", cnt == 3, fmt.Sprint(cnt))

// 3) 转让给 A
fmt.Println("\n===== 2. 转让房主 =====")
r, _ = call("POST", "/api/roomapi/transfer", tok1, map[string]string{"room_no": roomNo, "target_user_id": fmt.Sprint(uidA)})
check("房主转让成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
if d = data(r); d != nil {
check("返回的 owner_id 是新房主", int64(d["owner_id"].(float64)) == uidA, fmt.Sprint(d["owner_id"]))
}
_, isOwner1, _ := myRoomFlag(tok1, roomNo)
_, isOwnerA, _ := myRoomFlag(tokA, roomNo)
check("原房主 myRooms.is_owner=false", !isOwner1, "还是 true")
check("新房主 myRooms.is_owner=true", isOwnerA, "还是 false")

// 非房主不能转让
r, _ = call("POST", "/api/roomapi/transfer", tok1, map[string]string{"room_no": roomNo, "target_user_id": fmt.Sprint(uidB)})
check("非房主转让被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))

// 4) 房主 A 退出 -> 应自动转让给最早加入的老大1
fmt.Println("\n===== 3. 房主退出（自动转让）=====")
r, _ = call("POST", "/api/roomapi/leave", tokA, map[string]string{"room_no": roomNo})
check("退出成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
if d = data(r); d != nil {
check("自动转让给老大1", int64(d["new_owner_id"].(float64)) == 1, fmt.Sprint(d["new_owner_id"]))
check("未解散", d["dissolved"] == false, fmt.Sprint(d["dissolved"]))
}
_, isOwner1, cnt = myRoomFlag(tok1, roomNo)
check("老大1 重新成为房主", isOwner1, "不是房主")
check("剩余成员=2", cnt == 2, fmt.Sprint(cnt))

// 5) 普通成员 B 退出
fmt.Println("\n===== 4. 普通成员退出 =====")
r, _ = call("POST", "/api/roomapi/leave", tokB, map[string]string{"room_no": roomNo})
check("普通成员退出成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
if d = data(r); d != nil {
check("普通成员退出不会解散", d["dissolved"] == false, fmt.Sprint(d["dissolved"]))
}
_, _, cnt = myRoomFlag(tok1, roomNo)
check("剩余成员=1", cnt == 1, fmt.Sprint(cnt))

// 6) 只剩房主 -> 退出应自动解散
fmt.Println("\n===== 5. 房主独自退出（自动解散）=====")
r, _ = call("POST", "/api/roomapi/leave", tok1, map[string]string{"room_no": roomNo})
check("退出成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
if d = data(r); d != nil {
check("返回 dissolved=true", d["dissolved"] == true, fmt.Sprint(d["dissolved"]))
}
var left int64
db.Raw(`SELECT count(*) FROM room WHERE id = ?`, roomID).Scan(&left)
check("数据库里房间已被物理删除", left == 0, fmt.Sprint(left))
found, _, _ = myRoomFlag(tok1, roomNo)
check("myRooms 不再返回已解散房间", !found, "还在列表里")

fmt.Printf("\n===== 结果：失败 %d 项 =====\n", failed)
if failed == 0 {
fmt.Println("全部通过 ✅")
}
}