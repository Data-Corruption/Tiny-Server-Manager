package middleware

import (
	"strings"
)

func stripPort(ip string) string {
	ipSplit := strings.SplitN(ip, ":", 2)
	if len(ipSplit) > 0 {
		return ipSplit[0]
	}
	return ""
}

// Doesn't check if the ports match, only the IP
func IpsMatch(a, b string) bool {
	return stripPort(a) == stripPort(b)
}
