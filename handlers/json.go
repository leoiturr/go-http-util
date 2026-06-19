package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JSONRequest binds input form data for JSON manipulation
type JSONRequest struct {
	Text   string `form:"text" json:"text"`
	Indent string `form:"indent" json:"indent"` // "2", "4", "tab"
}

// PrettifyJSON handles formatting and beautifying JSON
func PrettifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": "Input is empty"})
		return
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": fmt.Sprintf("Invalid JSON: %v", err)})
		return
	}

	indentStr := "  "
	if req.Indent == "4" {
		indentStr = "    "
	} else if req.Indent == "tab" {
		indentStr = "\t"
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", indentStr)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(obj); err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": fmt.Sprintf("Failed to format JSON: %v", err)})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    strings.TrimSuffix(buf.String(), "\n"),
		"Operation": "Prettify",
	})
}

// MinifyJSON handles minifying / compacting JSON
func MinifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": "Input is empty"})
		return
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": fmt.Sprintf("Invalid JSON: %v", err)})
		return
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(trimmed)); err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": fmt.Sprintf("Failed to minify JSON: %v", err)})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    buf.String(),
		"Operation": "Minify",
	})
}

// ValidateJSON checks if JSON is syntactically valid
func ValidateJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": "Input is empty"})
		return
	}

	var obj interface{}
	err := json.Unmarshal([]byte(trimmed), &obj)
	if err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{
			"Error":     fmt.Sprintf("Invalid JSON: %v", err),
			"Operation": "Validate",
			"IsValid":   false,
		})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    "JSON is valid and well-formed!",
		"Operation": "Validate",
		"IsValid":   true,
	})
}
