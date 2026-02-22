package main

import (
	"log"

	"golang-base/internal/routers" // import package routers từ internal/routers
)

func main() {
	r := routers.Router() // gọi hàm router
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
