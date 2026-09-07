package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppCLIOTPURL is the default email-code login endpoint (overridden per
// environment via SPRINTCTL_API_BASE_URL by the login command).
const AppCLIOTPURL = "https://cloudsprints.com/api/auth/cli-otp"

// GraderSecret is injected at build time via:
//
//	-ldflags "-X github.com/cloudsprints/sprintctl/internal/auth.GraderSecret=VALUE"
//
// Falls back to SPRINTCTL_GRADER_SECRET env var for local development.
var GraderSecret string

// LoginWithOTP asks the app to email a one-time sign-in code
func LoginWithOTP(email, proxyURL string) error {
	if proxyURL == "" {
		proxyURL = AppCLIOTPURL
	}

	resp, err := postJSON(proxyURL, map[string]string{"email": email})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decodeError(resp)
	}
	return nil
}

// VerifyOTP exchanges the emailed code for a session token at the app and
// stores it. The token is an opaque BetterAuth session token (30-day sliding
// expiry, renewed server-side on use) sent as `Authorization: Bearer`.
func VerifyOTP(email, code, proxyURL string) error {
	if proxyURL == "" {
		proxyURL = AppCLIOTPURL
	}

	resp, err := postJSON(proxyURL, map[string]string{"email": email, "otp": code})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decodeError(resp)
	}

	var body struct {
		BaToken     string `json:"ba_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	token := body.BaToken
	if token == "" {
		token = body.AccessToken
	}
	if token == "" {
		return fmt.Errorf("verification succeeded but no session token was returned")
	}

	return StoreTokens(Tokens{AccessToken: token})
}

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
