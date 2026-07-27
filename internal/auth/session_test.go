package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func makeJWT(t *testing.T, exp int64) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]int64{"exp": exp})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	return fmt.Sprintf("%s.%s.signature", header, payload)
}

func TestIsJWT(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"supabase JWT", makeJWT(t, time.Now().Add(time.Hour).Unix()), true},
		{"opaque betterauth token", "vXk3mZq8TfB2LwNpYd6RgHs4Ju9AcEo1", false},
		{"empty", "", false},
		{"one dot", "abc.def", false},
		{"three dots", "a.b.c.d", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJWT(tt.token); got != tt.want {
				t.Errorf("isJWT(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestTokenExpiringSoon(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"valid for an hour", makeJWT(t, time.Now().Add(time.Hour).Unix()), false},
		{"expired", makeJWT(t, time.Now().Add(-time.Hour).Unix()), true},
		{"expires within leeway", makeJWT(t, time.Now().Add(30*time.Second).Unix()), true},
		{"not a JWT", "some-opaque-token", false},
		{"garbage payload", "aaa.bbb.ccc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tokenExpiringSoon(tt.token); got != tt.want {
				t.Errorf("tokenExpiringSoon() = %v, want %v", got, tt.want)
			}
		})
	}
}
