package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// WhatsAppRequest contains the recipient and optional pre-filled message.
type WhatsAppRequest struct {
	Phone   string `form:"phone" json:"phone"`
	Message string `form:"message" json:"message"`
}

// WhatsAppResult is shared by the HTMX fragment and JSON success response.
// Error remains present for template branches on both success and failure.
type WhatsAppResult struct {
	Phone          string `json:"phone"`
	FormattedPhone string `json:"formatted_phone"`
	Message        string `json:"message"`
	Link           string `json:"link"`
	DeepLink       string `json:"deep_link"`
	HasMessage     bool   `json:"has_message"`
	Error          string `json:"error"`
}

// BuildWhatsAppLink validates an international phone number and creates
// WhatsApp click-to-chat links without requiring a saved contact.
func BuildWhatsAppLink(phone, message string) (WhatsAppResult, error) {
	result := WhatsAppResult{Message: strings.TrimSpace(message)}

	input := strings.TrimSpace(phone)
	if input == "" {
		err := fmt.Errorf("enter a phone number with its country code, e.g. +1 415 555 0132")
		result.Error = err.Error()
		return result, err
	}

	var digits strings.Builder
	for _, r := range input {
		switch r {
		case ' ', '-', '.', '(', ')':
			continue
		default:
			digits.WriteRune(r)
		}
	}
	normalized := digits.String()
	if strings.HasPrefix(normalized, "+") {
		normalized = strings.TrimPrefix(normalized, "+")
	}
	if strings.HasPrefix(normalized, "00") {
		normalized = strings.TrimPrefix(normalized, "00")
	}

	for _, r := range normalized {
		if r < '0' || r > '9' {
			err := fmt.Errorf("phone number may contain only digits and common separators (spaces, dashes, dots, parentheses)")
			result.Error = err.Error()
			return result, err
		}
	}
	if len(normalized) < 7 || len(normalized) > 15 {
		err := fmt.Errorf("phone number must contain 7 to 15 digits including the country code, e.g. +1 415 555 0132")
		result.Error = err.Error()
		return result, err
	}

	result.Phone = normalized
	result.FormattedPhone = "+" + normalized
	result.HasMessage = result.Message != ""

	result.Link = "https://wa.me/" + normalized
	result.DeepLink = "whatsapp://send?phone=" + normalized
	if result.HasMessage {
		// Percent-encode spaces explicitly; QueryEscape's "+" is ambiguous
		// once the link is pasted or shared outside a form context.
		encodedMessage := strings.ReplaceAll(url.QueryEscape(result.Message), "+", "%20")
		result.Link += "?text=" + encodedMessage
		result.DeepLink += "&text=" + encodedMessage
	}

	return result, nil
}

// HTMXGenerateWhatsAppLink renders the WhatsApp link result fragment.
func HTMXGenerateWhatsAppLink(c *gin.Context) {
	var req WhatsAppRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "whatsapp-result", WhatsAppResult{
			Error: "Invalid request parameters",
		})
		return
	}

	result, err := BuildWhatsAppLink(req.Phone, req.Message)
	if err != nil {
		c.HTML(http.StatusBadRequest, "whatsapp-result", result)
		return
	}

	c.HTML(http.StatusOK, "whatsapp-result", result)
}

// V1GenerateWhatsAppLink returns a click-to-chat link as JSON.
func V1GenerateWhatsAppLink(c *gin.Context) {
	var req WhatsAppRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	result, err := BuildWhatsAppLink(req.Phone, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
