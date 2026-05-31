package auth

import (
	"encoding/json"
	"strings"
	
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
)

const (
	SupabaseProjectRef = "vlryocooywgsquopuvkp"
	SupabaseAnonKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InZscnlvY29veXdnc3F1b3B1dmtwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTE5OTIxNzIsImV4cCI6MjA2NzU2ODE3Mn0.7wwkIxLbLr0XC_YkXeBtxQ_epsjVNmEI4XXGDTCy7MY"
)

// NewClient creates a new GoTrue client
func NewClient() gotrue.Client {
	return gotrue.New(SupabaseProjectRef, SupabaseAnonKey)
}

// LoginWithOTP initiates the OTP login flow
func LoginWithOTP(email string) error {
	client := NewClient()
	
	err := client.OTP(types.OTPRequest{
		Email:      email,
		CreateUser: false,
	})
	return err
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
