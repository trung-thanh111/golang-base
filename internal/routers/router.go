package routers // như khai báo namespace trong php

import (
	"fmt"
	"net/http" // import package net/http dùng các trạng thái của api (200, 404, 500, ...)

	"github.com/gin-gonic/gin" // import package gin-gonic dùng để tạo server, framework web
)

func Router() *gin.Engine {
	// Create a Gin router with default middleware (logger and recovery). r mean router
	r := gin.Default()
	// có thể group lại thành v1 or v2 = cách như bên dưới
	v1 := r.Group("/v1")
	{
		// get, put, patch, delete, head, options
		v1.GET("/ping", Pong)
		v1.GET("/ping/:name", Pong) // truyền tham số (name)
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/ping", Pong)
	}
	// Define a simple GET endpoint
	r.GET("/ping", Pong)

	return r // trả về router
}

func AA(c *gin.Context) gin.HandlerFunc {
	fmt.Println("before --> AA")
	c.Next()
	fmt.Println("after --> AA")
	return nil
}
func BB(c *gin.Context) gin.HandlerFunc {
	fmt.Println("before --> BB")
	c.Next()
	fmt.Println("after --> BB")
	return nil
}
func CC(c *gin.Context) {
	fmt.Println("before --> CC")
	c.Next()
	fmt.Println("after --> CC")
}

// Hàm này là hàm xử lý request (c là viết tắt của context)
func Pong(c *gin.Context) {
	// name := c.Param("name") // param là tham số sau dấu /
	name := c.DefaultQuery("name", "trung") // defaultQuery là tham số sau dấu ? và có giá trị mặc định
	id := c.Query("id")                     // query là tham số sau dấu ?
	// Return JSON response with status 200 gin.H => map[string] (là trả về key:value)
	c.JSON(http.StatusOK, gin.H{
		"message": "pong" + name,
		"id":      id,
		"love":    []string{"golang", "php", "python", "121"}, // nhận vào mảng giá trị string
	})
}
