package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// quản lý cấu hình
type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Database []struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`
}

func main() {

	var config Config

	// khởi tạp viper cho đọc file config
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigType("yaml")
	v.SetConfigName("development")
	// read config
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s \n", err)
	}
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshalling config: %s \n", err)
	}

	for _, db := range config.Database {
		fmt.Println("Database host:", db.Host)
		fmt.Println("Database port:", db.Port)
		fmt.Println("Database user:", db.User)
		fmt.Println("Database password:", db.Password)
		fmt.Println("Database name:", db.DBName)
	}
	// or
	// fmt.Println("Database host:", v.GetString("database.host"))
	// fmt.Println("Database port:", v.GetInt("database.port"))
	// fmt.Println("Database user:", v.GetString("database.user"))
	// fmt.Println("Database password:", v.GetString("database.password"))
	// fmt.Println("Database name:", v.GetString("database.dbname"))
}
