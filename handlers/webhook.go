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

// webhookSession holds the recent requests (and liveness metadata) for a
// single generated webhook ID.
type webhookSession struct {
	Requests  []WebhookRequest
	CreatedAt time.Time
	LastSeen  time.Time
}

// WebhookStore holds the recent requests for each generated webhook ID. The
// store is intentionally bounded: sessions are pruned by idle time-to-live and
// by global capacity caps so that an always-on deployment cannot grow memory
// without limit.
type WebhookStore struct {
	mu           sync.RWMutex
	requests     map[string]*webhookSession
	ttl          time.Duration
	maxSessions  int
	maxBodyBytes int64
	now          func() time.Time
}

const (
	defaultWebhookTTL          = 24 * time.Hour
	defaultMaxWebhookSessions  = 500
	defaultMaxWebhookBodyBytes = 10 << 20 // 10 MiB of stored bodies across all sessions
)

// NewWebhookStore creates a bounded webhook store. The clock is injectable via
// the returned store's now field for deterministic tests.
func NewWebhookStore(ttl time.Duration, maxSessions int, maxBodyBytes int64) *WebhookStore {
	return &WebhookStore{
		requests:     make(map[string]*webhookSession),
		ttl:          ttl,
		maxSessions:  maxSessions,
		maxBodyBytes: maxBodyBytes,
		now:          time.Now,
	}
}

var webhookStore = NewWebhookStore(
	defaultWebhookTTL,
	defaultMaxWebhookSessions,
	defaultMaxWebhookBodyBytes,
)

const maxRequestsPerWebhook = 50

const maxDelayPerWebhook = 90

// generateWebhookID creates a random hex string
func generateWebhookID() string {
	b := make([]byte, 8) // 16 characters
	rand.Read(b)
	return hex.EncodeToString(b)
}

// register creates an empty session for a freshly generated webhook ID.
func (s *WebhookStore) register(id string) {
	now := s.now()
	s.mu.Lock()
	s.requests[id] = &webhookSession{
		Requests:  []WebhookRequest{},
		CreatedAt: now,
		LastSeen:  now,
	}
	s.mu.Unlock()
}

// record prepends a request to the session for id, creating the session on
// first use and refreshing its LastSeen timestamp.
func (s *WebhookStore) record(id string, req WebhookRequest) {
	now := s.now()
	s.mu.Lock()
	sess, exists := s.requests[id]
	if !exists {
		sess = &webhookSession{
			Requests:  []WebhookRequest{},
			CreatedAt: now,
		}
	}
	sess.LastSeen = now
	sess.Requests = append([]WebhookRequest{req}, sess.Requests...)
	if len(sess.Requests) > maxRequestsPerWebhook {
		sess.Requests = sess.Requests[:maxRequestsPerWebhook]
	}
	s.requests[id] = sess
	s.mu.Unlock()
}

// get returns a copy of the stored requests for id, refreshing LastSeen so
// that an actively polled session does not expire. The second return value
// reports whether the session still exists.
func (s *WebhookStore) get(id string) ([]WebhookRequest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, exists := s.requests[id]
	if !exists {
		return nil, false
	}
	sess.LastSeen = s.now()
	return append([]WebhookRequest(nil), sess.Requests...), true
}

// evict removes sessions that exceed the idle TTL, then trims the oldest
// sessions until the session-count and stored-body caps are satisfied.
func (s *WebhookStore) evict() {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, sess := range s.requests {
		if now.Sub(sess.LastSeen) > s.ttl {
			delete(s.requests, id)
		}
	}

	for len(s.requests) > s.maxSessions {
		delete(s.requests, s.oldestIDLocked())
	}

	for s.totalBodyBytesLocked() > s.maxBodyBytes {
		delete(s.requests, s.oldestIDLocked())
	}
}

// oldestIDLocked returns the ID of the session least recently seen. Callers
// must hold the write lock.
func (s *WebhookStore) oldestIDLocked() string {
	var oldestID string
	var oldest time.Time
	for id, sess := range s.requests {
		if oldestID == "" || sess.LastSeen.Before(oldest) {
			oldestID = id
			oldest = sess.LastSeen
		}
	}
	return oldestID
}

// totalBodyBytesLocked sums the stored request body sizes. Callers must hold
// the write lock.
func (s *WebhookStore) totalBodyBytesLocked() int64 {
	var total int64
	for _, sess := range s.requests {
		for _, req := range sess.Requests {
			total += int64(len(req.Body))
		}
	}
	return total
}

// ConfigureWebhookStore overrides the global store's retention limits. Values
// of zero or less leave the existing limit unchanged.
func ConfigureWebhookStore(ttl time.Duration, maxSessions int, maxBodyBytes int64) {
	webhookStore.mu.Lock()
	if ttl > 0 {
		webhookStore.ttl = ttl
	}
	if maxSessions > 0 {
		webhookStore.maxSessions = maxSessions
	}
	if maxBodyBytes > 0 {
		webhookStore.maxBodyBytes = maxBodyBytes
	}
	webhookStore.mu.Unlock()
}

// EvictWebhooks prunes expired or over-capacity webhook sessions. Intended to
// be called periodically from a background goroutine so that the in-memory
// store stays bounded on long-running (always-on) deployments.
func EvictWebhooks() {
	webhookStore.evict()
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
			if d > maxDelayPerWebhook*time.Second {
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

	webhookStore.record(id, reqRecord)

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

	webhookStore.register(id)

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

	reqs, exists := webhookStore.get(id)

	if !exists {
		c.HTML(http.StatusOK, "webhook_expired", gin.H{})
		return
	}

	if len(reqs) == 0 {
		c.String(http.StatusOK, `<div class="text-gray-400 text-sm italic py-4">No requests yet... Waiting for incoming webhooks.</div>`)
		return
	}

	c.HTML(http.StatusOK, "webhook_requests", gin.H{
		"Requests": reqs,
	})
}
