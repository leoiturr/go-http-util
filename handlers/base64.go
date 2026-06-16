package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Base64Request binds form data from HTMX
type Base64Request struct {
	Text string `form:"text" json:"text"`
}

// EncodeBase64 handles encoding plain text to Base64
func EncodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "base64-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(req.Text))

	c.HTML(http.StatusOK, "base64-result", gin.H{
		"Result":    encoded,
		"Original":  req.Text,
		"Operation": "Encode",
	})
}

// DecodeBase64 handles decoding Base64 to plain text
func DecodeBase64(c *gin.Context) {
	var req Base64Request
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "base64-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(req.Text)
	if err != nil {
		// Try URL-safe base64 decoding as fallback
		decodedBytes, err = base64.URLEncoding.DecodeString(req.Text)
		if err != nil {
			c.HTML(http.StatusOK, "base64-result", gin.H{
				"Error":     "Invalid Base64 string. Please verify the input.",
				"Original":  req.Text,
				"Operation": "Decode",
			})
			return
		}
	}

	c.HTML(http.StatusOK, "base64-result", gin.H{
		"Result":    string(decodedBytes),
		"Original":  req.Text,
		"Operation": "Decode",
	})
}
