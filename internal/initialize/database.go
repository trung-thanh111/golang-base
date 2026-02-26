package initialize

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase — khởi tạo GORM MySQL connection từ config
func InitDatabase(cfg DatabaseConfig, loggerCfg LoggerConfig, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// GORM log level theo env
	gormLogLevel := logger.Silent
	if loggerCfg.Env == "development" {
		gormLogLevel = logger.Info // dev: show tất cả SQL queries
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	// Connection pool config
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)                                    // số lượng kết nối rảnh rỗi tối đa
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)                                    // số lượng kết nối tối đa -> kiểm soát tài nguyên tránh crack
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second) // thời gian sống của kết nối -> tránh kết nối chết

	log.Info("database connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	return db, nil
}
