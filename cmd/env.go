package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cloudsprints/sprintctl/internal/auth"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Named environments. The CLI has to know where to talk before it can ask a
// server anything, so these live in the binary; --api-base-url (or
// SPRINTCTL_API_BASE_URL) remains the escape hatch for any other host.
var environments = map[string]string{
	"prod":    "https://cloudsprints.com/api/v1",
	"staging": "https://staging.cloudsprints.com/api/v1",
	"local":   "http://localhost:5173/api/v1",
}

var envAliases = map[string]string{
	"production": "prod",
	"localhost":  "local",
	"dev":        "local",
}

const defaultEnv = "prod"

// environment is the resolved target for this invocation.
type environment struct {
	Name   string // prod | staging | local | custom
	URL    string // API base URL (…/api/v1)
	Source string // where the choice came from, for `sprintctl env`
}

var currentEnv environment

func canonicalEnvName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if alias, ok := envAliases[name]; ok {
		return alias
	}
	return name
}

func envNameForURL(url string) string {
	url = strings.TrimRight(url, "/")
	for name, u := range environments {
		if u == url {
			return name
		}
	}
	return "custom"
}

// resolveEnvironment picks the API target. Precedence:
//  1. an explicit URL (--api-base-url flag or SPRINTCTL_API_BASE_URL) — how
//     lab boxes are pointed at their own environment
//  2. a named environment (--env flag, SPRINTCTL_ENV, or `env` in config)
//  3. a legacy config `api_base_url` written by releases before profiles
//  4. prod
func resolveEnvironment(explicitURL, envName, legacyURL string) environment {
	if explicitURL != "" {
		return environment{Name: envNameForURL(explicitURL), URL: strings.TrimRight(explicitURL, "/"), Source: "url override"}
	}
	if name := canonicalEnvName(envName); name != "" {
		if url, ok := environments[name]; ok {
			return environment{Name: name, URL: url, Source: "config"}
		}
		if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
			return environment{Name: "custom", URL: strings.TrimRight(name, "/"), Source: "config"}
		}
	}
	if legacyURL != "" {
		return environment{Name: envNameForURL(legacyURL), URL: strings.TrimRight(legacyURL, "/"), Source: "legacy config"}
	}
	return environment{Name: defaultEnv, URL: environments[defaultEnv], Source: "default"}
}

// apiBaseURL is what every command uses to build its API client.
func apiBaseURL() string {
	return currentEnv.URL
}

// appRoot is the app origin (API base without /api/v1), for auth endpoints.
func appRoot() string {
	return strings.TrimSuffix(apiBaseURL(), "/api/v1")
}

// tokenScopeFor keeps one stored session per environment so switching never
// sends a staging token to prod (or vice versa). prod keeps the historical
// unscoped entry so existing logins survive the upgrade. Inside a lab box the
// entrypoint seeds the unscoped entry and points the CLI at the box's own
// environment via SPRINTCTL_API_BASE_URL, so boxes always use that entry.
func tokenScopeFor(env environment) string {
	if os.Getenv(machineLessonEnv) != "" || os.Getenv("SPRINTCTL_TOKEN") != "" {
		return auth.DefaultTokenScope
	}
	if env.Name == "prod" {
		return auth.DefaultTokenScope
	}
	if env.Name != "custom" {
		return env.Name
	}
	host := env.URL
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	return strings.NewReplacer(":", "-", "/", "-").Replace(host)
}

// envLine is the one-line environment summary shown by login/status.
func envLine() string {
	return fmt.Sprintf("Environment: %s (%s)", currentEnv.Name, currentEnv.URL)
}

var envCmd = &cobra.Command{
	Use:   "env [prod|staging|local|<url>]",
	Short: "Show or switch the CloudSprints environment",
	Long: `Show which CloudSprints environment sprintctl talks to, or switch it.

Environments: prod (default), staging, local (http://localhost:5173).
A full URL selects a custom environment. The choice is saved to the config
file; --env or SPRINTCTL_ENV override it for a single command, and
--api-base-url / SPRINTCTL_API_BASE_URL override everything.

Sessions are stored per environment: after switching, run 'sprintctl login'
once for the new environment.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printEnvStatus()
			return
		}

		target := canonicalEnvName(args[0])
		url, known := environments[target]
		isURL := strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")
		if !known && !isURL {
			names := make([]string, 0, len(environments))
			for n := range environments {
				names = append(names, n)
			}
			sort.Strings(names)
			fmt.Println(styles.ErrorStyle.Render(" UNKNOWN ENVIRONMENT "))
			fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("%q is not an environment.\nUse one of: %s, or a full URL ending in /api/v1.", args[0], strings.Join(names, ", "))))
			os.Exit(1)
		}
		if isURL {
			url = strings.TrimRight(target, "/")
			target = "custom"
		}

		viper.Set("env", target)
		viper.Set("api_base_url", url)
		if err := viper.WriteConfig(); err != nil {
			fmt.Println(styles.ErrorStyle.Render(" CONFIG ERROR "), err)
			os.Exit(1)
		}

		currentEnv = environment{Name: target, URL: url, Source: "config"}
		auth.TokenScope = tokenScopeFor(currentEnv)

		fmt.Println(styles.SuccessStyle.Render(" ENVIRONMENT SET "))
		hint := "You are signed in to this environment."
		if !auth.IsAuthenticated() {
			hint = "Run 'sprintctl login' to sign in to this environment."
		}
		fmt.Println(styles.BoxStyle.Render(envLine() + "\n\n" + hint))
	},
}

func printEnvStatus() {
	signedIn := "not signed in"
	if auth.IsAuthenticated() {
		signedIn = "signed in"
	}
	fmt.Println(styles.InfoStyle.Render(" ENVIRONMENT "))
	fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("%s\nSource: %s\nSession: %s\n\nSwitch with: sprintctl env prod | staging | local", envLine(), currentEnv.Source, signedIn)))
}

func init() {
	rootCmd.AddCommand(envCmd)
}
