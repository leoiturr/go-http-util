package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"gin-example/handlers"
	"gin-example/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Disable trusting all proxies to secure client IP headers and silence the warning
	r.SetTrustedProxies(nil)

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

	// Health Check API
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Main web interface route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{})
	})

	// Set up rate limiter middleware from environment variables with safe defaults
	// Default: 2 requests per second (RPS) with a burst of 4 requests
	rps := 2.0
	burst := 4.0

	if rpsEnv := os.Getenv("RATE_LIMIT_RPS"); rpsEnv != "" {
		if val, err := strconv.ParseFloat(rpsEnv, 64); err == nil {
			rps = val
		}
	}
	if burstEnv := os.Getenv("RATE_LIMIT_BURST"); burstEnv != "" {
		if val, err := strconv.ParseFloat(burstEnv, 64); err == nil {
			burst = val
		}
	}

	limiter := middleware.NewRateLimiter(rps, burst)

	// API Route Group with Rate Limiting
	api := r.Group("/api")
	api.Use(limiter.Limit())
	{
		// Base64 Converter API
		api.POST("/base64/encode", handlers.EncodeBase64)
		api.POST("/base64/decode", handlers.DecodeBase64)

		// GUID Generator API
		api.POST("/guid/generate", handlers.GenerateGUIDs)

		// QR Code Generator API
		api.POST("/qrcode/generate", handlers.GenerateQRCode)
	}

	// Start Gin server on the configured port (Render uses the PORT environment variable)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
