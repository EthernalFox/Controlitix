package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func TestClientSendMessageClassifies429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.Header().Set("Retry-After", "30")
		responseWriter.WriteHeader(http.StatusTooManyRequests)
		_, _ = responseWriter.Write([]byte(`{"ok":false,"description":"Too Many Requests"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", time.Second, nil)
	err := client.SendMessage(context.Background(), 1001, "text", ParseModeMarkdownV2)
	if err == nil {
		t.Fatal("expected error")
	}

	rateLimitError, ok := domain.IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected rate limit error, got %T", err)
	}
	if rateLimitError.RetryAfter != 30*time.Second {
		t.Fatalf("unexpected retry-after: %s", rateLimitError.RetryAfter)
	}
}

func TestClientSendMessageClassifies4xxAsPermanent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusBadRequest)
		_, _ = responseWriter.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", time.Second, nil)
	err := client.SendMessage(context.Background(), 1001, "text", ParseModeMarkdownV2)
	if err == nil {
		t.Fatal("expected error")
	}
	if !domain.IsPermanentSendError(err) {
		t.Fatalf("expected permanent error, got %T", err)
	}
}

func TestClientSendMessageClassifies5xxAsTransient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(`boom`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", time.Second, nil)
	err := client.SendMessage(context.Background(), 1001, "text", ParseModeMarkdownV2)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := domain.IsRateLimitError(err); ok {
		t.Fatalf("did not expect rate limit error")
	}
	if domain.IsPermanentSendError(err) {
		t.Fatalf("did not expect permanent error")
	}
}
