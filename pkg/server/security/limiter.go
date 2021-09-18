package security

import (
	"sync"

	"github.com/robfig/cron/v3"
	"golang.org/x/time/rate"
)

type Limiter interface {
	Allow(string) bool
}

func NewLimiter(eps float64, b int) Limiter {
	return newLimiter(rate.Limit(eps), b)
}

type limiter struct {
	mu    sync.Mutex
	limit rate.Limit
	burst int // Maximum burst size
	store map[string]*rate.Limiter
}

func newLimiter(l rate.Limit, b int) *limiter {
	var lm = &limiter{
		limit: l,
		burst: b,
		store: make(map[string]*rate.Limiter),
	}
	var c = cron.New()
	if _, err := c.AddFunc("@every 3h", lm.clean); err != nil {
		panic(err)
	}
	return lm
}

func (l *limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	var lm, ok = l.store[ip]
	if !ok {
		lm = rate.NewLimiter(l.limit, l.burst)
		l.store[ip] = lm
	}
	return lm.Allow()
}

func (l *limiter) clean() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip := range l.store {
		delete(l.store, ip)
	}
}
