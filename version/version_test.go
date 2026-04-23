package version

import "testing"

func TestString_Defaults(t *testing.T) {
	s := String()
	if s != "dev (none) unknown" {
		t.Errorf("expected default version string, got %q", s)
	}
}

func TestString_Custom(t *testing.T) {
	Version = "1.0.0"
	Commit = "abc1234"
	Date = "2026-04-23"
	defer func() {
		Version = "dev"
		Commit = "none"
		Date = "unknown"
	}()

	s := String()
	if s != "1.0.0 (abc1234) 2026-04-23" {
		t.Errorf("expected custom version string, got %q", s)
	}
}
