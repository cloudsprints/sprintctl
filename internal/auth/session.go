package auth

import (
	"fmt"
	"strings"
)

// AccessToken returns the stored session token for API requests.
//
// Tokens are opaque BetterAuth session tokens: 30-day expiry, slid forward
// server-side on every authenticated request. There is no refresh token and
// nothing to renew client-side, so the server stays the sole authority on
// whether a token is still good.
func AccessToken() (string, error) {
	tokens, err := GetTokens()
	if err != nil {
		return "", fmt.Errorf("authentication required: please run 'sprintctl login' first")
	}

	if isLegacyJWT(tokens.AccessToken) {
		return "", fmt.Errorf("your saved session is from an older sprintctl release and no longer works: please run 'sprintctl login' again")
	}

	return tokens.AccessToken, nil
}

// isLegacyJWT reports whether a stored token is a JWT issued by the previous
// identity provider (three dot-separated segments). BetterAuth session tokens
// never contain dots.
func isLegacyJWT(token string) bool {
	return len(strings.Split(token, ".")) == 3
}
