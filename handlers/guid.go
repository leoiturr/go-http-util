package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GUIDRequest binds form data for generating GUIDs
type GUIDRequest struct {
	Count  string `form:"count" json:"count"`
	Type   string `form:"type" json:"type"`
	Prefix string `form:"prefix" json:"prefix"`
}

// GenerateGUIDs handles the generation request
func GenerateGUIDs(c *gin.Context) {
	var req GUIDRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "guid-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	// Parse and clamp count between 10 and 100
	count, err := strconv.Atoi(req.Count)
	if err != nil {
		count = 10 // Default fallback
	}
	if count < 10 {
		count = 10
	}
	if count > 100 {
		count = 100
	}

	guids := make([]string, 0, count)
	var genErr error

	for i := 0; i < count; i++ {
		var val string
		var err error
		switch req.Type {
		case "v7":
			id, errV7 := uuid.NewV7()
			if errV7 != nil {
				genErr = errV7
				break
			}
			val = id.String()
		case "prefixed":
			suffix, errHex := generateRandomHex(6) // 6 bytes = 12 hex characters
			if errHex != nil {
				genErr = errHex
				break
			}
			cleanPrefix := strings.TrimSpace(req.Prefix)
			if cleanPrefix == "" {
				cleanPrefix = "GUID"
			}
			val = fmt.Sprintf("%s-%s", cleanPrefix, suffix)
		case "prefixed_guid":
			val, err = generatePrefixedGUID(req.Prefix)
			if err != nil {
				genErr = err
				break
			}
		case "v4":
			fallthrough
		default:
			val = uuid.New().String()
		}
		if genErr != nil {
			break
		}
		guids = append(guids, val)
	}

	if genErr != nil {
		c.HTML(http.StatusInternalServerError, "guid-result", gin.H{
			"Error": fmt.Sprintf("Failed to generate GUIDs: %v", genErr),
		})
		return
	}

	c.HTML(http.StatusOK, "guid-result", gin.H{
		"GUIDs":     guids,
		"Count":     count,
		"Type":      req.Type,
		"Prefix":    req.Prefix,
		"Timestamp": true,
	})
}

// generatePrefixedGUID generates a GUID matching the exact standard format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,
// where the first group of 8 characters contains the user's custom prefix.
func generatePrefixedGUID(prefix string) (string, error) {
	// Clean the prefix (only keep alphanumeric characters)
	cleanPrefix := ""
	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			cleanPrefix += string(r)
		}
	}

	// We need 8 characters for the first group
	var firstGroup string
	if len(cleanPrefix) >= 8 {
		firstGroup = cleanPrefix[:8]
	} else {
		// Pad with random hex characters
		needed := 8 - len(cleanPrefix)
		randomHex, err := generateRandomHex((needed + 1) / 2) // Generate enough bytes
		if err != nil {
			return "", err
		}
		firstGroup = cleanPrefix + randomHex[:needed]
	}

	// Generate a new UUID v4
	id := uuid.New().String()
	// Split by hyphens: 5 groups
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		return id, nil // Fallback
	}

	// Replace the first group
	parts[0] = firstGroup

	// Join back
	return strings.Join(parts, "-"), nil
}

// generateRandomHex generates a cryptographically secure random hex string of size bytes
func generateRandomHex(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
