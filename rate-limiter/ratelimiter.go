// req: concurrency, single process, fixed window
// algo: receive req -> check user's state -> if req is allowed -> increment counter -> return, else -> return
// state: {count, windowStart}
// config: limit, windowSize
// ds: map[userId] *Counter
// API: Allow(userId)

package ratelimiter

import (
	"sync"
	"time"
)

type userID string

type Counter struct {
	mu sync.Mutex
	count int
	windowStart time.Time
}

type RateLimiter struct {
	mu sync.RWMutex
	windowSize time.Duration
	limit int
	users map[userID] *Counter
}

func NewRateLimiter(limit int, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		windowSize: windowSize,
		limit: limit,
		users: make(map[userID] *Counter),
	}
}

func(r *RateLimiter) Allow(userId userID) bool {
	now := time.Now()
	
	r.mu.RLock()
	counter, exists := r.users[userId]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()
		if counter, exists = r.users[userId]; !exists {
			counter = &Counter{
				windowStart: now,
			}
			r.users[userId] = counter
		}
		r.mu.Unlock()
	}
	
	counter.mu.Lock()
	defer counter.mu.Unlock()

	if now.Sub(counter.windowStart) > r.windowSize {
		counter.count = 0
		counter.windowStart = now
	}
	
	if counter.count >= r.limit {
		return false
	}

	counter.count++

	return true
}

func main(usrId userID) {
	rateLimiter := NewRateLimiter(100, time.Minute)

	if !rateLimiter.Allow(usrId) {
		println("rate limit exceeded")
	}
}










