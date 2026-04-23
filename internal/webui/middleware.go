package webui

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthMiddleware enforces basic API key authentication.
type AuthMiddleware struct {
	apiKey string
}

// NewAuthMiddleware creates an auth middleware. If apiKey is empty, auth is disabled.
func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{apiKey: apiKey}
}

func (a *AuthMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	if a.apiKey == "" {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(a.apiKey)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="Aigo"`)
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// rateLimitEntry tracks request timestamps for a client.
type rateLimitEntry struct {
	requests []time.Time
}

// RateLimitMiddleware provides simple in-memory rate limiting.
type RateLimitMiddleware struct {
	maxRequests int
	window      time.Duration
	clients     map[string]*rateLimitEntry
	mu          sync.Mutex
}

// NewRateLimitMiddleware creates a rate limiter. If maxRequests is 0, rate limiting is disabled.
func NewRateLimitMiddleware(maxRequests int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*rateLimitEntry),
	}
}

func (rl *RateLimitMiddleware) clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-Ip")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	// Strip port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (rl *RateLimitMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	if rl.maxRequests <= 0 {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ip := rl.clientIP(r)
		now := time.Now()

		rl.mu.Lock()
		entry, ok := rl.clients[ip]
		if !ok {
			entry = &rateLimitEntry{}
			rl.clients[ip] = entry
		}
		// Remove old requests outside the window
		cutoff := now.Add(-rl.window)
		var valid []time.Time
		for _, t := range entry.requests {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		entry.requests = valid
		allowed := len(entry.requests) < rl.maxRequests
		if allowed {
			entry.requests = append(entry.requests, now)
		}
		rl.mu.Unlock()

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.window.Seconds()))
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// CORSMiddleware adds CORS headers.
type CORSMiddleware struct {
	allowedOrigins []string
}

// NewCORSMiddleware creates a CORS middleware. If allowedOrigins is empty, allows all origins.
func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{allowedOrigins: allowedOrigins}
}

func (c *CORSMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		// Check allowed origins
		if len(c.allowedOrigins) > 0 {
			allowed := false
			for _, o := range c.allowedOrigins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}
			if !allowed {
				origin = c.allowedOrigins[0]
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
