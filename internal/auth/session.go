package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// expiryLeeway refreshes tokens slightly before they expire so a request
// doesn't fail mid-flight on a token that lapses in transit
const expiryLeeway = 60 * time.Second

// AccessToken returns a valid access token for API requests, refreshing the
// session first if the stored token is expired or about to expire
func AccessToken() (string, error) {
	tokens, err := GetTokens()
	if err != nil {
		return "", fmt.Errorf("authentication required: please run 'sprintctl login' first")
	}

	// Opaque BetterAuth session tokens carry no client-readable expiry and
	// are renewed server-side on use (sliding expiry), so they are returned
	// as-is with no refresh flow
	if !isJWT(tokens.AccessToken) {
		return tokens.AccessToken, nil
	}

	if !tokenExpiringSoon(tokens.AccessToken) {
		return tokens.AccessToken, nil
	}

	if tokens.RefreshToken == "" {
		return "", fmt.Errorf("session expired: please run 'sprintctl login' again")
	}

	response, err := NewClient().RefreshToken(tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("session expired: please run 'sprintctl login' again")
	}

	refreshed := Tokens{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	}
	if err := StoreTokens(refreshed); err != nil {
		return "", err
	}

	return refreshed.AccessToken, nil
}

// isJWT reports whether the token looks like a three-part JWT (as issued by
// Supabase GoTrue). Opaque BetterAuth session tokens contain no dots.
func isJWT(token string) bool {
	return len(strings.Split(token, ".")) == 3
}

// tokenExpiringSoon reports whether the JWT expires within expiryLeeway.
// Tokens that can't be decoded are treated as still valid so the server
// stays the authority on rejecting them.
func tokenExpiringSoon(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(payload, &claims) != nil || claims.Exp == 0 {
		return false
	}

	return time.Now().Add(expiryLeeway).After(time.Unix(claims.Exp, 0))
}
