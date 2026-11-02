package ws

import (
"encoding/json"
"sync"
"time"

"github.com/gorilla/websocket"
)

// ---------------- 消息协议 ----------------

// Message 服务端推送给小程序的消息
type Message struct {
Type string      `json:"type"`
Data interface{} `json:"data,omitempty"`
Msg  string      `json:"msg,omitempty"`
}

const (
TypeRoomState = "room_state" // 房间全量状态（增删改分后都会推）
TypeDissolved = "dissolved"  // 房间被房主解散
TypeKicked    = "kicked"     // 你被房主踢出
TypeError     = "error"      // 业务错误
)

const (
writeWait  = 10 * time.Second
pongWait   = 60 * time.Second
pingPeriod = (pongWait * 9) / 10
sendBuffer = 64
)

// outMessage 待发送的一帧：closeAfter=true 表示发完这一帧后优雅关闭连接
type outMessage struct {
payload    []byte
closeAfter bool
}

// ---------------- Client ----------------

// Client 一条 WebSocket 连接
type Client struct {
hub    *Hub
conn   *websocket.Conn
out    chan outMessage
done   chan struct{}
RoomNo string
UserID int64

shutdownOnce sync.Once
}

// ---------------- Hub ----------------

// Hub 房间广播中心：roomNo -> 连接集合
type Hub struct {
mu    sync.RWMutex
rooms map[string]map[*Client]struct{}
}

// Default 全局默认 Hub
var Default = NewHub()

func NewHub() *Hub {
return &Hub{rooms: make(map[string]map[*Client]struct{})}
}

// Serve 接管一条已升级的连接
func (h *Hub) Serve(roomNo string, userID int64, conn *websocket.Conn) *Client {
c := &Client{
hub:    h,
conn:   conn,
out:    make(chan outMessage, sendBuffer),
done:   make(chan struct{}),
RoomNo: roomNo,
UserID: userID,
}
h.register(c)
go c.writePump()
go c.readPump()
return c
}

func (h *Hub) register(c *Client) {
h.mu.Lock()
defer h.mu.Unlock()
if h.rooms[c.RoomNo] == nil {
h.rooms[c.RoomNo] = make(map[*Client]struct{})
}
h.rooms[c.RoomNo][c] = struct{}{}
}

// unregister 从房间集合中移除（幂等，不关闭通道，避免并发发送 panic）
func (h *Hub) unregister(c *Client) {
h.mu.Lock()
defer h.mu.Unlock()
if set, ok := h.rooms[c.RoomNo]; ok {
delete(set, c)
if len(set) == 0 {
delete(h.rooms, c.RoomNo)
}
}
}

// Broadcast 给房间内所有连接推送消息
func (h *Hub) Broadcast(roomNo string, msg Message) {
h.broadcast(roomNo, nil, msg, false)
}

// SendTo 给房间内指定用户的所有连接推送消息
func (h *Hub) SendTo(roomNo string, userID int64, msg Message) {
h.broadcast(roomNo, &userID, msg, false)
}

// CloseRoom 解散房间：先广播消息，再断开该房间所有连接
func (h *Hub) CloseRoom(roomNo string, msg Message) {
h.broadcast(roomNo, nil, msg, true)
}

// KickUser 踢人：先通知并断开该用户的连接
func (h *Hub) KickUser(roomNo string, userID int64, msg Message) {
h.broadcast(roomNo, &userID, msg, true)
}

func (h *Hub) broadcast(roomNo string, onlyUser *int64, msg Message, closeAfter bool) {
payload, err := json.Marshal(msg)
if err != nil {
return
}

h.mu.RLock()
var targets []*Client
for c := range h.rooms[roomNo] {
if onlyUser != nil && c.UserID != *onlyUser {
continue
}
targets = append(targets, c)
}
h.mu.RUnlock()

for _, c := range targets {
// 先把消息排进发送队列，writePump 会先写完这一帧再关闭连接
c.enqueue(payload, closeAfter)
if closeAfter {
// 立刻从房间移除，避免继续收到广播
h.unregister(c)
}
}
}

// OnlineCount 房间当前在线连接数
func (h *Hub) OnlineCount(roomNo string) int {
h.mu.RLock()
defer h.mu.RUnlock()
return len(h.rooms[roomNo])
}

// ---------------- Client 内部 ----------------

// enqueue 非阻塞入队；队列满或已关闭则强制断开
func (c *Client) enqueue(payload []byte, closeAfter bool) {
select {
case <-c.done:
return
default:
}

select {
case c.out <- outMessage{payload: payload, closeAfter: closeAfter}:
case <-c.done:
default:
c.forceClose()
}
}

// shutdown 幂等：注销 + 通知读写协程退出
func (c *Client) shutdown() {
c.shutdownOnce.Do(func() {
c.hub.unregister(c)
close(c.done)
})
}

// forceClose 立即断开
func (c *Client) forceClose() {
c.shutdown()
_ = c.conn.Close()
}

// Push 向本连接单独推送一条消息
func (c *Client) Push(msg Message) {
payload, err := json.Marshal(msg)
if err != nil {
return
}
c.enqueue(payload, false)
}

// ---------------- 读写循环 ----------------

func (c *Client) readPump() {
defer c.forceClose()

c.conn.SetReadLimit(4096)
_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
c.conn.SetPongHandler(func(string) error {
return c.conn.SetReadDeadline(time.Now().Add(pongWait))
})

for {
// 客户端只发心跳，读掉即可
if _, _, err := c.conn.ReadMessage(); err != nil {
return
}
_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
}

func (c *Client) writePump() {
ticker := time.NewTicker(pingPeriod)
defer func() {
ticker.Stop()
_ = c.conn.Close()
c.shutdown()
}()

writeClose := func() {
_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
_ = c.conn.WriteMessage(websocket.CloseMessage,
websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

for {
select {
case <-c.done:
writeClose()
return
case msg := <-c.out:
_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
if err := c.conn.WriteMessage(websocket.TextMessage, msg.payload); err != nil {
return
}
if msg.closeAfter {
writeClose()
return
}
case <-ticker.C:
_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
return
}
}
}
}