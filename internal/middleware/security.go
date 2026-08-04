package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds essential security headers to all responses
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Content Security Policy - restrict resource loading
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com https://cdn.tailwindcss.com https://cdnjs.cloudflare.com; " +
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com https://cdn.tailwindcss.com https://cdnjs.cloudflare.com; " +
				"font-src 'self' https://fonts.gstatic.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; " +
				"img-src 'self' data: https:; " +
				"connect-src 'self' https://api.openweathermap.org https://cdn.jsdelivr.net https://unpkg.com; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self'"
			c.Response().Header().Set("Content-Security-Policy", csp)

			// Prevent clickjacking
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// XSS Protection (legacy but still useful)
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer Policy - control referrer information
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy - disable unnecessary browser features
			c.Response().Header().Set("Permissions-Policy", 
				"geolocation=(), microphone=(), camera=(self), payment=()")

			// HSTS - force HTTPS (only if running on HTTPS)
			if c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", 
					"max-age=31536000; includeSubDomains; preload")
			}

			return next(c)
		}
	}
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter creates a new rate limiter
// limit: number of requests allowed per window
// window: time window duration (e.g., 1 minute, 1 hour)
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// Cleanup old visitors every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

// Middleware returns Echo middleware function
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			
			// Check if IP is allowed
			if !rl.allow(ip) {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Rate limit exceeded",
					"message": fmt.Sprintf("Maksimal %d request per %s. Coba lagi nanti.", rl.limit, rl.window),
				})
			}

			return next(c)
		}
	}
}

// allow checks if a request from the given IP is allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		// First request from this IP
		rl.visitors[ip] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// Check if window has expired
	if now.Sub(v.lastSeen) > rl.window {
		// Reset counter
		v.lastSeen = now
		v.count = 1
		return true
	}

	// Window still active, check limit
	if v.count >= rl.limit {
		return false
	}

	// Increment counter
	v.count++
	v.lastSeen = now
	return true
}

// cleanup removes old visitors to prevent memory leak
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

// APIRateLimiter creates a rate limiter for API endpoints
// Default: 100 requests per minute per IP
func APIRateLimiter() echo.MiddlewareFunc {
	limiter := NewRateLimiter(100, time.Minute)
	return limiter.Middleware()
}

// AuthRateLimiter creates a rate limiter for authentication endpoints
// More restrictive: 5 requests per minute per IP to prevent brute force
func AuthRateLimiter() echo.MiddlewareFunc {
	limiter := NewRateLimiter(5, time.Minute)
	return limiter.Middleware()
}

// AdminRateLimiter creates a rate limiter for admin write operations
// Moderate: 30 requests per minute per IP
func AdminRateLimiter() echo.MiddlewareFunc {
	limiter := NewRateLimiter(30, time.Minute)
	return limiter.Middleware()
}
