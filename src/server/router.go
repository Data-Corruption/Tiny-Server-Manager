package server

import (
	"net/http"
	"path/filepath"

	"tsm/src/files"
	"tsm/src/server/middleware"
	"tsm/src/server/routes"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
)

// DeniedAccessHandler serves a simple HTML page saying "Access Denied" in the same style as the login page.
func DeniedAccessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "text/html")
	deniedPageHtmlPath := filepath.Join("public", "denied.html")
	if !files.FileExists(deniedPageHtmlPath) {
		blog.Error("login.html not found")
		http.Error(w, "login.html not found", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, deniedPageHtmlPath)
}

// NewRouter creates and returns a new Chi router.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.RateLimitMiddleware)
	r.Use(middleware.SessionMiddleware)

	// Define routes
	routes.RegisterLoginRoutes(r, Instance.UsingTLS)
	routes.RegisterDashboardRoutes(r)
	r.Get("/denied", DeniedAccessHandler)

	// Serve static files
	r.Get("/public/logo.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml") // Set the correct content type for SVG
		http.ServeFile(w, r, filepath.Join("public", "logo.svg"))
	})
	r.Get("/public/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css") // Set the correct content type for CSS
		http.ServeFile(w, r, filepath.Join("public", "style.css"))
	})

	// Shutdown route in case remote shutdown is needed (might add a separate password for this in the future)
	r.Get("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		Instance.ShutdownRoute <- true
		w.Write([]byte("Goodnight!"))
	})

	return r
}
