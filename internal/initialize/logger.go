package initialize

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger — tạo zap.Logger từ config
// Development: console encoder có màu, debug level, chỉ stdout
// Production/Staging: JSON encoder, info level, file + stdout
func InitLogger(cfg LoggerConfig, app AppConfig) *zap.Logger {
	var encoder zapcore.Encoder
	var writer zapcore.WriteSyncer
	var level zapcore.LevelEnabler

	// Parse level từ config
	level = parseLogLevel(cfg.Level)

	if cfg.Env == "development" {
		encoder = devEncoder()
		writer = zapcore.AddSync(os.Stdout)
	} else {
		encoder = prodEncoder()
		writer = prodWriter(cfg)
	}

	return zap.New(
		zapcore.NewCore(encoder, writer, level),
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.Fields(
			zap.String("service", app.Name),
			zap.String("version", app.Version),
		),
	)
}

func parseLogLevel(lvl string) zapcore.LevelEnabler {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
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

func prodWriter(cfg LoggerConfig) zapcore.WriteSyncer {
	logDir := cfg.LogDir
	if logDir == "" {
		logDir = "logs"
	}
	filename := cfg.Filename
	if filename == "" {
		filename = "app.log"
	}
	maxSize := cfg.MaxSizeMB
	if maxSize == 0 {
		maxSize = 100
	}
	maxBackups := cfg.MaxBackups
	if maxBackups == 0 {
		maxBackups = 30
	}
	maxAge := cfg.MaxAgeDays
	if maxAge == 0 {
		maxAge = 60
	}

	os.MkdirAll(logDir, 0755)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("./%s/%s", logDir, filename),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   cfg.Compress,
	})

	// Production: log ra cả file + console (nếu bật)
	if cfg.EnableConsole {
		return zapcore.NewMultiWriteSyncer(file, zapcore.AddSync(os.Stdout))
	}
	return file
}
