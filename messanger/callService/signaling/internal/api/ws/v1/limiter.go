package v1

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	last     time.Time
	violations int
}

func newRateLimiter(perSec int) *rateLimiter {
	if perSec <= 0 {
		perSec = 50
	}
	return &rateLimiter{
		tokens: float64(perSec),
		max:    float64(perSec),
		rate:   float64(perSec),
		last:   time.Now(),
	}
}

func (l *rateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now
	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}
	if l.tokens < 1 {
		l.violations++
		return false
	}
	l.tokens--
	return true
}

func (l *rateLimiter) Violations() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.violations
}
