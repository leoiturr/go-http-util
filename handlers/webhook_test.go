package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func resetWebhookStoreForTest(now time.Time) {
	webhookStore = NewWebhookStore(defaultWebhookTTL, defaultMaxWebhookSessions, defaultMaxWebhookBodyBytes)
	webhookStore.now = func() time.Time { return now }
}

func TestWebhookStoreEvictsExpiredSessions(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := NewWebhookStore(time.Hour, 100, 1<<30)
	s.now = func() time.Time { return now }

	s.register("a")
	s.record("b", WebhookRequest{Body: "x"})

	now = now.Add(2 * time.Hour)
	s.evict()

	if _, ok := s.get("a"); ok {
		t.Fatal("session a should have been evicted by TTL")
	}
	if _, ok := s.get("b"); ok {
		t.Fatal("session b should have been evicted by TTL")
	}
}

func TestWebhookStoreGetRefreshesLastSeen(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := NewWebhookStore(time.Hour, 100, 1<<30)
	s.now = func() time.Time { return now }

	s.register("a")

	now = now.Add(45 * time.Minute)
	if _, ok := s.get("a"); !ok {
		t.Fatal("session should still exist before TTL")
	}

	now = now.Add(45 * time.Minute) // 90m since creation, 45m since refresh
	s.evict()

	if _, ok := s.get("a"); !ok {
		t.Fatal("session should survive because get refreshed LastSeen")
	}
}

func TestWebhookStoreCapsSessionCount(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := NewWebhookStore(time.Hour, 3, 1<<30)
	s.now = func() time.Time { return now }

	for i := 0; i < 5; i++ {
		s.register(fmt.Sprintf("id-%d", i))
		now = now.Add(time.Second)
	}
	s.evict()

	for i := 0; i < 2; i++ {
		if _, ok := s.get(fmt.Sprintf("id-%d", i)); ok {
			t.Fatalf("session id-%d should have been evicted (LRU)", i)
		}
	}
	for i := 2; i < 5; i++ {
		if _, ok := s.get(fmt.Sprintf("id-%d", i)); !ok {
			t.Fatalf("session id-%d should survive", i)
		}
	}
}

func TestWebhookStoreCapsBodyBytes(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := NewWebhookStore(time.Hour, 100, 100)
	s.now = func() time.Time { return now }

	s.record("a", WebhookRequest{Body: strings.Repeat("x", 60)})
	now = now.Add(time.Second)
	s.record("b", WebhookRequest{Body: strings.Repeat("y", 60)})
	s.evict()

	if _, ok := s.get("a"); ok {
		t.Fatal("session a should have been evicted by the body-byte cap")
	}
	if _, ok := s.get("b"); !ok {
		t.Fatal("session b should survive")
	}
}

func TestHTMXGetWebhookRequestsRendersExpiredNotice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetWebhookStoreForTest(time.Now())

	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})
	router.LoadHTMLGlob("../templates/*")
	router.GET("/api/htmx/webhook/:id/requests", HTMXGetWebhookRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/htmx/webhook/unknown-id/requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "expired") {
		t.Fatalf("expected expired notice, got: %s", rec.Body.String())
	}
}

func TestHTMXGetWebhookRequestsRendersEmptyState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetWebhookStoreForTest(time.Now())
	webhookStore.register("abc")

	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})
	router.LoadHTMLGlob("../templates/*")
	router.GET("/api/htmx/webhook/:id/requests", HTMXGetWebhookRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/htmx/webhook/abc/requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "No requests yet") {
		t.Fatalf("expected empty state, got: %s", rec.Body.String())
	}
}

func TestHTMXGetWebhookRequestsRendersRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetWebhookStoreForTest(time.Now())
	webhookStore.register("abc")
	webhookStore.record("abc", WebhookRequest{
		Timestamp:   time.Now(),
		Method:      "POST",
		Path:        "/w/abc",
		Headers:     http.Header{},
		QueryParams: "",
		Body:        "hello",
		StatusCode:  http.StatusOK,
		Delay:       "0s",
	})

	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})
	router.LoadHTMLGlob("../templates/*")
	router.GET("/api/htmx/webhook/:id/requests", HTMXGetWebhookRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/htmx/webhook/abc/requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "POST") || !strings.Contains(body, "hello") {
		t.Fatalf("expected rendered request, got: %s", body)
	}
}
