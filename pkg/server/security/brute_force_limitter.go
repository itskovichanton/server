package security

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type attempt struct {
	count          int8 // Number of failed attempts
	lastUpdateTime time.Time
	scheduled      bool // Flag to check whether attempt is scheduled to be destroyed
}

func newAttempt() *attempt {
	return &attempt{
		count:          1,
		lastUpdateTime: time.Now(),
		scheduled:      false,
	}
}

func (p *attempt) increment() {
	if p.count < 3 {
		p.count++
	}
	p.lastUpdateTime = time.Now()
}

func (p *attempt) isMature(interval time.Duration) bool {
	return time.Now().Sub(p.lastUpdateTime) >= interval
}

func (p *attempt) canCheck(interval time.Duration, limit int8) bool {
	return p.count < limit || (p.isMature(interval) && p.count == limit)
}

type AttemptLimiter struct {
	attempts     map[string]*attempt
	lock         sync.Mutex
	limit        int8
	errorMessage string
	interval     time.Duration
}

func NewAttemptLimiter(limit int8, interval time.Duration, errorMessage string) *AttemptLimiter {
	p := &AttemptLimiter{
		attempts:     map[string]*attempt{},
		limit:        limit,
		errorMessage: errorMessage,
		interval:     interval,
	}
	c := cron.New()
	if _, err := c.AddFunc("@every 3h", p.cleanupMature); err != nil {
		panic(err)
	}
	c.Start()
	return p
}

func (p *AttemptLimiter) Reset(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.attempts, key)
}

func (p *AttemptLimiter) canCheck(key string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	attempt, ok := (*p).attempts[key]
	if !ok {
		return true
	}
	return attempt.canCheck(p.interval, p.limit)
}

func (p *AttemptLimiter) Increment(key string) {
	p.lock.Lock()
	attempt, ok := (*p).attempts[key]
	if !ok {
		(*p).attempts[key] = newAttempt()
	} else {
		attempt.increment()
	}
	p.lock.Unlock()
	if !p.canCheck(key) {
		p.destroyAttemptAfterOrSkip(key)
	}
}

func (p *AttemptLimiter) CheckWithResponse(key string) error {
	if p.canCheck(key) {
		return nil
	}
	return errors.Errorf(p.errorMessage, p.interval.Minutes())
}

// cleanupMature is called periodically by cron job
func (p *AttemptLimiter) cleanupMature() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for key, value := range p.attempts {
		if value.isMature(p.interval) {
			delete(p.attempts, key)
		}
	}
}

func (p *AttemptLimiter) destroyAttemptAfterOrSkip(key string) {
	attempt, ok := p.attempts[key]
	if ok && !attempt.scheduled {
		time.AfterFunc(p.interval, func() {
			p.lock.Lock()
			delete(p.attempts, key)
			p.lock.Unlock()
		})
	}
}
