package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowAndBlock(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	limiter := New(3, time.Minute, func() time.Time { return currentTime })
	defer limiter.Close()

	key := "127.0.0.1|admin"
	for range 2 {
		allowed, _ := limiter.Allow(key)
		if !allowed {
			t.Fatal("request should be allowed below failure limit")
		}
		limiter.OnFailure(key)
	}

	allowed, _ := limiter.Allow(key)
	if !allowed {
		t.Fatal("request should still be allowed before reaching failure limit")
	}

	limiter.OnFailure(key)
	allowed, retryAfter := limiter.Allow(key)
	if allowed {
		t.Fatal("request should be blocked after reaching failure limit")
	}
	if retryAfter <= 0 {
		t.Fatalf("expected positive retry-after, got %s", retryAfter)
	}
}

func TestLimiterOnSuccessResetsFailures(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	limiter := New(2, time.Minute, func() time.Time { return currentTime })
	defer limiter.Close()

	key := "127.0.0.1|admin"
	limiter.OnFailure(key)
	limiter.OnFailure(key)

	allowed, _ := limiter.Allow(key)
	if allowed {
		t.Fatal("request should be blocked before success reset")
	}

	limiter.OnSuccess(key)
	allowed, _ = limiter.Allow(key)
	if !allowed {
		t.Fatal("request should be allowed after success reset")
	}
}

func TestLimiterWindowExpires(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	limiter := New(2, time.Minute, func() time.Time { return currentTime })
	defer limiter.Close()

	key := "127.0.0.1|admin"
	limiter.OnFailure(key)
	limiter.OnFailure(key)

	allowed, _ := limiter.Allow(key)
	if allowed {
		t.Fatal("request should be blocked within active window")
	}

	currentTime = currentTime.Add(2 * time.Minute)
	allowed, _ = limiter.Allow(key)
	if !allowed {
		t.Fatal("request should be allowed after window expires")
	}
}
