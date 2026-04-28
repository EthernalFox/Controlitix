package authctx

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultCacheTTL        = 15 * time.Minute
	defaultRequestTimeout  = 5 * time.Second
	defaultRefreshInterval = 30 * time.Second
)

type JWKSCache struct {
	url    string
	ttl    time.Duration
	client *http.Client
	logger *slog.Logger

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

type CacheOptions struct {
	URL            string
	TTL            time.Duration
	RequestTimeout time.Duration
	Logger         *slog.Logger
	HTTPClient     *http.Client
}

func NewJWKSCache(options CacheOptions) (*JWKSCache, error) {
	jwksURL := strings.TrimSpace(options.URL)
	if jwksURL == "" {
		return nil, fmt.Errorf("jwks url is required")
	}

	cacheTTL := options.TTL
	if cacheTTL <= 0 {
		cacheTTL = defaultCacheTTL
	}

	requestTimeout := options.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = defaultRequestTimeout
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: requestTimeout,
		}
	}

	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	cache := &JWKSCache{
		url:    jwksURL,
		ttl:    cacheTTL,
		client: httpClient,
		logger: logger,
		keys:   make(map[string]*rsa.PublicKey),
	}

	if err := cache.Refresh(context.Background()); err != nil {
		return nil, fmt.Errorf("initial jwks refresh failed: %w", err)
	}

	return cache, nil
}

func (cache *JWKSCache) GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	kid = strings.TrimSpace(kid)
	if kid == "" {
		return nil, ErrUnknownKey
	}

	cache.mu.RLock()
	key := cache.keys[kid]
	isStale := cache.isStaleLocked()
	cache.mu.RUnlock()

	if key != nil {
		return key, nil
	}

	if !isStale {
		return nil, ErrUnknownKey
	}

	if err := cache.Refresh(ctx); err != nil {
		cache.logger.Warn(
			"jwks refresh failed, using stale cache",
			"method",
			"JWKSCache.GetKey",
			"kid",
			kid,
			"error",
			err,
		)
	}

	cache.mu.RLock()
	key = cache.keys[kid]
	cache.mu.RUnlock()
	if key == nil {
		return nil, ErrUnknownKey
	}

	return key, nil
}

func (cache *JWKSCache) Start(ctx context.Context) (stop func()) {
	if ctx == nil {
		ctx = context.Background()
	}

	refreshEvery := cache.ttl / 2
	if refreshEvery <= 0 {
		refreshEvery = defaultRefreshInterval
	}

	runContext, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(refreshEvery)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()
		defer close(done)

		for {
			select {
			case <-runContext.Done():
				return
			case <-ticker.C:
				if err := cache.Refresh(runContext); err != nil {
					cache.logger.Warn(
						"jwks background refresh failed, stale cache retained",
						"method",
						"JWKSCache.Start",
						"error",
						err,
					)
				}
			}
		}
	}()

	var stopOnce sync.Once
	return func() {
		stopOnce.Do(func() {
			cancel()
			<-done
		})
	}
}

func (cache *JWKSCache) Refresh(ctx context.Context) error {
	keys, err := cache.fetchKeys(ctx)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	cache.keys = keys
	cache.fetchedAt = time.Now()
	cache.mu.Unlock()

	return nil
}

func (cache *JWKSCache) isStaleLocked() bool {
	if cache.fetchedAt.IsZero() {
		return true
	}

	return time.Since(cache.fetchedAt) >= cache.ttl
}

type jwksDocument struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (cache *JWKSCache) fetchKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cache.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build jwks request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := cache.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf(
			"fetch jwks: status %d body %q",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	parsedKeys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, key := range document.Keys {
		parsedKey, err := parseJWKSKey(key)
		if err != nil {
			continue
		}
		parsedKeys[key.Kid] = parsedKey
	}

	if len(parsedKeys) == 0 {
		return nil, fmt.Errorf("jwks has no valid rsa keys")
	}

	return parsedKeys, nil
}

func parseJWKSKey(key jwksKey) (*rsa.PublicKey, error) {
	kid := strings.TrimSpace(key.Kid)
	if kid == "" {
		return nil, fmt.Errorf("jwks key has empty kid")
	}
	if strings.TrimSpace(key.Kty) != "RSA" {
		return nil, fmt.Errorf("jwks key %s has unsupported kty", kid)
	}
	if algorithm := strings.TrimSpace(key.Alg); algorithm != "" && algorithm != "RS256" {
		return nil, fmt.Errorf("jwks key %s has unsupported alg", kid)
	}

	modulusBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(key.N))
	if err != nil {
		return nil, fmt.Errorf("decode jwks key modulus: %w", err)
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(key.E))
	if err != nil {
		return nil, fmt.Errorf("decode jwks key exponent: %w", err)
	}
	if len(modulusBytes) == 0 || len(exponentBytes) == 0 {
		return nil, fmt.Errorf("jwks key has empty modulus or exponent")
	}

	exponentBigInt := new(big.Int).SetBytes(exponentBytes)
	if !exponentBigInt.IsInt64() || exponentBigInt.Sign() <= 0 {
		return nil, fmt.Errorf("jwks key has invalid exponent")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: int(exponentBigInt.Int64()),
	}, nil
}
