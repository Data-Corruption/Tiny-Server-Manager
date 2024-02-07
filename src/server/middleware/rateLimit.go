package middleware

import (
	"fmt"
	"net/http"
	"time"

	"tsm/src/files"

	"github.com/Data-Corruption/blog"
)

// LocalIP is the IP of the local machine.
var LocalIP string

// IsIpRateLimited checks if a given IP is rate limited.
func IsIpRateLimited(ip string) (bool, *files.RateLimitedIp, error) {
	var rateLimitedIp files.RateLimitedIp
	result := files.DB.Where("ip = ?", ip).Limit(1).Find(&rateLimitedIp)

	// Check if the IP is found
	if result.RowsAffected == 0 {
		// No records found
		return false, nil, nil
	}

	if result.Error != nil {
		// Some other error occurred
		return false, nil, result.Error
	}

	if rateLimitedIp.Exp.After(time.Now()) {
		// IP is rate limited and not expired
		return true, &rateLimitedIp, nil
	}

	// IP is found but the rate limit has expired
	return false, &rateLimitedIp, nil
}

// AddRateLimitedIp adds a new rate limited IP to the database.
func AddRateLimitedIp(ip string) error {
	// calculate expiration time using files.Config.BanDurationHours
	exp := time.Now().Add(time.Hour * time.Duration(files.Config.BanDurationHours))
	blog.Info(fmt.Sprintf("Rate limiting IP %s until %s", ip, exp))

	// Create a new RateLimitedIp instance
	rateLimitedIp := files.RateLimitedIp{
		Ip:  ip,
		Exp: exp,
	}

	// Add the new record to the database
	result := files.DB.Create(&rateLimitedIp)

	return result.Error
}

// RateLimitMiddleware checks if the request's IP is rate limited and redirects to a denied access page if it is.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ignore on /denied
		if r.URL.Path == "/denied" {
			next.ServeHTTP(w, r)
			return
		}

		ip := r.RemoteAddr // might need to change this to get the real IP

		// if coming from local machine, skip rate limiting
		if IpsMatch(ip, LocalIP) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the IP is rate limited
		isLimited, _, err := IsIpRateLimited(ip)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// If rate limited, redirect to the denied access page
		if isLimited {
			http.Redirect(w, r, "/denied", http.StatusFound)
			return
		}

		// If not rate limited, proceed with the next handler
		next.ServeHTTP(w, r)
	})
}
