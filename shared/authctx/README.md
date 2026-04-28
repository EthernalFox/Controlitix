# shared/authctx

`shared/authctx` is a shared Go module for access-token validation in Controlitix
services.

## Features

- JWKS fetch over HTTP with in-memory cache.
- TTL-based cache invalidation and background refresh.
- Stale fallback when refresh fails.
- JWT validation for RS256 only.
- Validation of `iss`, `aud`, `exp`, `nbf`.
- Principal extraction into `context.Context`.
- HTTP middlewares: `RequireAuth`, `RequireRole`, `RequireScope`,
  `RequireAudience`.
- RFC7807 responses via `application/problem+json`.

## Usage

```go
cache, err := authctx.NewJWKSCache(authctx.CacheOptions{
    URL:    "http://ms-auth:8083/.well-known/jwks.json",
    Logger: logger,
})
if err != nil {
    return err
}
stop := cache.Start(ctx)
defer stop()

validator := authctx.NewValidator(authctx.ValidatorOptions{
    JWKSCache:          cache,
    ExpectedIssuer:     "controlitix-auth",
    ExpectedAudiences:  []string{"controlitix-api"},
    Logger:             logger,
})

httpHandler := authctx.RequireAuth(validator)(handler)
httpHandler = authctx.RequireRole("engineer", "admin")(httpHandler)
```

## Dev Bypass

For local development only, set `ValidatorOptions.Disabled=true`.

- Token validation is skipped.
- `Parse` always returns a synthetic principal
  `{Subject:"dev:anonymous", Roles:["admin"]}`.
- `RequireAuth` logs a warning that authentication is disabled.

This mode must not be used in production.
