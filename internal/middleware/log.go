package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err //nolint:wrapcheck // reimplement the interface and do not want to wrap the error
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLog(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rd := &responseData{
				status: 0,
				size:   0,
			}
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   rd,
			}

			h.ServeHTTP(lrw, r)

			duration := time.Since(start)

			if lrw.responseData.status == 0 {
				lrw.responseData.status = http.StatusOK
			}

			logger.InfoContext(r.Context(), "Request",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", lrw.responseData.status),
				slog.Int("size", lrw.responseData.size),
				slog.Duration("duration", duration),
			)
		})
	}
}
