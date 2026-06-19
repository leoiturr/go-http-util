package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// URLRequest binds raw text input for URL operations
type URLRequest struct {
	Text string `form:"text" json:"text"`
}

// QueryParam represents a key-value parameter in a URL query string
type QueryParam struct {
	Key   string
	Value string
}

// EncodeURL encodes query parameters to be URL-safe
func EncodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": "Input is empty"})
		return
	}

	encoded := url.QueryEscape(trimmed)
	c.HTML(http.StatusOK, "url-result", gin.H{
		"Result":    encoded,
		"Operation": "Encode",
	})
}

// DecodeURL decodes percent-encoded URL query strings back to raw strings
func DecodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": "Input is empty"})
		return
	}

	decoded, err := url.QueryUnescape(trimmed)
	if err != nil {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": fmt.Sprintf("Failed to decode URL: %v", err)})
		return
	}

	c.HTML(http.StatusOK, "url-result", gin.H{
		"Result":    decoded,
		"Operation": "Decode",
	})
}

// ParseURL dissects a URL or raw query string into scheme, host, path, and detailed key-value query parameters
func ParseURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	trimmed := strings.TrimSpace(req.Text)
	if trimmed == "" {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": "Input is empty"})
		return
	}

	// Try parsing standard URL
	u, err := url.Parse(trimmed)
	if err != nil || (u.Scheme == "" && u.Host == "" && !strings.Contains(trimmed, "/")) {
		// Fallback: Try parsing as raw query parameters if it doesn't look like a full URL
		queryParams, errQ := url.ParseQuery(trimmed)
		if errQ != nil {
			c.HTML(http.StatusOK, "url-result", gin.H{"Error": fmt.Sprintf("Failed to parse URL/Query: %v", err)})
			return
		}

		var params []QueryParam
		for k, values := range queryParams {
			for _, v := range values {
				params = append(params, QueryParam{Key: k, Value: v})
			}
		}
		c.HTML(http.StatusOK, "url-result", gin.H{
			"Operation": "ParseQueryString",
			"Params":    params,
		})
		return
	}

	// Succeeded standard URL parsing
	queryParams := u.Query()
	var params []QueryParam
	for k, values := range queryParams {
		for _, v := range values {
			params = append(params, QueryParam{Key: k, Value: v})
		}
	}

	c.HTML(http.StatusOK, "url-result", gin.H{
		"Operation": "Parse",
		"Scheme":    u.Scheme,
		"Host":      u.Host,
		"Path":      u.Path,
		"Params":    params,
	})
}
