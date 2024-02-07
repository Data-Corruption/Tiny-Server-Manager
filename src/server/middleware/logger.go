package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5/middleware"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use a custom ResponseWriter to capture the status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		blog.Info(fmt.Sprintf("Started %s %s", r.Method, r.URL.Path))
		start := time.Now()
		next.ServeHTTP(ww, r)
		blog.Info(fmt.Sprintf("Completed %s in %v with status %d", r.URL.Path, time.Since(start), ww.Status()))
	})
}
