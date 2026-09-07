package auth

import "testing"

func TestIsLegacyJWT(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"betterauth session token", "eYeAft-o8KJ2ZkfyzvErk3M5Qs4Tr2cU", false},
		{"empty", "", false},
		{"legacy jwt", "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjF9.signature", true},
		{"two segments", "aaa.bbb", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLegacyJWT(tt.token); got != tt.want {
				t.Errorf("isLegacyJWT(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}
