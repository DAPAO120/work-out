package global

import (
	"Project001/config"
	"Project001/logger"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	Config *config.AllConfig // 全局Config
	Log    logger.ILogger
	DB     *gorm.DB
	Redis  *redis.Client
)
