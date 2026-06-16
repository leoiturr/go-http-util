package main

import (
	"html/template"
	"log"
	"net/http"

	"gin-example/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode in production, debug mode in development
	gin.SetMode(gin.DebugMode)

	// Create Gin router
	r := gin.Default()

	// Register custom template functions before loading HTML templates
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	})

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Main web interface route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{})
	})

	// Base64 Converter API
	r.POST("/api/base64/encode", handlers.EncodeBase64)
	r.POST("/api/base64/decode", handlers.DecodeBase64)

	// GUID Generator API
	r.POST("/api/guid/generate", handlers.GenerateGUIDs)

	// QR Code Generator API
	r.POST("/api/qrcode/generate", handlers.GenerateQRCode)

	// Start Gin server on port 8080
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
