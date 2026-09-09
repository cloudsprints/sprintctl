package auth

import (
	"net/http"
)

// GraderSecret is injected at build time via:
//
//	-ldflags "-X github.com/cloudsprints/sprintctl/internal/auth.GraderSecret=VALUE"
//
// Falls back to SPRINTCTL_GRADER_SECRET env var for local development.
var GraderSecret string

// Logout revokes the session server-side (best effort) and removes the
// stored token. appRoot is the app origin (api_base_url without /api/v1).
func Logout(appRoot string) error {
	if tokens, err := GetTokens(); err == nil && tokens.AccessToken != "" && !isLegacyJWT(tokens.AccessToken) {
		// Ignore errors: the session may already be expired or revoked, and
		// local cleanup should proceed regardless
		if req, err := http.NewRequest("POST", appRoot+"/api/ba/sign-out", nil); err == nil {
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			if resp, err := http.DefaultClient.Do(req); err == nil {
				resp.Body.Close()
			}
		}
	}
	return DeleteToken()
}

// IsAuthenticated reports whether a usable session token is stored. Tokens
// saved by releases that authenticated against the old identity provider
// (JWTs) are not usable and count as signed out so `login` can replace them.
func IsAuthenticated() bool {
	tokens, err := GetTokens()
	if err != nil {
		return false
	}
	return !isLegacyJWT(tokens.AccessToken)
}
