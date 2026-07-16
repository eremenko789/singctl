package config

import (
	"fmt"
	"strings"
)

// NormalizeStoredToken validates a bare API token for storage (no Bearer prefix).
func NormalizeStoredToken(input string) (string, error) {
	token := strings.TrimSpace(input)
	if token == "" {
		return "", fmt.Errorf("токен не задан")
	}
	if strings.HasPrefix(token, "Bearer ") {
		return "", fmt.Errorf("передайте токен без префикса Bearer")
	}
	return token, nil
}

// MaskToken returns a display-safe token with only the first and last four characters visible.
func MaskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) < 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// AuthorizationHeader builds an HTTP Authorization value with a Bearer prefix.
func AuthorizationHeader(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	return "Bearer " + token
}
