package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: msg,
	})
}

func AuthMiddleware(allowedScopes ...Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "Missing Authorization header")
				return
			}

			// Expect: Bearer <key>
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSONError(w, http.StatusUnauthorized, "Invalid Authorization format")
				return
			}

			apiKey := parts[1]

			scope, valid, err := ValidateAPIKeyWithScope(apiKey)
			if err != nil || !valid {
				writeJSONError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// check scope
			allowed := false
			for _, s := range allowedScopes {
				if s == scope {
					allowed = true
					break
				}
			}

			if !allowed {
				writeJSONError(w, http.StatusForbidden, "Forbidden: insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func WSAuthMiddleware(allowedScopes ...Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			apiKey := r.URL.Query().Get("api_token")
			if apiKey == "" {
				writeJSONError(w, http.StatusUnauthorized, "Missing Authorization header")
				return
			}

			scope, valid, err := ValidateAPIKeyWithScope(apiKey)
			if err != nil || !valid {
				writeJSONError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// check scope
			allowed := false
			for _, s := range allowedScopes {
				if s == scope {
					allowed = true
					break
				}
			}

			if !allowed {
				writeJSONError(w, http.StatusForbidden, "Forbidden: insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
