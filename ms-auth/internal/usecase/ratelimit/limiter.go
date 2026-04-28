package ratelimit

import (
	"strings"
	"sync"
	"time"
)

type failureState struct {
	count       int
	firstFailAt time.Time
}

type Limiter struct {
	maxFailures int
	window      time.Duration
	clock       func() time.Time

	states sync.Map
	stopGC chan struct{}
	once   sync.Once
}

func New(maxFailures int, window time.Duration, clock func() time.Time) *Limiter {
	if clock == nil {
		clock = time.Now
	}
	if maxFailures <= 0 {
		maxFailures = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	limiter := &Limiter{
		maxFailures: maxFailures,
		window:      window,
		clock:       clock,
		stopGC:      make(chan struct{}),
	}

	go limiter.gcLoop()
	return limiter
}

func (limiter *Limiter) Allow(key string) (bool, time.Duration) {
	key = normalizeKey(key)
	if key == "" {
		return true, 0
	}

	now := limiter.clock()
	loadedState, ok := limiter.states.Load(key)
	if !ok {
		return true, 0
	}

	state := loadedState.(failureState)
	if now.Sub(state.firstFailAt) > limiter.window {
		limiter.states.Delete(key)
		return true, 0
	}

	if state.count >= limiter.maxFailures {
		retryAfter := limiter.window - now.Sub(state.firstFailAt)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	return true, 0
}

func (limiter *Limiter) OnFailure(key string) {
	key = normalizeKey(key)
	if key == "" {
		return
	}

	now := limiter.clock()
	loadedState, ok := limiter.states.Load(key)
	if !ok {
		limiter.states.Store(key, failureState{
			count:       1,
			firstFailAt: now,
		})
		return
	}

	state := loadedState.(failureState)
	if now.Sub(state.firstFailAt) > limiter.window {
		limiter.states.Store(key, failureState{
			count:       1,
			firstFailAt: now,
		})
		return
	}

	state.count++
	limiter.states.Store(key, state)
}

func (limiter *Limiter) OnSuccess(key string) {
	key = normalizeKey(key)
	if key == "" {
		return
	}

	limiter.states.Delete(key)
}

func (limiter *Limiter) Close() {
	limiter.once.Do(func() {
		close(limiter.stopGC)
	})
}

func (limiter *Limiter) gcLoop() {
	ticker := time.NewTicker(limiter.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := limiter.clock()
			limiter.states.Range(func(key, value any) bool {
				state := value.(failureState)
				if now.Sub(state.firstFailAt) > limiter.window {
					limiter.states.Delete(key)
				}
				return true
			})
		case <-limiter.stopGC:
			return
		}
	}
}

func normalizeKey(key string) string {
	return strings.TrimSpace(key)
}
