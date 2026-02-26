package initialize

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-base/internal/middlewares"
	"golang-base/internal/routers"
)

// Global references — accessible từ bất kỳ đâu trong app
var (
	Logger *zap.Logger
	DB     *gorm.DB
	Cfg    *Config
)

// Run — bootstrap toàn bộ ứng dụng
// Flow: LoadConfig → InitLogger → InitDB → InitRouter → GracefulShutdown
func Run() {
	// 1. Xác định environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// 2. Load config
	cfg, err := LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}
	Cfg = cfg

	// 3. Init logger
	log := InitLogger(cfg.Logger, cfg.App)
	Logger = log
	defer log.Sync()

	log.Info("config loaded",
		zap.String("env", env),
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
	)

	// 4. Init database
	db, err := InitDatabase(cfg.Database, cfg.Logger, log)
	if err != nil {
		log.Fatal("database init failed", zap.Error(err))
	}
	DB = db
	log.Info("database initialized")

	// 5. Init router
	r := routers.NewRouter(log)

	// Global middleware chain
	r.Use(
		middlewares.Recovery(log),        // Recover panic → JSON error
		middlewares.CORSMiddleware(),     // CORS headers
		middlewares.RequestTimer(),       // Track execution time
		middlewares.RequestLogger(log),   // Structured request logging
		middlewares.RateLimiter(100, 10), // 100 req/s burst, 10 req/s sustained per IP
	)

	// Register routes
	routers.RegisterRoutes(r, log, db)

	// 6. Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server trong goroutine
	go func() {
		log.Info("server started", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	// Chờ signal để graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Cho server 10s để xử lý nốt requests đang chạy
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	// Close DB connection
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Info("server exited properly")
}
