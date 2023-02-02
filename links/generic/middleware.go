package generic

import (
	"context"
	"net/http"
	"time"
)

type CtxKey string

const RequestBeginTimeKey CtxKey = "request_begin_time"

func RequestBeginTime(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), RequestBeginTimeKey, time.Now().UTC()))
		h.ServeHTTP(w, r)
	})
}
