package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*userLimit
	mu       sync.RWMutex
}

type userLimit struct {
	count      int
	resetTime  time.Time
	mu         sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userLimit),
	}
	// Cleanup goroutine
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Limit(requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			if userID == "" {
				// Use IP if no user ID
				userID = r.RemoteAddr
			}

			if !rl.allow(userID, requestsPerSecond) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"RATE_LIMITED","message":"Too many requests","retry_after":1}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(userID string, limit int) bool {
	rl.mu.Lock()
	ul, exists := rl.requests[userID]
	if !exists {
		ul = &userLimit{
			count:     0,
			resetTime: time.Now().Add(time.Second),
		}
		rl.requests[userID] = ul
	}
	rl.mu.Unlock()

	ul.mu.Lock()
	defer ul.mu.Unlock()

	now := time.Now()
	if now.After(ul.resetTime) {
		ul.count = 0
		ul.resetTime = now.Add(time.Second)
	}

	if ul.count >= limit {
		return false
	}

	ul.count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for userID, ul := range rl.requests {
			ul.mu.Lock()
			if now.After(ul.resetTime.Add(1 * time.Minute)) {
				delete(rl.requests, userID)
			}
			ul.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
