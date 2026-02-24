package logger

import (
	"os"   // thư viện tương tác với hệ điều hành
	"time" // thư viện tính toán thời gian

	"github.com/gin-gonic/gin"         // thư viện web framework
	"github.com/google/uuid"           // thư viện tạo id duy nhất
	"go.uber.org/zap"                  // thư viện log
	"go.uber.org/zap/zapcore"          // zapcore được sử dụng để tạo encoder và writer
	"gopkg.in/natefinch/lumberjack.v2" // thư viện giới hạn size file log, tạo log mới khi quá size, xóa log cũ theo ngày or bản backup
)

func New(env string) *zap.Logger {
	var encoder zapcore.Encoder
	var writer zapcore.WriteSyncer
	var level zapcore.LevelEnabler

	if env == "development" {
		encoder = devEncoder()
		writer = zapcore.AddSync(os.Stdout)
		level = zap.DebugLevel
	} else {
		encoder = prodEncoder()
		writer = prodWriter()
		level = zap.InfoLevel
	}

	return zap.New(
		zapcore.NewCore(encoder, writer, level),
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.Fields(
			zap.String("service", os.Getenv("APP_NAME")),
			zap.String("version", os.Getenv("APP_VERSION")),
		),
	)
}

func devEncoder() zapcore.Encoder {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	return zapcore.NewConsoleEncoder(cfg)
}

func prodEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.LevelKey = "level"
	cfg.MessageKey = "message"
	cfg.CallerKey = "caller"
	cfg.StacktraceKey = "stacktrace"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(cfg)
}

func prodWriter() zapcore.WriteSyncer {
	os.MkdirAll("logs", 0755)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    100, // MB
		MaxBackups: 30,
		MaxAge:     60, // ngày
		Compress:   true,
	})

	return zapcore.NewMultiWriteSyncer(file, zapcore.AddSync(os.Stdout))
}

// RequestLogger — middleware Gin log mỗi HTTP request
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		status := c.Writer.Status()
		logFn := log.Info
		if status >= 500 {
			logFn = log.Error
		} else if status >= 400 {
			logFn = log.Warn
		}

		logFn("request",
			zap.String("id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}

func generateID() string {
	return uuid.New().String() // "550e8400-e29b-41d4-a716-446655440000"
}
