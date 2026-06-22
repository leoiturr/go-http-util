package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
)

// YAMLRequest binds input form/json data for YAML manipulation
type YAMLRequest struct {
	Text   string `form:"text" json:"text"`
	Indent string `form:"indent" json:"indent"` // "2", "4"
	Quotes string `form:"quotes" json:"quotes"` // "default", "single", "double"
}

// JSONToYAMLLogic converts JSON to YAML
func JSONToYAMLLogic(text string, indentStr string, quotes string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("input is empty")
	}

	cleaned := cleanLooseJSONGo(trimmed)
	yamlBytes, err := yaml.JSONToYAML([]byte(cleaned))
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	var obj interface{}
	if err := yaml.UnmarshalWithOptions(yamlBytes, &obj, yaml.UseOrderedMap()); err != nil {
		return "", fmt.Errorf("failed to process intermediate YAML: %v", err)
	}

	indentSpaces := 2
	if indentStr == "4" {
		indentSpaces = 4
	}

	var opts []yaml.EncodeOption
	opts = append(opts, yaml.Indent(indentSpaces))

	if quotes == "single" {
		opts = append(opts, yaml.UseSingleQuote(true))
	} else if quotes == "double" {
		opts = append(opts, yaml.UseSingleQuote(false))
	}

	resBytes, err := yaml.MarshalWithOptions(obj, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to format YAML: %v", err)
	}

	return string(resBytes), nil
}

// YAMLToJSONLogic converts YAML to JSON preserving key order
func YAMLToJSONLogic(text string, indentStr string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("input is empty")
	}

	var obj interface{}
	if err := yaml.UnmarshalWithOptions([]byte(trimmed), &obj, yaml.UseOrderedMap()); err != nil {
		return "", fmt.Errorf("invalid YAML:\n%s", yaml.FormatError(err, false, true))
	}

	compactJSON, err := yaml.MarshalWithOptions(obj, yaml.JSON())
	if err != nil {
		return "", fmt.Errorf("failed to encode JSON: %v", err)
	}

	indentVal := "  "
	if indentStr == "4" {
		indentVal = "    "
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, compactJSON, "", indentVal); err != nil {
		return "", fmt.Errorf("failed to format JSON: %v", err)
	}

	return prettyJSON.String(), nil
}

// PrettifyYAMLLogic formats YAML
func PrettifyYAMLLogic(text string, indentStr string, quotes string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("input is empty")
	}

	var obj interface{}
	// Unmarshal using UseOrderedMap() to preserve key order recursively
	if err := yaml.UnmarshalWithOptions([]byte(trimmed), &obj, yaml.UseOrderedMap()); err != nil {
		return "", fmt.Errorf("invalid YAML:\n%s", yaml.FormatError(err, false, true))
	}

	indentSpaces := 2
	if indentStr == "4" {
		indentSpaces = 4
	}

	var opts []yaml.EncodeOption
	opts = append(opts, yaml.Indent(indentSpaces))

	if quotes == "single" {
		opts = append(opts, yaml.UseSingleQuote(true))
	} else if quotes == "double" {
		opts = append(opts, yaml.UseSingleQuote(false))
	}

	yamlBytes, err := yaml.MarshalWithOptions(obj, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to format YAML: %v", err)
	}

	return string(yamlBytes), nil
}

// ValidateYAMLLogic validates YAML and returns message, isValid, error
func ValidateYAMLLogic(text string) (string, bool, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false, fmt.Errorf("input is empty")
	}

	var obj interface{}
	if err := yaml.UnmarshalWithOptions([]byte(trimmed), &obj, yaml.UseOrderedMap()); err != nil {
		formattedErr := yaml.FormatError(err, false, true)
		return fmt.Sprintf("Invalid YAML:\n%s", formattedErr), false, nil
	}

	return "YAML is valid and well-formed!", true, nil
}

// HTMXJSONToYAML handles HTMX requests to convert JSON to YAML
func HTMXJSONToYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "yaml-result", gin.H{"Error": "Invalid request", "Operation": "JSON to YAML"})
		return
	}

	result, err := JSONToYAMLLogic(req.Text, req.Indent, req.Quotes)
	if err != nil {
		c.HTML(http.StatusOK, "yaml-result", gin.H{"Error": err.Error(), "Operation": "JSON to YAML"})
		return
	}

	c.HTML(http.StatusOK, "yaml-result", gin.H{
		"Result":    result,
		"Operation": "JSON to YAML",
	})
}

// HTMXYAMLToJSON handles HTMX requests to convert YAML to JSON
func HTMXYAMLToJSON(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "yaml-result", gin.H{"Error": "Invalid request", "Operation": "Convert to JSON"})
		return
	}

	result, err := YAMLToJSONLogic(req.Text, req.Indent)
	if err != nil {
		c.HTML(http.StatusOK, "yaml-result", gin.H{"Error": err.Error(), "Operation": "Convert to JSON"})
		return
	}

	c.HTML(http.StatusOK, "yaml-result", gin.H{
		"Result":    result,
		"Operation": "Convert to JSON",
	})
}

// HTMXPrettifyYAML handles HTMX requests to prettify YAML
func HTMXPrettifyYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "yaml-result", gin.H{"Error": "Invalid request", "Operation": "Format"})
		return
	}

	result, err := PrettifyYAMLLogic(req.Text, req.Indent, req.Quotes)
	if err != nil {
		c.HTML(http.StatusOK, "yaml-result", gin.H{"Error": err.Error(), "Operation": "Format"})
		return
	}

	c.HTML(http.StatusOK, "yaml-result", gin.H{
		"Result":    result,
		"Operation": "Format",
	})
}

// HTMXValidateYAML handles HTMX requests to validate YAML
func HTMXValidateYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "yaml-result", gin.H{"Error": "Invalid request", "Operation": "Validate"})
		return
	}

	msg, isValid, err := ValidateYAMLLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "yaml-result", gin.H{"Error": err.Error(), "Operation": "Validate"})
		return
	}

	if !isValid {
		c.HTML(http.StatusOK, "yaml-result", gin.H{
			"Error":     msg,
			"Operation": "Validate",
			"IsValid":   false,
		})
		return
	}

	c.HTML(http.StatusOK, "yaml-result", gin.H{
		"Result":    msg,
		"Operation": "Validate",
		"IsValid":   true,
	})
}

// V1JSONToYAML handles REST API requests to convert JSON to YAML
func V1JSONToYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := JSONToYAMLLogic(req.Text, req.Indent, req.Quotes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    result,
		"operation": "json_to_yaml",
	})
}

// V1YAMLToJSON handles REST API requests to convert YAML to JSON
func V1YAMLToJSON(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := YAMLToJSONLogic(req.Text, req.Indent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    result,
		"operation": "yaml_to_json",
	})
}

// V1PrettifyYAML handles REST API requests to prettify YAML
func V1PrettifyYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := PrettifyYAMLLogic(req.Text, req.Indent, req.Quotes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    result,
		"operation": "prettify",
	})
}

// V1ValidateYAML handles REST API requests to validate YAML
func V1ValidateYAML(c *gin.Context) {
	var req YAMLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	msg, isValid, err := ValidateYAMLLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    msg,
		"is_valid":  isValid,
		"operation": "validate",
	})
}
