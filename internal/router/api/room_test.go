package api

import (
"net/http"
"net/http/httptest"
"testing"

"Project001/config"
"Project001/global"

"github.com/gin-gonic/gin"
)

func setupRoomRouter(t *testing.T) *gin.Engine {
t.Helper()
gin.SetMode(gin.TestMode)
// ParseToken 会用到 jwt secret，这里给个测试值
global.Config = &config.AllConfig{Jwt: config.Jwt{Secret: "test-secret"}}

r := gin.New()
parent := r.Group("/api")
(&RoomRouter{}).InitApiRouter(parent)
return r
}

func TestRoomRoutesRegistered(t *testing.T) {
r := setupRoomRouter(t)

want := []string{
"GET /api/roomapi/ws",
"POST /api/roomapi/create",
"POST /api/roomapi/join",
"GET /api/roomapi/detail",
"POST /api/roomapi/score",
"POST /api/roomapi/nextRound",
"POST /api/roomapi/reset",
"POST /api/roomapi/kick",
"POST /api/roomapi/dissolve",
"POST /api/roomapi/leave",
"GET /api/roomapi/logs",
}

got := make(map[string]bool)
for _, rt := range r.Routes() {
got[rt.Method+" "+rt.Path] = true
}
for _, w := range want {
if !got[w] {
t.Errorf("缺少路由: %s", w)
}
}
}

func TestRoomPrivateRoutesRequireAuth(t *testing.T) {
r := setupRoomRouter(t)

cases := []struct{ method, path string }{
{"POST", "/api/roomapi/create"},
{"POST", "/api/roomapi/join"},
{"POST", "/api/roomapi/score"},
{"POST", "/api/roomapi/nextRound"},
{"POST", "/api/roomapi/reset"},
{"POST", "/api/roomapi/kick"},
{"POST", "/api/roomapi/dissolve"},
{"POST", "/api/roomapi/leave"},
{"GET", "/api/roomapi/detail?room_no=123456"},
{"GET", "/api/roomapi/logs?room_no=123456"},
}

for _, c := range cases {
w := httptest.NewRecorder()
req, _ := http.NewRequest(c.method, c.path, nil)
r.ServeHTTP(w, req)
if w.Code != http.StatusUnauthorized {
t.Errorf("%s %s 期望 401，实际 %d", c.method, c.path, w.Code)
}
}
}

func TestRoomSocketRejectsInvalidToken(t *testing.T) {
r := setupRoomRouter(t)

w := httptest.NewRecorder()
req, _ := http.NewRequest("GET", "/api/roomapi/ws?room_no=123456&token=not-a-jwt", nil)
r.ServeHTTP(w, req)
if w.Code != http.StatusUnauthorized {
t.Errorf("ws 无效 token 期望 401，实际 %d", w.Code)
}
}

func TestRoomSocketRequiresToken(t *testing.T) {
r := setupRoomRouter(t)

w := httptest.NewRecorder()
req, _ := http.NewRequest("GET", "/api/roomapi/ws?room_no=123456", nil)
r.ServeHTTP(w, req)
if w.Code != http.StatusUnauthorized {
t.Errorf("ws 缺 token 期望 401，实际 %d", w.Code)
}
}