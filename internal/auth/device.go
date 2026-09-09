package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ErrDeviceFlowUnsupported indicates the server doesn't expose the device
// login endpoints (an app build older than this CLI)
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

// PollDeviceAuth waits for the browser approval, then stores the session
// token the app minted for this login
func PollDeviceAuth(appRoot string, da *DeviceAuth) error {
	interval := time.Duration(da.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(da.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		token, done, err := pollOnce(appRoot, da.DeviceCode)
		if err != nil {
			return err
		}
		if done {
			return StoreTokens(Tokens{AccessToken: token})
		}
	}

	return fmt.Errorf("login request expired: please run 'sprintctl login' again")
}

// pollOnce checks the request status; done is true once approval was granted
// and the app has handed over the session token
func pollOnce(appRoot, deviceCode string) (token string, done bool, err error) {
	resp, err := postJSON(appRoot+"/api/auth/device/token", map[string]string{
		"device_code": deviceCode,
	})
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, decodeError(resp)
	}

	var body struct {
		Status  string `json:"status"`
		BaToken string `json:"ba_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", false, err
	}

	if body.Status == "approved" {
		if body.BaToken == "" {
			return "", false, fmt.Errorf("login was approved but no session token was returned")
		}
		return body.BaToken, true, nil
	}
	return "", false, nil
}
