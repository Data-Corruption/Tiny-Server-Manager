package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"tsm/src/files"

	"github.com/Data-Corruption/blog"
	"github.com/google/uuid"
)

const LoggedInKey string = "loggedIn"

var (
	// publicRoutes are the routes that are allowed to be accessed without a session
	// If a route ends with "/", it's a prefix match, otherwise it's an exact match
	publicRoutes = []string{
		"/login",
		"/denied",
		"/public/",
	}
	ErrSessionNotFoundOrExpired = errors.New("session not found")
)

func IsPublicRoute(path string) bool {
	for _, route := range publicRoutes {
		// If route ends with "/", it's a prefix match
		if strings.HasSuffix(route, "/") {
			if strings.HasPrefix(path, route) {
				return true
			}
		} else {
			if path == route {
				return true
			}
		}
	}
	return false
}

func IsSessionValid(sessionID uuid.UUID) (files.Session, error) {
	// Check if session exists and is not expired
	var session files.Session
	result := files.DB.Where("id = ? AND exp > ?", sessionID, time.Now()).Limit(1).Find(&session)

	// Check if any session is found
	if result.RowsAffected == 0 {
		return session, ErrSessionNotFoundOrExpired
	}

	// Handle any other errors that might occur
	if result.Error != nil {
		blog.Error(result.Error.Error())
		return session, result.Error
	}

	return session, nil
}

// SessionMiddleware checks for a valid session and redirects to login page if not found.
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the route is public, it doesn't require a session, so go ahead and serve it
		if IsPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			if err == http.ErrNoCookie {
				// No cookie, redirect to login
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Some other error occurred
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Validate session ID
		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			// Invalid session ID, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if session exists and is not expired
		var session files.Session
		if session, err = IsSessionValid(sessionID); err != nil {
			// If session is not found, redirect to login, otherwise return an error
			if err == ErrSessionNotFoundOrExpired {
				http.Redirect(w, r, "/login", http.StatusFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Check if the session IP matches the request IP
		if !IpsMatch(session.IP, r.RemoteAddr) {
			blog.Warn("Session IP mismatch: " + session.IP + " != " + r.RemoteAddr)
			// IP mismatch, add to rate limit and redirect to denied access page
			if err := AddRateLimitedIp(r.RemoteAddr); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/denied", http.StatusFound)
			return
		}

		// Session is valid, add loggedIn context value and continue
		ctx := context.WithValue(r.Context(), LoggedInKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateSession creates a new session for the given IP address and returns the session ID.
func CreateSession(ip string) (uuid.UUID, error) {
	// calculate expiration time using files.Config.SessionDurMins
	exp := time.Now().Add(time.Minute * time.Duration(files.Config.SessionDurMins))

	// Create a new Session instance
	session := files.Session{
		IP:  ip,
		Exp: exp,
	}

	// Add the new record to the database
	result := files.DB.Create(&session)
	if result.Error != nil {
		return uuid.Nil, result.Error
	} else {
		return session.ID, nil
	}
}
