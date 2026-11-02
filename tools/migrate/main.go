package main

import (
"fmt"
"os"
"strings"

"Project001/common/enum"
"Project001/config"

"gorm.io/driver/postgres"
"gorm.io/gorm"
"gorm.io/gorm/logger"
)

// 只执行 script/work_out.sql 里「大厅 / 房间计分模块」之后的建表语句
func main() {
raw, err := os.ReadFile("script/work_out.sql")
if err != nil {
fmt.Println("读取 SQL 失败:", err)
return
}

const marker = "-- 大厅 / 房间计分模块"
idx := strings.Index(string(raw), marker)
if idx < 0 {
fmt.Println("SQL 文件里找不到标记:", marker)
return
}
ddl := string(raw)[idx:]
// 去掉紧跟标记的那行注释，保留后面的语句
if p := strings.Index(ddl, "\n"); p >= 0 {
ddl = ddl[p+1:]
}

cfg := config.InitLoadConfig()
db, err := gorm.Open(postgres.New(postgres.Config{
DSN:                  cfg.DataSource.Dsn(enum.PostgreSql),
PreferSimpleProtocol: true, // 允许多语句一次性执行
}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
if err != nil {
fmt.Println("连接数据库失败:", err)
return
}

if err := db.Exec(ddl).Error; err != nil {
fmt.Println("执行 DDL 失败:", err)
return
}
fmt.Println("DDL 执行成功")

for _, t := range []string{"room", "room_member", "room_score"} {
var exists bool
db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name = ?)`, t).Scan(&exists)
fmt.Printf("  %-14s %v\n", t, exists)
}
}