package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	buckets         map[string]*bucket
	mu              sync.RWMutex
	rate            int           // tokens per minute
	burst           int           // max tokens
	cleanupInterval time.Duration // cleanup interval for expired buckets
}

// bucket represents a token bucket for a single API key
type bucket struct {
	tokens   int
	lastSeen time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per minute
// burst: maximum number of requests in a burst
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets:         make(map[string]*bucket),
		rate:            rate,
		burst:           burst,
		cleanupInterval: 5 * time.Minute,
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request should be allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	b, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		// Create new bucket
		rl.mu.Lock()
		// Double-check after acquiring write lock
		b, exists = rl.buckets[key]
		if !exists {
			b = &bucket{
				tokens:   rl.burst,
				lastSeen: time.Now(),
			}
			rl.buckets[key] = b
		}
		rl.mu.Unlock()
	}

	// Check and update bucket
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastSeen)

	// Add tokens based on time elapsed
	tokensToAdd := int(elapsed.Minutes() * float64(rl.rate))
	b.tokens += tokensToAdd
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}

	b.lastSeen = now

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanupLoop periodically removes inactive buckets
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanupExpired()
	}
}

// cleanupExpired removes buckets that haven't been accessed recently
func (rl *RateLimiter) cleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)

	for key, b := range rl.buckets {
		b.mu.Lock()
		lastSeen := b.lastSeen
		b.mu.Unlock()

		if lastSeen.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

// RateLimitMiddleware creates a middleware that rate limits requests by API key
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from context (set by auth middleware)
			keyID, ok := r.Context().Value("api_key_id").(int)
			if !ok {
				// No API key in context, skip rate limiting (auth middleware will handle)
				next.ServeHTTP(w, r)
				return
			}

			// Use key ID as the rate limit key
			key := string(rune(keyID))

			if !limiter.Allow(key) {
				ErrorRateLimit(w, 60) // Retry after 60 seconds
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
