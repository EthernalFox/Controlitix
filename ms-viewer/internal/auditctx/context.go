package auditctx

import "context"

type contextKey string

const (
	requestIDContextKey contextKey = "request_id"
	ipContextKey        contextKey = "client_ip"
	userAgentContextKey contextKey = "user_agent"
)

type RequestMetadata struct {
	RequestID string
	IP        string
	UserAgent string
}

func WithRequestMetadata(ctx context.Context, metadata RequestMetadata) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx = context.WithValue(ctx, requestIDContextKey, metadata.RequestID)
	ctx = context.WithValue(ctx, ipContextKey, metadata.IP)
	ctx = context.WithValue(ctx, userAgentContextKey, metadata.UserAgent)
	return ctx
}

func RequestMetadataFromContext(ctx context.Context) RequestMetadata {
	if ctx == nil {
		return RequestMetadata{}
	}

	requestID, _ := ctx.Value(requestIDContextKey).(string)
	ip, _ := ctx.Value(ipContextKey).(string)
	userAgent, _ := ctx.Value(userAgentContextKey).(string)

	return RequestMetadata{
		RequestID: requestID,
		IP:        ip,
		UserAgent: userAgent,
	}
}
