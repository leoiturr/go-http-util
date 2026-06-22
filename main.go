package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gin-example/handlers"
	"gin-example/middleware"
	"github.com/gin-gonic/gin"
)

var appVersion = getVersion()

func getVersion() string {
	// Check Render.com injected environment variable first
	if renderCommit := os.Getenv("RENDER_GIT_COMMIT"); renderCommit != "" {
		if len(renderCommit) > 7 {
			return "v1.0.1-" + renderCommit[:7]
		}
		return "v1.0.1-" + renderCommit
	}

	// Fallback to local git CLI command execution (useful in local dev)
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		commit := strings.TrimSpace(string(out))
		if commit != "" {
			return "v1.0.2-" + commit
		}
	}
	return "v1.0.2"
}

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

	// Docs Web Route
	r.GET("/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "docs", gin.H{"Version": appVersion})
	})

	// Main web interface route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{"Version": appVersion})
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

	// HTMX API Route Group with Rate Limiting
	htmxGroup := r.Group("/api/htmx")
	htmxGroup.Use(limiter.Limit())
	{
		// Base64 Converter API
		htmxGroup.POST("/base64/encode", handlers.HTMXEncodeBase64)
		htmxGroup.POST("/base64/decode", handlers.HTMXDecodeBase64)

		// GUID Generator API
		htmxGroup.POST("/guid/generate", handlers.HTMXGenerateGUIDs)

		// QR Code Generator API
		htmxGroup.POST("/qrcode/generate", handlers.HTMXGenerateQRCode)

		// JSON Prettifier / Minifier / Validator API
		htmxGroup.POST("/json/prettify", handlers.HTMXPrettifyJSON)
		htmxGroup.POST("/json/minify", handlers.HTMXMinifyJSON)
		htmxGroup.POST("/json/validate", handlers.HTMXValidateJSON)

		// URL Encoder / Decoder / Parser API
		htmxGroup.POST("/url/encode", handlers.HTMXEncodeURL)
		htmxGroup.POST("/url/decode", handlers.HTMXDecodeURL)
		htmxGroup.POST("/url/parse", handlers.HTMXParseURL)

		// JWT Debugger API
		htmxGroup.POST("/jwt/decode", handlers.HTMXDecodeJWT)

		// Epoch / Unix Timestamp Converter API
		htmxGroup.POST("/epoch/convert", handlers.HTMXConvertEpoch)

		// YAML Tools API (JSON to YAML, YAML to JSON, YAML Formatter, YAML Validator)
		htmxGroup.POST("/yaml/json2yaml", handlers.HTMXJSONToYAML)
		htmxGroup.POST("/yaml/yaml2json", handlers.HTMXYAMLToJSON)
		htmxGroup.POST("/yaml/prettify", handlers.HTMXPrettifyYAML)
		htmxGroup.POST("/yaml/validate", handlers.HTMXValidateYAML)
	}

	// REST v1 JSON API Route Group with Rate Limiting
	v1Group := r.Group("/api/v1")
	v1Group.Use(limiter.Limit())
	{
		// Base64 Converter API
		v1Group.POST("/base64/encode", handlers.V1EncodeBase64)
		v1Group.POST("/base64/decode", handlers.V1DecodeBase64)

		// GUID Generator API
		v1Group.POST("/guid/generate", handlers.V1GenerateGUIDs)

		// QR Code Generator API
		v1Group.POST("/qrcode/generate", handlers.V1GenerateQRCode)

		// JSON Prettifier / Minifier / Validator API
		v1Group.POST("/json/prettify", handlers.V1PrettifyJSON)
		v1Group.POST("/json/minify", handlers.V1MinifyJSON)
		v1Group.POST("/json/validate", handlers.V1ValidateJSON)

		// URL Encoder / Decoder / Parser API
		v1Group.POST("/url/encode", handlers.V1EncodeURL)
		v1Group.POST("/url/decode", handlers.V1DecodeURL)
		v1Group.POST("/url/parse", handlers.V1ParseURL)

		// JWT Debugger API
		v1Group.POST("/jwt/decode", handlers.V1DecodeJWT)

		// Epoch / Unix Timestamp Converter API
		v1Group.POST("/epoch/convert", handlers.V1ConvertEpoch)

		// YAML Tools API (JSON to YAML, YAML to JSON, YAML Formatter, YAML Validator)
		v1Group.POST("/yaml/json2yaml", handlers.V1JSONToYAML)
		v1Group.POST("/yaml/yaml2json", handlers.V1YAMLToJSON)
		v1Group.POST("/yaml/prettify", handlers.V1PrettifyYAML)
		v1Group.POST("/yaml/validate", handlers.V1ValidateYAML)
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
