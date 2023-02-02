package csrf

import (
	"context"
	"net/http"

	"github.com/derinil/links/links/web/responder"
)

type CtxKey string

const (
	CookieName  string = "csrf"
	TokenKey    CtxKey = "csrf_token"
	TokenKeyStr string = string(TokenKey)
)

func ValidateCSRF(csrfHandler Handler, responderHandler responder.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.FormValue(TokenKeyStr)
			r.Form.Del(TokenKeyStr)

			c, err := r.Cookie(CookieName)
			if err != nil {
				responderHandler.Respond(w, r, &responder.ResponseCmd{
					Path:     r.URL.Path,
					ErrorMsg: "CSRF token is empty or invalid",
				})
				return
			}

			cookie := c.Value

			if ok, err := csrfHandler.ValidateCookieToken(cookie, token); err != nil || !ok {
				responderHandler.Respond(w, r, &responder.ResponseCmd{
					Path:     r.URL.Path,
					ErrorMsg: "CSRF token is empty or invalid",
				})
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func InjectCSRF(csrfHandler Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cookie string

			if c, err := r.Cookie(CookieName); err != nil || c == nil {
				cookie, err = csrfHandler.CreateCookie()
				if err != nil {
					h.ServeHTTP(w, r)
					return
				}

				http.SetCookie(w, Cookie(cookie))
			} else {
				cookie = c.Value
			}

			token, err := csrfHandler.CreateToken(cookie)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), TokenKey, token))

			h.ServeHTTP(w, r)
		})
	}
}

func Cookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:  CookieName,
		Value: token,
	}
}
