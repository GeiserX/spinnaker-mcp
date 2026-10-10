package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGateConfig_Defaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GATE_COOKIE", "")
	t.Setenv("GATE_COOKIE_FILE", "")
	os.Unsetenv("GATE_URL")
	os.Unsetenv("GATE_TOKEN")
	os.Unsetenv("GATE_USER")
	os.Unsetenv("GATE_PASS")
	os.Unsetenv("GATE_CERT_FILE")
	os.Unsetenv("GATE_KEY_FILE")
	os.Unsetenv("GATE_INSECURE")

	cfg := LoadGateConfig()
	if cfg.BaseURL != "http://localhost:8084" {
		t.Errorf("expected default BaseURL, got %q", cfg.BaseURL)
	}
	if cfg.Token != "" {
		t.Errorf("expected empty Token, got %q", cfg.Token)
	}
	if cfg.User != "" {
		t.Errorf("expected empty User, got %q", cfg.User)
	}
	if cfg.Pass != "" {
		t.Errorf("expected empty Pass, got %q", cfg.Pass)
	}
	if cfg.Insecure {
		t.Error("expected Insecure=false by default")
	}
}

func TestLoadGateConfig_EnvOverrides(t *testing.T) {
	os.Setenv("GATE_URL", "https://gate.example.com")
	os.Setenv("GATE_TOKEN", "my-token")
	os.Setenv("GATE_INSECURE", "true")
	defer func() {
		os.Unsetenv("GATE_URL")
		os.Unsetenv("GATE_TOKEN")
		os.Unsetenv("GATE_INSECURE")
	}()

	cfg := LoadGateConfig()
	if cfg.BaseURL != "https://gate.example.com" {
		t.Errorf("expected overridden BaseURL, got %q", cfg.BaseURL)
	}
	if cfg.Token != "my-token" {
		t.Errorf("expected overridden Token, got %q", cfg.Token)
	}
	if !cfg.Insecure {
		t.Error("expected Insecure=true")
	}
}

func TestLoadGateConfig_CookieFromEnv(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GATE_COOKIE", " SESSION=abc ")
	t.Setenv("GATE_COOKIE_FILE", "")
	if got := LoadGateConfig().Cookie; got != "SESSION=abc" {
		t.Errorf("Cookie = %q, want SESSION=abc", got)
	}
}

func TestLoadGateConfig_CookieFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookie")
	if err := os.WriteFile(path, []byte("SESSION=fromfile\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("GATE_COOKIE", "")
	t.Setenv("GATE_COOKIE_FILE", path)
	if got := LoadGateConfig().Cookie; got != "SESSION=fromfile" {
		t.Errorf("Cookie = %q, want SESSION=fromfile", got)
	}
	t.Setenv("GATE_COOKIE_FILE", filepath.Join(dir, "missing"))
	if got := LoadGateConfig().Cookie; got != "" {
		t.Errorf("a missing GATE_COOKIE_FILE must yield no cookie, got %q", got)
	}
}

func TestLoadGateConfig_CookieFromDefaultFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("GATE_COOKIE", "")
	t.Setenv("GATE_COOKIE_FILE", "")
	t.Setenv("GATE_URL", "https://gate.example.com:8443")
	want := filepath.Join(dir, "spinnaker-mcp", "cookies", "gate.example.com_8443")
	if got := DefaultCookieFile("https://gate.example.com:8443"); got != want {
		t.Fatalf("DefaultCookieFile = %q, want %q", got, want)
	}
	if err := os.MkdirAll(filepath.Dir(want), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, []byte("SESSION=default\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := LoadGateConfig().Cookie; got != "SESSION=default" {
		t.Errorf("Cookie = %q, want SESSION=default", got)
	}
	if DefaultCookieFile("not a url") != "" || DefaultCookieFile("/relative") != "" {
		t.Error("DefaultCookieFile must be empty for a URL without a host")
	}
}

func TestLoadGateConfig_NoCookieWithoutHostOrHome(t *testing.T) {
	t.Setenv("GATE_COOKIE", "")
	t.Setenv("GATE_COOKIE_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GATE_URL", "not a url")
	if got := LoadGateConfig().Cookie; got != "" {
		t.Errorf("a Gate URL without a host must yield no cookie, got %q", got)
	}
	t.Setenv("GATE_URL", "https://gate.example.com")
	t.Setenv("HOME", "")
	if got := DefaultCookieFile("https://gate.example.com"); got != "" {
		t.Errorf("without HOME and XDG_CONFIG_HOME there is no default file, got %q", got)
	}
}

func TestGetenv_UnexpandedPlaceholderReadsAsUnset(t *testing.T) {
	cases := map[string]string{
		"${GATE_TOKEN}":    "",
		"${TOOLSETS}":      "",
		" ${OTHER_NAME} ":  "",
		"real-token":       "real-token",
		"tok${GATE_TOKEN}": "tok${GATE_TOKEN}",
		"${not a name}":    "${not a name}",
		"$GATE_TOKEN":      "$GATE_TOKEN",
		"":                 "",
	}
	for in, want := range cases {
		t.Setenv("SPINNAKER_MCP_TEST_VAR", in)
		if got := Getenv("SPINNAKER_MCP_TEST_VAR"); got != want {
			t.Errorf("Getenv with %q = %q, want %q", in, got, want)
		}
	}
}

func TestLoadGateConfig_PlaceholderValuesAreUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, k := range []string{"GATE_URL", "GATE_TOKEN", "GATE_USER", "GATE_PASS", "GATE_CERT_FILE", "GATE_KEY_FILE", "GATE_INSECURE", "GATE_COOKIE", "GATE_COOKIE_FILE"} {
		t.Setenv(k, "${"+k+"}")
	}

	cfg := LoadGateConfig()
	if cfg.BaseURL != "http://localhost:8084" {
		t.Errorf("BaseURL = %q, want the default", cfg.BaseURL)
	}
	if cfg.Token != "" || cfg.User != "" || cfg.Pass != "" || cfg.CertFile != "" || cfg.KeyFile != "" || cfg.Cookie != "" {
		t.Errorf("placeholder leaked into config: %+v", cfg)
	}
}
