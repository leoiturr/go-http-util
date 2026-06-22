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
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParsedURLResult contains the results of dissecting a URL/query string
type ParsedURLResult struct {
	Operation string       `json:"operation"`
	Scheme    string       `json:"scheme,omitempty"`
	Host      string       `json:"host,omitempty"`
	Path      string       `json:"path,omitempty"`
	Params    []QueryParam `json:"params"`
}

// EncodeURLLogic URL-encodes a string
func EncodeURLLogic(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("input is empty")
	}
	return url.QueryEscape(trimmed), nil
}

// DecodeURLLogic URL-decodes a percent-encoded string
func DecodeURLLogic(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("input is empty")
	}
	decoded, err := url.QueryUnescape(trimmed)
	if err != nil {
		return "", fmt.Errorf("failed to decode URL: %v", err)
	}
	return decoded, nil
}

// ParseURLLogic dissects a URL or query string
func ParseURLLogic(text string) (ParsedURLResult, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ParsedURLResult{}, fmt.Errorf("input is empty")
	}

	// Try parsing standard URL
	u, err := url.Parse(trimmed)
	if err != nil || (u.Scheme == "" && u.Host == "" && !strings.Contains(trimmed, "/")) {
		// Fallback: Try parsing as raw query parameters if it doesn't look like a full URL
		queryParams, errQ := url.ParseQuery(trimmed)
		if errQ != nil {
			return ParsedURLResult{}, fmt.Errorf("failed to parse URL/Query: %v", err)
		}

		var params []QueryParam
		for k, values := range queryParams {
			for _, v := range values {
				params = append(params, QueryParam{Key: k, Value: v})
			}
		}
		return ParsedURLResult{
			Operation: "ParseQueryString",
			Params:    params,
		}, nil
	}

	// Succeeded standard URL parsing
	queryParams := u.Query()
	var params []QueryParam
	for k, values := range queryParams {
		for _, v := range values {
			params = append(params, QueryParam{Key: k, Value: v})
		}
	}

	return ParsedURLResult{
		Operation: "Parse",
		Scheme:    u.Scheme,
		Host:      u.Host,
		Path:      u.Path,
		Params:    params,
	}, nil
}

// HTMXEncodeURL handles encode requests from HTMX
func HTMXEncodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	encoded, err := EncodeURLLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "url-result", gin.H{
		"Result":    encoded,
		"Operation": "Encode",
	})
}

// V1EncodeURL handles JSON REST API requests to encode a URL
func V1EncodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	encoded, err := EncodeURLLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    encoded,
		"operation": "Encode",
	})
}

// HTMXDecodeURL handles decode requests from HTMX
func HTMXDecodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	decoded, err := DecodeURLLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "url-result", gin.H{
		"Result":    decoded,
		"Operation": "Decode",
	})
}

// V1DecodeURL handles JSON REST API requests to decode a URL
func V1DecodeURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	decoded, err := DecodeURLLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    decoded,
		"operation": "Decode",
	})
}

// HTMXParseURL handles parse requests from HTMX
func HTMXParseURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "url-result", gin.H{"Error": "Invalid request"})
		return
	}

	res, err := ParseURLLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "url-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "url-result", gin.H{
		"Operation": res.Operation,
		"Scheme":    res.Scheme,
		"Host":      res.Host,
		"Path":      res.Path,
		"Params":    res.Params,
	})
}

// V1ParseURL handles JSON REST API requests to parse a URL
func V1ParseURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	res, err := ParseURLLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
