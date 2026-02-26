package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// RATE LIMITER — Token bucket per IP
//
// Mỗi IP có 1 bucket. Bucket bắt đầu với `burst` tokens.
// Mỗi giây thêm `rate` tokens. Mỗi request tiêu 1 token.
// Hết token → 429 Too Many Requests.
//
// Giải pháp đơn giản, production nên dùng Redis-based
// (VD: github.com/ulule/limiter) để hoạt động trên multi-instance.
// ============================================================

type ipBucket struct {
	tokens     float64
	lastRefill time.Time
}

type rateLimiterStore struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	rate    float64 // tokens/second
	burst   float64 // max tokens
}

func newRateLimiterStore(rate, burst float64) *rateLimiterStore {
	store := &rateLimiterStore{
		buckets: make(map[string]*ipBucket),
		rate:    rate,
		burst:   burst,
	}

	// Cleanup expired buckets mỗi 5 phút
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.cleanup()
		}
	}()

	return store
}

func (s *rateLimiterStore) allow(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[ip]
	now := time.Now()

	if !exists {
		s.buckets[ip] = &ipBucket{
			tokens:     s.burst - 1, // tiêu 1 token ngay
			lastRefill: now,
		}
		return true
	}

	// Refill tokens dựa trên thời gian trôi qua
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * s.rate
	if bucket.tokens > s.burst {
		bucket.tokens = s.burst
	}
	bucket.lastRefill = now

	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--
	return true
}

func (s *rateLimiterStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for ip, bucket := range s.buckets {
		if bucket.lastRefill.Before(cutoff) {
			delete(s.buckets, ip)
		}
	}
}

// RateLimiter — middleware giới hạn request per IP
// burst: số request tối đa cùng lúc (VD: 100)
// rate: số request/giây bền vững (VD: 10)
func RateLimiter(burst, rate float64) gin.HandlerFunc {
	store := newRateLimiterStore(rate, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !store.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"code":    http.StatusTooManyRequests,
				"message": "too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}
