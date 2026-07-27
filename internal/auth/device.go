package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/supabase-community/gotrue-go/types"
)

// ErrDeviceFlowUnsupported indicates the server doesn't expose the device
// login endpoints, so the caller should fall back to the OTP flow
var ErrDeviceFlowUnsupported = errors.New("browser login is not supported by this server")

// DeviceAuth describes a pending browser login request
type DeviceAuth struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func graderSecret() string {
	if GraderSecret != "" {
		return GraderSecret
	}
	return os.Getenv("SPRINTCTL_GRADER_SECRET")
}

func postJSON(url string, body any) (*http.Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grader-Token", graderSecret())

	return http.DefaultClient.Do(req)
}

func decodeError(resp *http.Response) error {
	var errBody struct {
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&errBody)
	if errBody.Error == "" {
		return fmt.Errorf("response status code %d", resp.StatusCode)
	}
	return fmt.Errorf("%s", errBody.Error)
}

// StartDeviceAuth asks the app to begin a browser login and returns the
// codes the user needs to approve it
func StartDeviceAuth(appRoot string) (*DeviceAuth, error) {
	resp, err := postJSON(appRoot+"/api/auth/device", struct{}{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDeviceFlowUnsupported
	}
	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}

	var da DeviceAuth
	if err := json.NewDecoder(resp.Body).Decode(&da); err != nil {
		return nil, err
	}
	if da.DeviceCode == "" || da.UserCode == "" {
		return nil, fmt.Errorf("invalid response from server")
	}
	return &da, nil
}

// PollDeviceAuth waits for the browser approval, then exchanges the issued
// token hash for a session and stores it
func PollDeviceAuth(appRoot string, da *DeviceAuth) error {
	interval := time.Duration(da.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(da.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		approval, done, err := pollOnce(appRoot, da.DeviceCode)
		if err != nil {
			return err
		}
		if done {
			// Prefer the BetterAuth session token when the server issues one:
			// it is used directly as a bearer token with a server-side sliding
			// expiry, so no GoTrue exchange or refresh flow is needed
			if approval.BAToken != "" {
				return StoreTokens(Tokens{AccessToken: approval.BAToken})
			}
			return exchangeTokenHash(approval.TokenHash)
		}
	}

	return fmt.Errorf("login request expired: please run 'sprintctl login' again")
}

// deviceApproval is the payload returned once the user approves the login.
// BAToken is only set by servers that have migrated to BetterAuth.
type deviceApproval struct {
	TokenHash string `json:"token_hash"`
	BAToken   string `json:"ba_token"`
}

// pollOnce checks the request status; done is true once approval was granted
func pollOnce(appRoot, deviceCode string) (approval deviceApproval, done bool, err error) {
	resp, err := postJSON(appRoot+"/api/auth/device/token", map[string]string{
		"device_code": deviceCode,
	})
	if err != nil {
		return deviceApproval{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return deviceApproval{}, false, decodeError(resp)
	}

	var body struct {
		Status string `json:"status"`
		deviceApproval
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return deviceApproval{}, false, err
	}

	if body.Status == "approved" && (body.TokenHash != "" || body.BAToken != "") {
		return body.deviceApproval, true, nil
	}
	return deviceApproval{}, false, nil
}

// exchangeTokenHash trades the single-use token hash for a session at GoTrue
// and stores the resulting tokens
func exchangeTokenHash(tokenHash string) error {
	response, err := NewClient().Verify(types.VerifyRequest{
		Type:       types.VerificationTypeMagiclink,
		Token:      tokenHash,
		RedirectTo: "https://cloudsprints.com", // Required but not used for CLI
	})
	if err != nil {
		return err
	}
	if response.Error != "" || response.AccessToken == "" {
		return fmt.Errorf("verification failed: %s %s", response.Error, response.ErrorDescription)
	}

	return StoreTokens(Tokens{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}
