package cmd

import (
	"testing"

	"github.com/cloudsprints/sprintctl/internal/auth"
)

func TestResolveEnvironmentPrecedence(t *testing.T) {
	tests := []struct {
		name                          string
		explicitURL, envName, legacy  string
		wantName, wantURL, wantSource string
	}{
		{"default is prod", "", "", "", "prod", environments["prod"], "default"},
		{"named env", "", "staging", "", "staging", environments["staging"], "config"},
		{"alias", "", "localhost", "", "local", environments["local"], "config"},
		{"explicit url beats env", "https://staging.cloudsprints.com/api/v1/", "prod", "", "staging", environments["staging"], "url override"},
		{"explicit unknown url is custom", "http://10.0.0.5:5173/api/v1", "prod", "", "custom", "http://10.0.0.5:5173/api/v1", "url override"},
		{"env beats legacy config url", "", "local", environments["prod"], "local", environments["local"], "config"},
		{"legacy config url maps to a profile", "", "", "https://staging.cloudsprints.com/api/v1", "staging", environments["staging"], "legacy config"},
		{"legacy custom url", "", "", "https://foo.example/api/v1", "custom", "https://foo.example/api/v1", "legacy config"},
		{"unknown env name falls through", "", "nope", environments["staging"], "staging", environments["staging"], "legacy config"},
		{"url as env name is custom", "", "https://x.example/api/v1", "", "custom", "https://x.example/api/v1", "config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveEnvironment(tt.explicitURL, tt.envName, tt.legacy)
			if got.Name != tt.wantName || got.URL != tt.wantURL || got.Source != tt.wantSource {
				t.Errorf("got %+v, want name=%q url=%q source=%q", got, tt.wantName, tt.wantURL, tt.wantSource)
			}
		})
	}
}

func TestTokenScopeFor(t *testing.T) {
	t.Setenv(machineLessonEnv, "")
	t.Setenv("SPRINTCTL_TOKEN", "")
	if s := tokenScopeFor(environment{Name: "prod"}); s != auth.DefaultTokenScope {
		t.Errorf("prod scope = %q, want the default entry", s)
	}
	if s := tokenScopeFor(environment{Name: "staging"}); s != "staging" {
		t.Errorf("staging scope = %q", s)
	}
	if s := tokenScopeFor(environment{Name: "custom", URL: "http://10.0.0.5:5173/api/v1"}); s != "10.0.0.5-5173" {
		t.Errorf("custom scope = %q", s)
	}
	// Lab boxes: the entrypoint seeds the default entry and points the CLI at
	// the box's environment — the seeded token must be used regardless.
	t.Setenv(machineLessonEnv, "cmenvenvenvenvenv")
	if s := tokenScopeFor(environment{Name: "staging"}); s != auth.DefaultTokenScope {
		t.Errorf("box scope = %q, want the default entry", s)
	}
}
