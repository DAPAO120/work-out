package main

import (
"encoding/json"
"fmt"
"io"
"net/http"
"net/url"
"os"
"strings"
"time"

"Project001/common/enum"
"Project001/common/utils"
"Project001/config"
"Project001/global"

"github.com/gorilla/websocket"
"gorm.io/driver/postgres"
"gorm.io/gorm"
"gorm.io/gorm/logger"
)

const base = "http://127.0.0.1:8080"
const wsBase = "ws://127.0.0.1:8080"

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
req, err := http.NewRequest(method, base+path, body)
if err != nil {
return nil, err
}
if form != nil {
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}
if token != "" {
req.Header.Set("Authorization", token)
}
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
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

func dataOf(m map[string]interface{}) map[string]interface{} {
if d, ok := m["data"].(map[string]interface{}); ok {
return d
}
return nil
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
os.Exit(1)
}

// ---- 准备两个登录用户：已存在的 id=1，加一个临时用户 ----
var tmpID int64
if err := db.Raw(`INSERT INTO users (open_id, nickname, created_time) VALUES (?, ?, NOW()) RETURNING id`,
"e2e_tmp_"+fmt.Sprint(time.Now().UnixNano()), "E2E临时用户").Scan(&tmpID).Error; err != nil {
fmt.Println("创建临时用户失败:", err)
os.Exit(1)
}
var tmpRoomID int64
defer func() {
db.Exec(`DELETE FROM room_score WHERE user_id IN (1, ?)`, tmpID)
if tmpRoomID > 0 {
db.Exec(`DELETE FROM room_member WHERE room_id = ?`, tmpRoomID)
db.Exec(`DELETE FROM room_score WHERE room_id = ?`, tmpRoomID)
db.Exec(`DELETE FROM room WHERE id = ?`, tmpRoomID)
}
db.Exec(`DELETE FROM users WHERE id = ?`, tmpID)
fmt.Println("\n(已清理测试数据)")
}()

tok1, _ := utils.GenerateToken(1, true)
tok2, _ := utils.GenerateToken(tmpID, false)
fmt.Printf("用户1 token 长度=%d, 临时用户 id=%d\n\n", len(tok1), tmpID)

// ---- 1. 创建房间 ----
fmt.Println("===== 1. 创建房间 =====")
r, err := call("POST", "/api/roomapi/create", tok1, map[string]string{"title": "E2E房间"})
if err != nil {
fmt.Println("请求失败:", err)
return
}
check("创建房间返回 code=0", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
d := dataOf(r)
if d == nil {
fmt.Println("没有 data，终止")
return
}
roomNo := d["room_no"].(string)
tmpRoomID = int64(d["owner_id"].(float64))
db.Raw(`SELECT id FROM room WHERE room_no = ?`, roomNo).Scan(&tmpRoomID)
check("房间号是 6 位数字", len(roomNo) == 6, roomNo)
check("创建者是房主", d["is_owner"].(bool), fmt.Sprint(d["is_owner"]))
check("初始 1 名成员", len(d["members"].([]interface{})) == 1, fmt.Sprint(len(d["members"].([]interface{}))))
check("初始轮次=1", d["round"].(float64) == 1, fmt.Sprint(d["round"]))
fmt.Printf("  房间号 = %s\n", roomNo)

// ---- 2. 第二个用户加入 ----
fmt.Println("\n===== 2. 加入房间 =====")
r, _ = call("POST", "/api/roomapi/join", tok2, map[string]string{"room_no": roomNo})
check("加入成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
d = dataOf(r)
if d != nil {
check("成员数=2", len(d["members"].([]interface{})) == 2, fmt.Sprint(len(d["members"].([]interface{}))))
check("新人不是房主", d["is_owner"].(bool) == false, fmt.Sprint(d["is_owner"]))
}

// 错误房间号
r, _ = call("POST", "/api/roomapi/join", tok2, map[string]string{"room_no": "000000"})
check("不存在的房间号被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))

// ---- 3. 建 WebSocket 连接（用户1）----
fmt.Println("\n===== 3. WebSocket 实时推送 =====")
wsURL := fmt.Sprintf("%s/api/roomapi/ws?room_no=%s&token=%s", wsBase, roomNo, url.QueryEscape(tok1))
conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
if err != nil {
check("WebSocket 连接", false, err.Error())
fmt.Println("无法继续测试实时推送")
} else {
check("WebSocket 连接", true, "")
defer conn.Close()

// 连上应立刻收到一次 room_state
_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
_, msg, err := conn.ReadMessage()
check("连上后立即收到 room_state", err == nil && strings.Contains(string(msg), "room_state"), string(msg))

// ---- 4. 用户2 加分，用户1 应实时收到 ----
fmt.Println("\n===== 4. 计分 + 实时同步 =====")
r, _ = call("POST", "/api/roomapi/score", tok2, map[string]string{"room_no": roomNo, "delta": "25"})
check("用户2 加 25 分", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))

_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
_, msg, err = conn.ReadMessage()
got := err == nil && strings.Contains(string(msg), `"total_score":25`)
check("用户1 实时收到 25 分", got, string(msg))

// ---- 5. 负分 ----
r, _ = call("POST", "/api/roomapi/score", tok2, map[string]string{"room_no": roomNo, "delta": "-10"})
check("用户2 扣 10 分", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
_, msg, _ = conn.ReadMessage()
check("实时收到 15 分", strings.Contains(string(msg), `"total_score":15`), string(msg))

// 非法分值
r, _ = call("POST", "/api/roomapi/score", tok2, map[string]string{"room_no": roomNo, "delta": "0"})
check("分值为 0 被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))

// ---- 6. 房主踢人 ----
fmt.Println("\n===== 5. 房主权限 =====")
// 先让临时用户也建一个 WS，验证被踢时能收到 kicked
wsURL2 := fmt.Sprintf("%s/api/roomapi/ws?room_no=%s&token=%s", wsBase, roomNo, url.QueryEscape(tok2))
conn2, _, err2 := websocket.DefaultDialer.Dial(wsURL2, nil)
if err2 == nil {
_ = conn2.SetReadDeadline(time.Now().Add(3 * time.Second))
_, _, _ = conn2.ReadMessage() // 丢弃初始状态
}

// 非房主踢人应失败
r, _ = call("POST", "/api/roomapi/kick", tok2, map[string]string{"room_no": roomNo, "target_user_id": "1"})
check("非房主踢人被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))
// 非房主解散应失败
r, _ = call("POST", "/api/roomapi/dissolve", tok2, map[string]string{"room_no": roomNo})
check("非房主解散被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))
// 非房主重置应失败
r, _ = call("POST", "/api/roomapi/reset", tok2, map[string]string{"room_no": roomNo})
check("非房主重置被拒绝", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))

// 房主踢人
r, _ = call("POST", "/api/roomapi/kick", tok1, map[string]string{"room_no": roomNo, "target_user_id": fmt.Sprint(tmpID)})
check("房主踢人成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
d = dataOf(r)
if d != nil {
check("踢人后成员数=1", len(d["members"].([]interface{})) == 1, fmt.Sprint(len(d["members"].([]interface{}))))
}
if err2 == nil {
_ = conn2.SetReadDeadline(time.Now().Add(3 * time.Second))
_, msg2, err := conn2.ReadMessage()
check("被踢用户收到 kicked 通知", err == nil && strings.Contains(string(msg2), "kicked"), string(msg2))
conn2.Close()
}
// 被踢的人不能再进来
r, _ = call("POST", "/api/roomapi/join", tok2, map[string]string{"room_no": roomNo})
check("被踢用户无法重新加入", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))

// ---- 7. 下一轮 / 重置 ----
fmt.Println("\n===== 6. 轮次与重置 =====")
r, _ = call("POST", "/api/roomapi/nextRound", tok1, map[string]string{"room_no": roomNo})
check("房主切换下一轮", r["code"].(float64) == 0 && dataOf(r)["round"].(float64) == 2, fmt.Sprint(r["msg"]))

_, _ = call("POST", "/api/roomapi/score", tok1, map[string]string{"room_no": roomNo, "delta": "7"})
r, _ = call("POST", "/api/roomapi/reset", tok1, map[string]string{"room_no": roomNo})
d = dataOf(r)
check("重置后轮次=1", d != nil && d["round"].(float64) == 1, fmt.Sprint(d["round"]))
if d != nil {
ms := d["members"].([]interface{})
first := ms[0].(map[string]interface{})
check("重置后分数=0", first["total_score"].(float64) == 0, fmt.Sprint(first["total_score"]))
}

// ---- 8. 无 token 访问 ----
fmt.Println("\n===== 7. 鉴权 =====")
r, _ = call("GET", "/api/roomapi/detail?room_no="+roomNo, "", nil)
check("无 token 拿详情失败", r["code"] == nil || fmt.Sprint(r["msg"]) != "", fmt.Sprint(r))

// ---- 9. 解散 ----
fmt.Println("\n===== 8. 解散房间 =====")
r, _ = call("POST", "/api/roomapi/dissolve", tok1, map[string]string{"room_no": roomNo})
check("房主解散成功", r["code"].(float64) == 0, fmt.Sprint(r["msg"]))
_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
_, msg, err = conn.ReadMessage()
gotDissolved := false
for kk := 0; kk < 6; kk++ {
_, msgX, errX := conn.ReadMessage()
if errX != nil {
break
}
if strings.Contains(string(msgX), "dissolved") {
gotDissolved = true
msg = []byte("dissolved")
break
}
}
check("收到 dissolved 通知", gotDissolved, string(msg))

// 解散后不能再操作
r, _ = call("POST", "/api/roomapi/score", tok1, map[string]string{"room_no": roomNo, "delta": "1"})
check("解散后无法计分", r["code"].(float64) != 0, fmt.Sprint(r["msg"]))
}

fmt.Printf("\n===== 结果：失败 %d 项 =====\n", failed)
if failed == 0 {
fmt.Println("全部通过 ✅")
}
}
