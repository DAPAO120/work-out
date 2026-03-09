package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ILogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	LogrusGinMiddleware() gin.HandlerFunc
}
type SLogger struct {
	logger *logrus.Logger
}

type LogEmailHook struct {
}

// Levels 需要监控的日志等级，只有命中列表中的日志等级才会触发Hook
func (l *LogEmailHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
	}
}

// Fire 触发钩子函数，本实例为触发后发送邮件报警。
func (l *LogEmailHook) Fire(entry *logrus.Entry) error {
	// 触发loggerHook函数调用
	fmt.Println("触发loggerHook函数调用")
	return nil
}

func NewLogger(level string, filePath string) ILogger {
	parseLevel, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err.Error())
	}
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile " + filePath)
		panic(err)
	}
	log := &logrus.Logger{
		Out:   io.MultiWriter(f, os.Stdout),         // 文件 + 控制台输出
		Level: parseLevel,                           // Debug日志等级
		Hooks: make(map[logrus.Level][]logrus.Hook), // 初始化Hook Map,否则导致Hook添加过程中的空指针引用。
		Formatter: &logrus.TextFormatter{ // 文本格式输出
			FullTimestamp:   true,                  // 展示日期
			TimestampFormat: "2006-01-02 15:04:05", //日期格式
			ForceColors:     false,                 // 颜色日志
		},
	}
	log.AddHook(&LogEmailHook{})
	log.Infof("日志开启成功")
	return &SLogger{logger: log}
}
func (l *SLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *SLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *SLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *SLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *SLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *SLogger) LogrusGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 记录请求开始时间
		startTime := time.Now()

		// 2. 处理请求（调用后续的中间件/接口逻辑）
		c.Next()

		// 3. 请求结束后，计算耗时并记录日志
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 提取请求关键信息
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 用 logrus 输出结构化日志（推荐 JSON 格式，便于日志分析）
		logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency, // 耗时
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
		}).Info("gin request log")
	}
}
