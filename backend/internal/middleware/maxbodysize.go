package middleware

import (
	"net/http"
)

// NewMaxBodySizeHandler returns middleware that rejects request bodies larger than
// limit bytes with 413 Request Entity Too Large.
//
// Two enforcement layers:
//  1. Content-Length early check — if the client announces a body larger than limit,
//     the request is rejected immediately without reading any bytes.
//  2. http.MaxBytesReader wrapping — even without a Content-Length header, any
//     attempt by the handler to read beyond limit bytes returns a MaxBytesError.
//     Handlers (and oapi-codegen's generated decoder) surface this as 413.
func NewMaxBodySizeHandler(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				_, _ = w.Write([]byte(`{"error":{"code":"request_too_large","message":"request body exceeds size limit"}}`))
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
