package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EpochRequest binds epoch integer or standard datetime value
type EpochRequest struct {
	Value string `form:"value" json:"value"`
}

// ConvertEpoch converts epoch integer strings or date strings to various formats
func ConvertEpoch(c *gin.Context) {
	var req EpochRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "epoch-result", gin.H{"Error": "Invalid request parameters"})
		return
	}

	val := strings.TrimSpace(req.Value)
	if val == "" {
		// Fallback to current Unix epoch
		val = strconv.FormatInt(time.Now().Unix(), 10)
	}

	var targetTime time.Time
	var parsed bool

	// 1. Try parsing as an integer (Unix timestamp in seconds or milliseconds)
	if num, err := strconv.ParseInt(val, 10, 64); err == nil {
		// Milliseconds vs Seconds: if digit length > 11, treat as milliseconds
		if len(val) > 11 {
			targetTime = time.UnixMilli(num)
		} else {
			targetTime = time.Unix(num, 0)
		}
		parsed = true
	} else {
		// 2. Try parsing as human-readable datetime formats
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04",
			"2006-01-02",
			time.RFC3339,
			time.RFC1123,
			time.ANSIC,
		}

		for _, layout := range layouts {
			// Try parsing in local timezone first
			if t, errLayout := time.ParseInLocation(layout, val, time.Local); errLayout == nil {
				targetTime = t
				parsed = true
				break
			}
			// Fallback: Try parsing in UTC
			if t, errLayout := time.Parse(layout, val); errLayout == nil {
				targetTime = t
				parsed = true
				break
			}
		}
	}

	if !parsed {
		c.HTML(http.StatusOK, "epoch-result", gin.H{
			"Error": "Could not parse date/time value. We support Unix epoch integers (seconds or milliseconds), ISO8601 strings, YYYY-MM-DD HH:MM:SS, and common RFC formats.",
		})
		return
	}

	// Prepare result metrics
	seconds := targetTime.Unix()
	millis := targetTime.UnixMilli()
	utcStr := targetTime.UTC().Format("Mon, 02 Jan 2006 15:04:05 UTC")
	localStr := targetTime.Local().Format("2006-01-02 15:04:05 -0700 MST")

	// Relative representation
	var relative string
	now := time.Now()
	if targetTime.Before(now) {
		diff := now.Sub(targetTime).Round(time.Second)
		if diff < time.Minute {
			relative = fmt.Sprintf("%d seconds ago", int(diff.Seconds()))
		} else if diff < time.Hour {
			relative = fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			relative = fmt.Sprintf("%d hours ago", int(diff.Hours()))
		} else {
			relative = fmt.Sprintf("%d days ago", int(diff.Hours()/24))
		}
	} else {
		diff := targetTime.Sub(now).Round(time.Second)
		if diff < time.Minute {
			relative = fmt.Sprintf("in %d seconds", int(diff.Seconds()))
		} else if diff < time.Hour {
			relative = fmt.Sprintf("in %d minutes", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			relative = fmt.Sprintf("in %d hours", int(diff.Hours()))
		} else {
			relative = fmt.Sprintf("in %d days", int(diff.Hours()/24))
		}
	}

	c.HTML(http.StatusOK, "epoch-result", gin.H{
		"Seconds":  seconds,
		"Millis":   millis,
		"UTC":      utcStr,
		"Local":    localStr,
		"Relative": relative,
		"Value":    val,
	})
}
