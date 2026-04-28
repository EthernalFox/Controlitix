package jwks

import "net/http"

func NewHandler(keystore *Keystore) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("Cache-Control", "public, max-age=900")
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write(keystore.JWKS())
	})
}
