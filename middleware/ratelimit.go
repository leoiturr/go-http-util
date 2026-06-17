package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rps     float64
	burst   float64
}

// NewRateLimiter creates a new rate limiter instance.
// rps: requests per second per IP
// burst: maximum burst capacity per IP
func NewRateLimiter(rps, burst float64) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rps:     rps,
		burst:   burst,
	}
	// Start background cleanup of inactive clients to prevent memory leaks
	go rl.cleanupClients()
	return rl
}

// Limit returns a Gin middleware handler.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		cl, exists := rl.clients[ip]
		if !exists {
			cl = &client{
				tokens:   rl.burst,
				lastSeen: time.Now(),
			}
			rl.clients[ip] = cl
		}

		now := time.Now()
		elapsed := now.Sub(cl.lastSeen).Seconds()
		cl.lastSeen = now

		// Add new tokens accrued over elapsed time
		cl.tokens += elapsed * rl.rps
		if cl.tokens > rl.burst {
			cl.tokens = rl.burst
		}

		// If less than 1 token is available, block the request
		if cl.tokens < 1.0 {
			log.Printf("[RateLimiter Warning] IP: %s BLOCKED (Too many requests)", ip)
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		// Consume 1 token
		cl.tokens -= 1.0
		rl.mu.Unlock()
		c.Next()
	}
}

// cleanupClients runs periodically in the background, deleting records of clients
// who haven't made a request in the last 3 minutes.
func (rl *RateLimiter) cleanupClients() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mu.Lock()
		for ip, cl := range rl.clients {
			if time.Since(cl.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
