package config

import (
	"os"
	"testing"
)

func TestLoadGateConfig_Defaults(t *testing.T) {
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
