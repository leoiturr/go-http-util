package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// JSONRequest binds input form data for JSON manipulation
type JSONRequest struct {
	Text   string `form:"text" json:"text"`
	Indent string `form:"indent" json:"indent"` // "2", "4", "tab"
}

var (
	commentBlockRegex  = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	commentLineRegex   = regexp.MustCompile(`(?m)(?:^|[^:])//.*$`)
	singleQuoteRegex   = regexp.MustCompile(`'([^'\\]*(?:\\.[^'\\]*)*)'`)
	unquotedKeyRegex   = regexp.MustCompile(`([{,]\s*)([a-zA-Z_$][a-zA-Z0-9_$-]*)\s*:`)
	trailingCommaRegex = regexp.MustCompile(`,\s*([}\]])`)
)

// cleanLooseJSONGo sanitizes JavaScript object notation / loose JSON to strict JSON
func cleanLooseJSONGo(input string) string {
	// 1. Strip block comments
	cleaned := commentBlockRegex.ReplaceAllString(input, "")

	// 2. Strip single line comments
	cleaned = commentLineRegex.ReplaceAllStringFunc(cleaned, func(match string) string {
		trimmed := strings.TrimSpace(match)
		if strings.HasPrefix(trimmed, "://") {
			return match
		}
		return ""
	})

	// 3. Convert single quotes to double quotes
	cleaned = singleQuoteRegex.ReplaceAllString(cleaned, `"$1"`)

	// 4. Wrap unquoted keys
	cleaned = unquotedKeyRegex.ReplaceAllString(cleaned, `$1"$2":`)

	// 5. Strip trailing commas
	cleaned = trailingCommaRegex.ReplaceAllString(cleaned, `$1`)

	return cleaned
}

// parseJSONLoose attempts strict JSON unmarshal first, then falls back to cleaning and unmarshaling
func parseJSONLoose(text string) (interface{}, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("input is empty")
	}

	var obj interface{}
	// Try strict unmarshal first
	if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
		return obj, nil
	}

	// Fallback to loose cleaning
	cleaned := cleanLooseJSONGo(trimmed)
	if err := json.Unmarshal([]byte(cleaned), &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON/JS Object notation: %v", err)
	}

	return obj, nil
}

// PrettifyJSONLogic handles formatting and beautifying JSON (supporting loose notation)
func PrettifyJSONLogic(text, indent string) (string, error) {
	obj, err := parseJSONLoose(text)
	if err != nil {
		return "", err
	}

	indentStr := "  "
	if indent == "4" {
		indentStr = "    "
	} else if indent == "tab" {
		indentStr = "\t"
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", indentStr)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(obj); err != nil {
		return "", fmt.Errorf("failed to format JSON: %v", err)
	}

	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// MinifyJSONLogic handles minifying / compacting JSON (supporting loose notation)
func MinifyJSONLogic(text string) (string, error) {
	obj, err := parseJSONLoose(text)
	if err != nil {
		return "", err
	}

	compactBytes, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to minify JSON: %v", err)
	}

	return string(compactBytes), nil
}

// ValidateJSONLogic checks if JSON is syntactically valid (supporting loose notation)
func ValidateJSONLogic(text string) (string, bool, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false, fmt.Errorf("input is empty")
	}

	var obj interface{}
	// Try strict JSON unmarshal
	if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
		return "JSON is valid and well-formed!", true, nil
	}

	// Try loose clean fallback
	cleaned := cleanLooseJSONGo(trimmed)
	if err := json.Unmarshal([]byte(cleaned), &obj); err == nil {
		return "Valid JavaScript Object / Loose JSON (cleaned successfully)!", true, nil
	}

	// Re-run validation on clean to get specific error
	var syntaxErr interface{}
	err := json.Unmarshal([]byte(cleaned), &syntaxErr)
	return fmt.Sprintf("Invalid JSON: %v", err), false, nil
}

// HTMXPrettifyJSON handles formatting requests from HTMX
func HTMXPrettifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	result, err := PrettifyJSONLogic(req.Text, req.Indent)
	if err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    result,
		"Operation": "Prettify",
	})
}

// V1PrettifyJSON handles JSON REST API requests to prettify JSON
func V1PrettifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := PrettifyJSONLogic(req.Text, req.Indent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    result,
		"operation": "Prettify",
	})
}

// HTMXMinifyJSON handles minifying requests from HTMX
func HTMXMinifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	result, err := MinifyJSONLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    result,
		"Operation": "Minify",
	})
}

// V1MinifyJSON handles JSON REST API requests to minify JSON
func V1MinifyJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := MinifyJSONLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    result,
		"operation": "Minify",
	})
}

// HTMXValidateJSON handles validating requests from HTMX
func HTMXValidateJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "json-result", gin.H{"Error": "Invalid request"})
		return
	}

	msg, isValid, err := ValidateJSONLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "json-result", gin.H{"Error": err.Error()})
		return
	}

	if !isValid {
		c.HTML(http.StatusOK, "json-result", gin.H{
			"Error":     msg,
			"Operation": "Validate",
			"IsValid":   false,
		})
		return
	}

	c.HTML(http.StatusOK, "json-result", gin.H{
		"Result":    msg,
		"Operation": "Validate",
		"IsValid":   true,
	})
}

// V1ValidateJSON handles JSON REST API requests to validate JSON
func V1ValidateJSON(c *gin.Context) {
	var req JSONRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	msg, isValid, err := ValidateJSONLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    msg,
		"is_valid":  isValid,
		"operation": "Validate",
	})
}
