package main

import (
	"log"

	"golang-base/internal/middlewares" // import middleware request time
	"golang-base/internal/routers"     // import package routers từ internal/routers
)

func main() {
	r := routers.Router()             // gọi hàm router
	r.Use(middlewares.RequestTimer()) // dùng middle để tgian bắt đầu chạy request
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
