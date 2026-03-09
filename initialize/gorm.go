package initialize

import (
	"Project001/common/enum"
	"Project001/global"
	"errors"
	"time"

	"gorm.io/driver/postgres"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	GormToManyRequestError = errors.New("gorm: to many request")
)

func InitDatabase(dsn string, database enum.Database) *gorm.DB {
	if database == enum.Mysql {
		db := MysqlDatabase(dsn)
		return db
	}
	if database == enum.PostgreSql {
		db := PostgreDatabase(dsn)
		return db
	}
	return PostgreDatabase(dsn)
}

func MysqlDatabase(dsn string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)  //设置连接池，空闲
	sqlDB.SetMaxOpenConns(100) //打开
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	// 慢日志中间件
	SlowQueryLog(db)
	return db
}

func PostgreDatabase(dsn string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}

	// PostgreSQL 连接配置
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,   // DSN data source name
		PreferSimpleProtocol: false, // 禁用隐式预处理语句，如果设置为true，则不会使用PREPARE语句
	}), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		// PostgreSQL特有的配置选项
		PrepareStmt:            false, // 是否开启预处理语句缓存
		SkipDefaultTransaction: true,  // 跳过默认事务以提升性能
	})

	if err != nil {
		panic(err)
	}

	// 获取底层的 sql.DB 对象并设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 连接池配置（这些配置与MySQL版本相同，可以保留）
	sqlDB.SetMaxIdleConns(20)                  // 设置空闲连接池中的最大连接数
	sqlDB.SetMaxOpenConns(100)                 // 设置数据库连接最大打开数
	sqlDB.SetConnMaxLifetime(time.Second * 30) // 设置连接可复用的最大时间

	// 注册慢查询回调
	SlowQueryLog(db)

	return db
}

// SlowQueryLog 慢查询日志
func SlowQueryLog(db *gorm.DB) {
	err := db.Callback().Query().Before("*").Register("slow_query_start", func(d *gorm.DB) {
		now := time.Now()
		d.Set("start_time", now)
	})
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().After("*").Register("slow_query_end", func(d *gorm.DB) {
		now := time.Now()
		start, ok := d.Get("start_time")
		if ok {
			duration := now.Sub(start.(time.Time))
			// 一般认为 200 Ms 为Sql慢查询
			if duration > time.Millisecond*200 {
				global.Log.Error(d.Statement.Context, "慢查询", "SQL:", d.Statement.SQL.String())
			}
		}
	})
	if err != nil {
		panic(err)
	}
}
