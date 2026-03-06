package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const SessionContextKey contextKey = "session"

func (h *Handler) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")
		if id == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid session id")
			return
		}

		sess, ok := h.Engine.Sessions.Get(id)
		if !ok {
			h.Engine.Logger.Error(
				"Session with this ID does not exist",
				"error", "session_does_not_exist",
			)

			writeJSONError(w, http.StatusNotFound, "Session with this ID does not exist")
			return
		}

		// attach session to context
		ctx := context.WithValue(r.Context(), SessionContextKey, sess)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
