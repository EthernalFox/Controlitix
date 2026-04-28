package authctx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewJWKSCacheInitialFetchSuccess(t *testing.T) {
	t.Parallel()

	privateKey := generateTestRSAKey(t)
	serverState := &jwksServerState{
		statusCode: http.StatusOK,
		body:       buildJWKSResponse(t, "kid-1", &privateKey.PublicKey),
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	cache, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		TTL:    time.Minute,
		Logger: testLogger(),
	})
	if err != nil {
		t.Fatalf("create jwks cache: %v", err)
	}

	if _, err := cache.GetKey(context.Background(), "kid-1"); err != nil {
		t.Fatalf("get key from cache: %v", err)
	}
	if serverState.Calls() != 1 {
		t.Fatalf("expected one jwks call, got %d", serverState.Calls())
	}
}

func TestGetKeyUnknownKidFreshCacheNoRefresh(t *testing.T) {
	t.Parallel()

	privateKey := generateTestRSAKey(t)
	serverState := &jwksServerState{
		statusCode: http.StatusOK,
		body:       buildJWKSResponse(t, "kid-1", &privateKey.PublicKey),
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	cache, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		TTL:    time.Minute,
		Logger: testLogger(),
	})
	if err != nil {
		t.Fatalf("create jwks cache: %v", err)
	}

	_, err = cache.GetKey(context.Background(), "kid-2")
	if err == nil || err != ErrUnknownKey {
		t.Fatalf("expected ErrUnknownKey, got %v", err)
	}
	if serverState.Calls() != 1 {
		t.Fatalf("expected no extra jwks call, got %d", serverState.Calls())
	}
}

func TestGetKeyUnknownKidStaleCacheTriggersRefresh(t *testing.T) {
	t.Parallel()

	privateKeyOne := generateTestRSAKey(t)
	privateKeyTwo := generateTestRSAKey(t)
	serverState := &jwksServerState{
		statusCode: http.StatusOK,
		body:       buildJWKSResponse(t, "kid-1", &privateKeyOne.PublicKey),
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	cache, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		TTL:    20 * time.Millisecond,
		Logger: testLogger(),
	})
	if err != nil {
		t.Fatalf("create jwks cache: %v", err)
	}

	time.Sleep(40 * time.Millisecond)
	serverState.Set(http.StatusOK, buildJWKSResponse(t, "kid-2", &privateKeyTwo.PublicKey))

	if _, err := cache.GetKey(context.Background(), "kid-2"); err != nil {
		t.Fatalf("expected refreshed key, got %v", err)
	}
	if serverState.Calls() < 2 {
		t.Fatalf("expected refresh call, got %d calls", serverState.Calls())
	}
}

func TestRefreshFailureKeepsStaleKeys(t *testing.T) {
	t.Parallel()

	privateKey := generateTestRSAKey(t)
	serverState := &jwksServerState{
		statusCode: http.StatusOK,
		body:       buildJWKSResponse(t, "kid-1", &privateKey.PublicKey),
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	cache, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		TTL:    time.Minute,
		Logger: testLogger(),
	})
	if err != nil {
		t.Fatalf("create jwks cache: %v", err)
	}

	serverState.Set(http.StatusInternalServerError, `{"error":"boom"}`)
	if err := cache.Refresh(context.Background()); err == nil {
		t.Fatal("expected refresh error")
	}

	if _, err := cache.GetKey(context.Background(), "kid-1"); err != nil {
		t.Fatalf("expected stale key to remain available, got %v", err)
	}
}

func TestNewJWKSCacheInitialFetchFailure(t *testing.T) {
	t.Parallel()

	serverState := &jwksServerState{
		statusCode: http.StatusInternalServerError,
		body:       `{"error":"boom"}`,
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	if _, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		Logger: testLogger(),
	}); err == nil {
		t.Fatal("expected initial refresh failure")
	}
}

func TestStartStop(t *testing.T) {
	t.Parallel()

	privateKey := generateTestRSAKey(t)
	serverState := &jwksServerState{
		statusCode: http.StatusOK,
		body:       buildJWKSResponse(t, "kid-1", &privateKey.PublicKey),
	}
	server := httptest.NewServer(serverState)
	defer server.Close()

	cache, err := NewJWKSCache(CacheOptions{
		URL:    server.URL,
		TTL:    20 * time.Millisecond,
		Logger: testLogger(),
	})
	if err != nil {
		t.Fatalf("create jwks cache: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := cache.Start(ctx)
	time.Sleep(60 * time.Millisecond)
	cancel()

	stopDone := make(chan struct{})
	go func() {
		stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("stop did not finish")
	}
}

type jwksServerState struct {
	mu         sync.RWMutex
	statusCode int
	body       string
	calls      int
}

func (serverState *jwksServerState) Set(statusCode int, body string) {
	serverState.mu.Lock()
	serverState.statusCode = statusCode
	serverState.body = body
	serverState.mu.Unlock()
}

func (serverState *jwksServerState) Calls() int {
	serverState.mu.RLock()
	defer serverState.mu.RUnlock()
	return serverState.calls
}

func (serverState *jwksServerState) ServeHTTP(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	serverState.mu.Lock()
	serverState.calls++
	statusCode := serverState.statusCode
	body := serverState.body
	serverState.mu.Unlock()

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	_, _ = responseWriter.Write([]byte(body))
}

func buildJWKSResponse(t *testing.T, kid string, publicKey *rsa.PublicKey) string {
	t.Helper()

	modulus := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	exponent := base64.RawURLEncoding.EncodeToString(
		big.NewInt(int64(publicKey.E)).Bytes(),
	)

	return fmt.Sprintf(
		`{"keys":[{"kty":"RSA","kid":"%s","alg":"RS256","n":"%s","e":"%s"}]}`,
		kid,
		modulus,
		exponent,
	)
}

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	return privateKey
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
