package service

import (
	"sync"
	"time"
)

type RateLimiter struct {
	tokens     int
	capacity   int
	refillRate int
	mutex      sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		tokens:     5, 
		capacity:   5, 
		refillRate: 1, // refill 1 token per second
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			rl.mutex.Lock()
			if rl.tokens < rl.capacity {
				rl.tokens++
			}
			rl.mutex.Unlock()
		}
	}()

	return rl
}
func (r *RateLimiter) Allow() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}
