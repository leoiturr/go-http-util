package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookRequest represents an incoming request to a generated webhook URL
type WebhookRequest struct {
	Timestamp   time.Time
	Method      string
	Path        string
	Headers     http.Header
	QueryParams string
	Body        string
	StatusCode  int
	Delay       string
}

// WebhookStore holds the recent requests for each generated webhook ID
type WebhookStore struct {
	mu       sync.RWMutex
	requests map[string][]WebhookRequest
}

var webhookStore = &WebhookStore{
	requests: make(map[string][]WebhookRequest),
}

const maxRequestsPerWebhook = 50

const maxDelayPerWebhook = 90

// generateWebhookID creates a random hex string
func generateWebhookID() string {
	b := make([]byte, 8) // 16 characters
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleWebhookReceive handles any request hitting /w/:id
func HandleWebhookReceive(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.String(http.StatusBadRequest, "Missing Webhook ID")
		return
	}

	// Parse configured behavior
	statusStr := c.Query("status")
	delayStr := c.Query("delay")

	statusCode := http.StatusOK
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil && s >= 100 && s <= 599 {
			statusCode = s
		}
	}

	delayDuration := time.Duration(0)
	if delayStr != "" {
		if d, err := time.ParseDuration(delayStr); err == nil {
			// Limit delay to {maxDelayPerWebhook} seconds max
			if d > maxDelayPerWebhook * time.Second {
				d = maxDelayPerWebhook * time.Second
			}
			delayDuration = d
		}
	}

	// Read body (up to a limit, say 1MB)
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading body")
		return
	}
	// Restore body for any subsequent reading if necessary
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Record the request
	reqRecord := WebhookRequest{
		Timestamp:   time.Now(),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Headers:     c.Request.Header.Clone(),
		QueryParams: c.Request.URL.RawQuery,
		Body:        string(bodyBytes),
		StatusCode:  statusCode,
		Delay:       delayDuration.String(),
	}

	webhookStore.mu.Lock()
	reqs, exists := webhookStore.requests[id]
	if !exists {
		reqs = []WebhookRequest{}
	}
	// Prepend to list
	reqs = append([]WebhookRequest{reqRecord}, reqs...)
	if len(reqs) > maxRequestsPerWebhook {
		reqs = reqs[:maxRequestsPerWebhook]
	}
	webhookStore.requests[id] = reqs
	webhookStore.mu.Unlock()

	// Apply delay
	if delayDuration > 0 {
		time.Sleep(delayDuration)
	}

	// Send response
	c.String(statusCode, "Webhook received by DevUtils")
}

// HTMXGenerateWebhook generates a new webhook and renders the UI
func HTMXGenerateWebhook(c *gin.Context) {
	id := generateWebhookID()

	webhookStore.mu.Lock()
	webhookStore.requests[id] = []WebhookRequest{}
	webhookStore.mu.Unlock()

	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	webhookURL := fmt.Sprintf("%s://%s/w/%s", scheme, host, id)

	c.HTML(http.StatusOK, "webhook_active", gin.H{
		"ID":  id,
		"URL": webhookURL,
	})
}

// HTMXGetWebhookRequests returns the HTML list of recent requests for a given webhook
func HTMXGetWebhookRequests(c *gin.Context) {
	id := c.Param("id")

	webhookStore.mu.RLock()
	reqs, exists := webhookStore.requests[id]
	webhookStore.mu.RUnlock()

	if !exists || len(reqs) == 0 {
		c.String(http.StatusOK, `<div class="text-gray-400 text-sm italic py-4">No requests yet... Waiting for incoming webhooks.</div>`)
		return
	}

	c.HTML(http.StatusOK, "webhook_requests", gin.H{
		"Requests": reqs,
	})
}
