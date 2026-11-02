package global

import (
	"Project001/config"
	"Project001/logger"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	Config *config.AllConfig
	Log    logger.ILogger
	DB     *gorm.DB
	Redis  *redis.Client
)
