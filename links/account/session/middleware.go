package session

import (
	"context"
	"net/http"

	"github.com/derinil/links/links/web/responder"
)

const (
	SessionTokenKey  CtxKey = "session_token"
	SessionObjectKey CtxKey = "session_object"
)

func ParseSession(sessionHandler Handler) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CookieName)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}

			if c.Value == "" {
				h.ServeHTTP(w, r)
				return
			}

			t := c.Value

			// NOTE: If this ends up taking too much time
			// we can convert the token into an encrypted
			// packet that holds the token itself and the handle
			// sort of a mix of jwt and this
			se, err := sessionHandler.Get(r.Context(), t)
			if err != nil || se == nil {
				h.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), SessionTokenKey, t))
			r = r.WithContext(context.WithValue(r.Context(), SessionObjectKey, se))

			h.ServeHTTP(w, r)
		})
	}
}

func ForceSession(responderHandler responder.Handler) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(SessionObjectKey) == nil {
				responderHandler.Respond(w, r, &responder.ResponseCmd{
					Path:     "/login",
					ErrorMsg: "You are not authenticated!",
				})
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func ForceNoSession(responderHandler responder.Handler) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(SessionObjectKey) != nil {
				responderHandler.Respond(w, r, &responder.ResponseCmd{
					Path:     "/",
					ErrorMsg: "You are already authenticated!",
				})
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
