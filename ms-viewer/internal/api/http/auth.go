package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/EthernalFox/Controlitix/shared/authctx"
)

func BuildAuthMiddleware(
	ctx context.Context,
	jwksURL string,
	issuer string,
	audience string,
	logger *slog.Logger,
) (func(http.Handler) http.Handler, *authctx.Validator, func(), error) {
	if logger == nil {
		logger = slog.Default()
	}

	jwksCache, cacheError := authctx.NewJWKSCache(authctx.CacheOptions{
		URL:    jwksURL,
		Logger: logger,
	})
	if cacheError != nil {
		return nil, nil, nil, fmt.Errorf("initialize jwks cache: %w", cacheError)
	}

	stopJWKSRefresh := jwksCache.Start(ctx)
	validator := authctx.NewValidator(authctx.ValidatorOptions{
		JWKSCache:         jwksCache,
		ExpectedIssuer:    issuer,
		ExpectedAudiences: []string{audience},
		Logger:            logger,
	})

	return authctx.RequireAuth(validator), validator, stopJWKSRefresh, nil
}
