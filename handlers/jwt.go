package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JWTRequest binds token payload
type JWTRequest struct {
	Token string `form:"token" json:"token"`
}

// DecodeJWT decodes a JWT without checking signatures (client-facing details tool)
func DecodeJWT(c *gin.Context) {
	var req JWTRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "jwt-result", gin.H{"Error": "Invalid request parameters"})
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		c.HTML(http.StatusOK, "jwt-result", gin.H{"Error": "JWT token input is empty"})
		return
	}

	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		c.HTML(http.StatusOK, "jwt-result", gin.H{"Error": "Invalid JWT format. A valid token must have exactly three segments separated by dots (header.payload.signature)."})
		return
	}

	headerDec, err := decodeBase64URL(parts[0])
	if err != nil {
		c.HTML(http.StatusOK, "jwt-result", gin.H{"Error": fmt.Sprintf("Failed to decode token header: %v", err)})
		return
	}

	payloadDec, err := decodeBase64URL(parts[1])
	if err != nil {
		c.HTML(http.StatusOK, "jwt-result", gin.H{"Error": fmt.Sprintf("Failed to decode token payload: %v", err)})
		return
	}

	// Prettify JSON header
	var headerObj interface{}
	var headerPretty string
	if err := json.Unmarshal(headerDec, &headerObj); err == nil {
		if headerBytes, err := json.MarshalIndent(headerObj, "", "  "); err == nil {
			headerPretty = string(headerBytes)
		}
	}
	if headerPretty == "" {
		headerPretty = string(headerDec)
	}

	// Prettify JSON payload
	var payloadObj map[string]interface{}
	var payloadPretty string
	if err := json.Unmarshal(payloadDec, &payloadObj); err == nil {
		if payloadBytes, err := json.MarshalIndent(payloadObj, "", "  "); err == nil {
			payloadPretty = string(payloadBytes)
		}
	}
	if payloadPretty == "" {
		payloadPretty = string(payloadDec)
	}

	// Extract standard claims: exp, iat, nbf, sub, iss
	var hasExp bool
	var isExpired bool
	var expTimeStr string
	var expInStr string
	var issuedAtStr string

	if payloadObj != nil {
		// exp (Expiration Time)
		if expVal, ok := payloadObj["exp"]; ok {
			var expFloat float64
			switch val := expVal.(type) {
			case float64:
				expFloat = val
			case int64:
				expFloat = float64(val)
			}

			if expFloat > 0 {
				hasExp = true
				expTime := time.Unix(int64(expFloat), 0)
				expTimeStr = expTime.Format("2006-01-02 15:04:05 MST")

				now := time.Now()
				if now.After(expTime) {
					isExpired = true
					duration := now.Sub(expTime).Round(time.Second)
					expInStr = fmt.Sprintf("Expired %v ago", duration)
				} else {
					isExpired = false
					duration := expTime.Sub(now).Round(time.Second)
					expInStr = fmt.Sprintf("Expires in %v", duration)
				}
			}
		}

		// iat (Issued At)
		if iatVal, ok := payloadObj["iat"]; ok {
			var iatFloat float64
			switch val := iatVal.(type) {
			case float64:
				iatFloat = val
			case int64:
				iatFloat = float64(val)
			}
			if iatFloat > 0 {
				iatTime := time.Unix(int64(iatFloat), 0)
				issuedAtStr = iatTime.Format("2006-01-02 15:04:05 MST")
			}
		}
	}

	c.HTML(http.StatusOK, "jwt-result", gin.H{
		"Header":    headerPretty,
		"Payload":   payloadPretty,
		"Signature": parts[2],
		"HasExp":    hasExp,
		"IsExpired": isExpired,
		"ExpTime":   expTimeStr,
		"ExpIn":     expInStr,
		"IssuedAt":  issuedAtStr,
	})
}

// decodeBase64URL decodes base64url padding-less encoded data (standard JWT strings)
func decodeBase64URL(s string) ([]byte, error) {
	// Add padding if missing
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
