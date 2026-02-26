package initialize

import (
	"fmt"

	"github.com/spf13/viper"
)

// ============================================================
// CONFIG — Cấu hình ứng dụng, mapping 1:1 với file YAML
// ============================================================

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AWS      AWSConfig      `mapstructure:"aws"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type LoggerConfig struct {
	Env           string `mapstructure:"env"`
	Level         string `mapstructure:"level"`
	LogDir        string `mapstructure:"log_dir"`
	Filename      string `mapstructure:"filename"`
	MaxSizeMB     int    `mapstructure:"max_size_mb"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxAgeDays    int    `mapstructure:"max_age_days"`
	Compress      bool   `mapstructure:"compress"`
	EnableConsole bool   `mapstructure:"enable_console"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	SSLMode         string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
}

// LoadConfig — load YAML config + .env override
// env: "development" | "staging" | "production"
func LoadConfig(env string) (*Config, error) {
	v := viper.New()

	// 1. Load .env file trước (nếu có) cho sensitive values
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.MergeInConfig() // ignore error nếu không có .env

	// 2. Load YAML config chính
	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./config/")
	v.AddConfigPath("../config/")
	v.AddConfigPath("../../config/")

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("load config %s.yaml failed: %w", env, err)
	}

	// 3. Cho phép ENV vars override YAML
	// VD: DATABASE_PASSWORD=xxx sẽ override database.password
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	return &cfg, nil
}
