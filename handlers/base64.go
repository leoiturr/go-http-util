package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Base64Request binds form data or JSON payloads
type Base64Request struct {
	Text string `form:"text" json:"text"`
}

// EncodeBase64Logic encodes plain text to Base64
func EncodeBase64Logic(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// DecodeBase64Logic decodes Base64 (standard or URL-safe) to plain text
func DecodeBase64Logic(text string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		// Try URL-safe base64 decoding as fallback
		decodedBytes, err = base64.URLEncoding.DecodeString(text)
		if err != nil {
			return "", err
		}
	}
	return string(decodedBytes), nil
}

// HTMXEncodeBase64 handles encoding requests from HTMX
func HTMXEncodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "base64-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	encoded := EncodeBase64Logic(req.Text)

	c.HTML(http.StatusOK, "base64-result", gin.H{
		"Result":    encoded,
		"Original":  req.Text,
		"Operation": "Encode",
	})
}

// V1EncodeBase64 handles JSON REST API requests to encode base64
func V1EncodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	encoded := EncodeBase64Logic(req.Text)

	c.JSON(http.StatusOK, gin.H{
		"result":    encoded,
		"original":  req.Text,
		"operation": "Encode",
	})
}

// HTMXDecodeBase64 handles decoding requests from HTMX
func HTMXDecodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "base64-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	decoded, err := DecodeBase64Logic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "base64-result", gin.H{
			"Error":     "Invalid Base64 string. Please verify the input.",
			"Original":  req.Text,
			"Operation": "Decode",
		})
		return
	}

	c.HTML(http.StatusOK, "base64-result", gin.H{
		"Result":    decoded,
		"Original":  req.Text,
		"Operation": "Decode",
	})
}

// V1DecodeBase64 handles JSON REST API requests to decode base64
func V1DecodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	decoded, err := DecodeBase64Logic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid Base64 string. Please verify the input.",
			"original":  req.Text,
			"operation": "Decode",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    decoded,
		"original":  req.Text,
		"operation": "Decode",
	})
}
