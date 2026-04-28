package jwks

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerReturnsJWKS(t *testing.T) {
	t.Parallel()

	keyPEM, _ := generatePEMPrivateKey(t, 2048, false)
	keystore, err := NewKeystore(keyPEM, "test-kid")
	if err != nil {
		t.Fatalf("new keystore: %v", err)
	}

	handler := NewHandler(keystore)
	request := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type: %s", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("Cache-Control") != "public, max-age=900" {
		t.Fatalf("unexpected cache control: %s", response.Header().Get("Cache-Control"))
	}
	if response.Body.String() != string(keystore.JWKS()) {
		t.Fatal("unexpected jwks payload")
	}
}
