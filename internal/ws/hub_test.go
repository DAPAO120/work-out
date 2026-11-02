package ws

import (
"net/http"
"net/http/httptest"
"strconv"
"strings"
"testing"
"time"

"github.com/gorilla/websocket"
)

// 起一个测试服务器：websocket 连接会被交给 Hub 管理
func newTestServer(t *testing.T, hub *Hub) *httptest.Server {
t.Helper()
upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
roomNo := r.URL.Query().Get("room")
userID, _ := strconv.ParseInt(r.URL.Query().Get("user"), 10, 64)
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
return
}
hub.Serve(roomNo, userID, conn)
}))
t.Cleanup(srv.Close)
return srv
}

func dial(t *testing.T, srv *httptest.Server, room string, user int64) *websocket.Conn {
t.Helper()
url := "ws" + strings.TrimPrefix(srv.URL, "http") +
"?room=" + room + "&user=" + strconv.FormatInt(user, 10)
c, _, err := websocket.DefaultDialer.Dial(url, nil)
if err != nil {
t.Fatalf("dial 失败: %v", err)
}
t.Cleanup(func() { _ = c.Close() })
return c
}

func readJSON(t *testing.T, c *websocket.Conn) string {
t.Helper()
_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
_, data, err := c.ReadMessage()
if err != nil {
t.Fatalf("读取消息失败: %v", err)
}
return string(data)
}

// 广播应该发给房间内所有连接
func TestHubBroadcastToRoom(t *testing.T) {
hub := NewHub()
srv := newTestServer(t, hub)

c1 := dial(t, srv, "123456", 1)
c2 := dial(t, srv, "123456", 2)
other := dial(t, srv, "999999", 3) // 别的房间，不应该收到

time.Sleep(150 * time.Millisecond) // 等注册完成

if got := hub.OnlineCount("123456"); got != 2 {
t.Fatalf("房间 123456 在线数 = %d, 期望 2", got)
}
if got := hub.OnlineCount("999999"); got != 1 {
t.Fatalf("房间 999999 在线数 = %d, 期望 1", got)
}

hub.Broadcast("123456", Message{Type: TypeRoomState, Data: map[string]string{"room_no": "123456"}})

for i, c := range []*websocket.Conn{c1, c2} {
msg := readJSON(t, c)
if !strings.Contains(msg, "room_state") || !strings.Contains(msg, "123456") {
t.Fatalf("客户端 %d 收到 %s", i, msg)
}
}

// other 不该收到任何消息
_ = other.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
if _, _, err := other.ReadMessage(); err == nil {
t.Fatal("别的房间不应该收到消息")
}
}

// 踢人：只通知目标用户，并且断开他的连接
func TestHubKickUser(t *testing.T) {
hub := NewHub()
srv := newTestServer(t, hub)

victim := dial(t, srv, "222222", 7)
bystander := dial(t, srv, "222222", 8)
time.Sleep(150 * time.Millisecond)

hub.KickUser("222222", 7, Message{Type: TypeKicked, Msg: "你已被移出"})

msg := readJSON(t, victim)
if !strings.Contains(msg, "kicked") {
t.Fatalf("被踢用户收到 %s", msg)
}

// 被踢用户的连接应被关闭
_ = victim.SetReadDeadline(time.Now().Add(2 * time.Second))
for {
if _, _, err := victim.ReadMessage(); err != nil {
break
}
}

// 旁边的人不受影响
if got := hub.OnlineCount("222222"); got != 1 {
t.Fatalf("踢人后在线数 = %d, 期望 1", got)
}

// 旁边的人仍然能收到广播
hub.Broadcast("222222", Message{Type: TypeRoomState, Data: map[string]string{"room_no": "222222"}})
if msg := readJSON(t, bystander); !strings.Contains(msg, "222222") {
t.Fatalf("旁观者收到 %s", msg)
}
}

// 解散：广播后清空房间所有连接
func TestHubCloseRoom(t *testing.T) {
hub := NewHub()
srv := newTestServer(t, hub)

c1 := dial(t, srv, "333333", 1)
c2 := dial(t, srv, "333333", 2)
time.Sleep(150 * time.Millisecond)

hub.CloseRoom("333333", Message{Type: TypeDissolved, Msg: "房间已解散"})

for i, c := range []*websocket.Conn{c1, c2} {
if msg := readJSON(t, c); !strings.Contains(msg, "dissolved") {
t.Fatalf("客户端 %d 收到 %s", i, msg)
}
}

time.Sleep(200 * time.Millisecond)
if got := hub.OnlineCount("333333"); got != 0 {
t.Fatalf("解散后在线数 = %d, 期望 0", got)
}
}