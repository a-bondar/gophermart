package middleware

import (
	"log/slog"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func WithLog(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			logger.InfoContext(r.Context(),
				"Incoming request",
				slog.String("method", r.Method),
				slog.String("status", http.StatusText(ww.status)),
				slog.String("uri", r.RequestURI),
			)

			next.ServeHTTP(ww, r)
		})
	}
}
