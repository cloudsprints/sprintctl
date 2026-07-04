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
