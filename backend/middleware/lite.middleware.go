package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"aigateway-backend/repositories"

	"github.com/gin-gonic/gin"
)

type LiteMiddleware struct {
	userRepo *repositories.UserRepository
}

func NewLiteMiddleware(userRepo *repositories.UserRepository) *LiteMiddleware {
	return &LiteMiddleware{userRepo: userRepo}
}

func (m *LiteMiddleware) ValidateAccessKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			key = c.GetHeader("X-Access-Key")
		}
		if key == "" || !strings.HasPrefix(key, "uk_") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "access key required",
			})
			return
		}

		user, err := m.userRepo.GetByAccessKey(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid access key",
			})
			return
		}

		SetCurrentUser(c, user)
		c.Next()
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	reqs := rl.requests[ip]
	valid := make([]time.Time, 0, len(reqs))
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, reqs := range rl.requests {
			valid := make([]time.Time, 0)
			for _, t := range reqs {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

var liteRateLimiter = newRateLimiter(10, time.Minute)

func LiteRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !liteRateLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
