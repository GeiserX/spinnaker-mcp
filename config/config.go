package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type GateConfig struct {
	BaseURL  string
	Token    string
	User     string
	Pass     string
	CertFile string
	KeyFile  string
	Insecure bool
	// Cookie is the session cookie for a Gate behind SSO. It comes from GATE_COOKIE, else
	// the file named by GATE_COOKIE_FILE, else the per-host file `spinnaker-mcp login` writes.
	Cookie string
}

func LoadGateConfig() GateConfig {
	base := getEnv("GATE_URL", "http://localhost:8084")
	return GateConfig{
		BaseURL:  base,
		Token:    getEnv("GATE_TOKEN", ""),
		User:     getEnv("GATE_USER", ""),
		Pass:     getEnv("GATE_PASS", ""),
		CertFile: getEnv("GATE_CERT_FILE", ""),
		KeyFile:  getEnv("GATE_KEY_FILE", ""),
		Insecure: getEnv("GATE_INSECURE", "") == "true",
		Cookie:   loadCookie(base),
	}
}

// loadCookie resolves the session cookie. GATE_COOKIE wins; then the file named by
// GATE_COOKIE_FILE; then the default cookie file for the Gate host, if it exists.
func loadCookie(gateURL string) string {
	if v := strings.TrimSpace(os.Getenv("GATE_COOKIE")); v != "" {
		return v
	}
	path := os.Getenv("GATE_COOKIE_FILE")
	if path == "" {
		path = DefaultCookieFile(gateURL)
	}
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// DefaultCookieFile is where `spinnaker-mcp login` stores the session cookie for a Gate:
// $XDG_CONFIG_HOME/spinnaker-mcp/cookies/<host>, with XDG_CONFIG_HOME defaulting to ~/.config.
// It returns "" when the Gate URL has no host or no home directory can be found.
func DefaultCookieFile(gateURL string) string {
	u, err := url.Parse(gateURL)
	if err != nil || u.Host == "" {
		return ""
	}
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "spinnaker-mcp", "cookies", strings.ReplaceAll(u.Host, ":", "_"))
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
