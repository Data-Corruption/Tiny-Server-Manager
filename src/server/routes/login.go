package routes

import (
	"net/http"
	"path/filepath"

	"tsm/src/files"
	"tsm/src/server/middleware"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
)

func RegisterLoginRoutes(r *chi.Mux, usingTLS bool) {
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		// if loggedIn redirect to dashboard
		loggedIn, ok := r.Context().Value(middleware.LoggedInKey).(bool)
		if ok && loggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// else serve login page
		loginPageHtmlPath := filepath.Join("public", "login.html")
		if !files.FileExists(loginPageHtmlPath) {
			blog.Error("login.html not found")
			http.Error(w, "login.html not found", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, loginPageHtmlPath)
	})
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Retrieve the password and check if it is correct
		password := r.FormValue("password")
		if password != files.Config.AdminPassword {
			// ban the ip
			blog.Warn("Invalid password, rate limiting IP: " + r.RemoteAddr)
			if err := middleware.AddRateLimitedIp(r.RemoteAddr); err != nil {
				blog.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/denied", http.StatusFound)
			return
		}

		// Create a new session
		sessionUUID, err := middleware.CreateSession(r.RemoteAddr)
		if err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionUUID.String(),
			HttpOnly: true,
			Secure:   usingTLS,
			SameSite: http.SameSiteStrictMode,
		})

		// Redirect to dashboard
		http.Redirect(w, r, "/", http.StatusFound)
	})
}
