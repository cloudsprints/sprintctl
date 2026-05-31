package auth

import (
	"os"
	"path/filepath"
	"strings"
	
	"github.com/zalando/go-keyring"
)

const (
	service = "sprintctl"
	account = "default"
)

// StoreToken stores the access token securely in the system keychain
// Falls back to file storage if keyring is unavailable (e.g., no D-Bus)
func StoreToken(token string) error {
	err := keyring.Set(service, account, token)
	if err != nil && isKeyringUnavailable(err) {
		return storeTokenFile(token)
	}
	return err
}

// GetToken retrieves the access token from the system keychain
// Falls back to file storage if keyring is unavailable
func GetToken() (string, error) {
	token, err := keyring.Get(service, account)
	if err != nil && isKeyringUnavailable(err) {
		return getTokenFile()
	}
	return token, err
}

// DeleteToken removes the access token from the system keychain
// Falls back to file storage if keyring is unavailable
func DeleteToken() error {
	err := keyring.Delete(service, account)
	if err != nil && isKeyringUnavailable(err) {
		return deleteTokenFile()
	}
	return err
}

// HasToken checks if a token exists in the keychain
func HasToken() bool {
	_, err := GetToken()
	return err == nil
}

// isKeyringUnavailable checks if the error is due to keyring being unavailable
func isKeyringUnavailable(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "dbus") || 
	       strings.Contains(errStr, "d-bus") ||
	       strings.Contains(errStr, "secret service") ||
	       strings.Contains(errStr, "/run/user") ||
	       strings.Contains(errStr, "no such file or directory") ||
	       strings.Contains(errStr, "connection refused")
}

// getTokenFilePath returns the path to the token file
func getTokenFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "sprintctl")
	return filepath.Join(configDir, ".token"), nil
}

// storeTokenFile stores the token in a file
func storeTokenFile(token string) error {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return err
	}
	
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	
	// Write token to file with restricted permissions
	return os.WriteFile(tokenPath, []byte(token), 0600)
}

// getTokenFile retrieves the token from a file
func getTokenFile() (string, error) {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}
	
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// deleteTokenFile removes the token file
func deleteTokenFile() error {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return err
	}
	
	// Ignore error if file doesn't exist
	err = os.Remove(tokenPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
