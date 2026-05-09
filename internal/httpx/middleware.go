package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

func SimulatedInfra(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Demo-CDN", "simulated")
			w.Header().Set("X-Demo-API-Gateway", "simulated")
			w.Header().Set("X-Demo-Load-Balancer", "simulated")

			start := time.Now()
			logger.InfoContext(r.Context(), "request entered through simulated CDN -> API Gateway -> Load Balancer -> API service",
				"method", r.Method,
				"path", r.URL.Path,
			)
			next.ServeHTTP(w, r)
			logger.InfoContext(r.Context(), "request_completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
