package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/droskat/BH-Backend/models"
)

type rateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64
	lastTime time.Time
}

func newRateLimiter(maxTokens, refillRate float64) *rateLimiter {
	return &rateLimiter{
		tokens:   maxTokens,
		max:      maxTokens,
		rate:     refillRate,
		lastTime: time.Now(),
	}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.tokens += elapsed * rl.rate
	if rl.tokens > rl.max {
		rl.tokens = rl.max
	}
	rl.lastTime = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

var globalLimiter = newRateLimiter(10000, 10000)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLimiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
