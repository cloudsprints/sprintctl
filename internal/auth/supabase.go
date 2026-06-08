package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	graderSecret := GraderSecret
	if graderSecret == "" {
		graderSecret = os.Getenv("SPRINTCTL_GRADER_SECRET")
	}

	body, err := json.Marshal(map[string]string{"email": email})
	if err != nil {
		return err
	}

	if proxyURL == "" {
		proxyURL = AppCLIOTPURL
	}

	req, err := http.NewRequest("POST", proxyURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grader-Token", graderSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("response status code %d: %s", resp.StatusCode, errBody.Error)
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
				
				// Parse the JSON to extract access token
				var tokenData struct {
					AccessToken string `json:"access_token"`
				}
				if json.Unmarshal([]byte(jsonStr), &tokenData) == nil && tokenData.AccessToken != "" {
					return StoreToken(tokenData.AccessToken)
				}
			}
		}
		return err
	}
	
	// Store the access token securely
	return StoreToken(response.AccessToken)
}

// Logout removes the stored token
func Logout() error {
	return DeleteToken()
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated() bool {
	return HasToken()
}
