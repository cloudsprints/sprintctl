package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
)

const (
	SupabaseProjectRef = "vlryocooywgsquopuvkp"
	SupabaseAnonKey    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InZscnlvY29veXdnc3F1b3B1dmtwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTE5OTIxNzIsImV4cCI6MjA2NzU2ODE3Mn0.7wwkIxLbLr0XC_YkXeBtxQ_epsjVNmEI4XXGDTCy7MY"
	AppCLIOTPURL       = "https://cloudsprints.com/api/auth/cli-otp"
)

// GraderSecret is injected at build time via:
//
//	-ldflags "-X github.com/cloudsprints/sprintctl/internal/auth.GraderSecret=VALUE"
//
// Falls back to SPRINTCTL_GRADER_SECRET env var for local development.
var GraderSecret string

// NewClient creates a new GoTrue client
func NewClient() gotrue.Client {
	return gotrue.New(SupabaseProjectRef, SupabaseAnonKey)
}

// LoginWithOTP initiates the OTP login flow via the app proxy (bypasses Supabase captcha)
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
		return fmt.Errorf("response status code %d: %s", resp.StatusCode, decodeError(resp))
	}

	return nil
}

// VerifyOTP verifies the OTP code and stores the token
func VerifyOTP(email, code string) error {
	client := NewClient()

	response, err := client.VerifyForUser(types.VerifyForUserRequest{
		Type:       "email",
		Token:      code,
		Email:      email,
		RedirectTo: "https://cloudsprints.com", // Required but not used for CLI
	})

	// Debug: Check what we actually got
	if err != nil {
		// Check if the error message contains a valid access token (GoTrue client bug)
		errStr := err.Error()
		if strings.Contains(errStr, "access_token") && strings.Contains(errStr, "response status code 200") {
			// Extract the JSON from the error message
			start := strings.Index(errStr, `{"access_token"`)
			if start != -1 {
				jsonStr := errStr[start:]

				// Parse the JSON to extract the session tokens
				var tokenData Tokens
				if json.Unmarshal([]byte(jsonStr), &tokenData) == nil && tokenData.AccessToken != "" {
					return StoreTokens(tokenData)
				}
			}
		}
		return err
	}

	// Store the session tokens securely
	return StoreTokens(Tokens{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}

// Logout revokes the session server-side (best effort) and removes the
// stored tokens
func Logout() error {
	if tokens, err := GetTokens(); err == nil && tokens.AccessToken != "" {
		// Ignore errors: the token may already be expired or revoked, and
		// local cleanup should proceed regardless
		_ = NewClient().WithToken(tokens.AccessToken).Logout()
	}
	return DeleteToken()
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated() bool {
	return HasToken()
}
