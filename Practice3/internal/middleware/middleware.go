package middleware

import (
	"log"
	"net/http"
	"time"
)

const validAPIKey = "ali"

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		log.Printf(
			"[%s] method=%s endpoint=%s status=%d duration=%s",
			start.Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
		)
	})
}

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-KEY")
		if key == "" {
			http.Error(w, `{"error":"missing X-API-KEY header"}`, http.StatusUnauthorized)
			return
		}
		if key != validAPIKey {
			http.Error(w, `{"error":"invalid X-API-KEY"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
