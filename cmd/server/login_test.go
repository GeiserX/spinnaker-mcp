package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ssoGate(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "SESSION=good" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.URL.Path == "/auth/user" {
			w.Write([]byte(`{"email":"sergio@example.com","username":"sergio","roles":[]}`))
			return
		}
		w.Write([]byte(`[]`))
	}))
}

func TestRunLogin_StoresVerifiedCookie(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	path := filepath.Join(t.TempDir(), "cookies", "gate")
	var stderr bytes.Buffer

	code := runLogin([]string{"--gate=" + srv.URL, "--cookie=good", "--cookie-file=" + path, "--no-browser"}, strings.NewReader(""), &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr.String())
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cookie file: %v", err)
	}
	if strings.TrimSpace(string(b)) != "SESSION=good" {
		t.Errorf("stored %q, want SESSION=good", string(b))
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("cookie file mode %o, want 0600", info.Mode().Perm())
	}
	if !strings.Contains(stderr.String(), "as sergio") {
		t.Errorf("expected the username in the confirmation, got: %s", stderr.String())
	}
	if strings.Contains(stderr.String(), "good") {
		t.Errorf("the cookie value must never be printed: %s", stderr.String())
	}
}

func TestRunLogin_ReadsCookieFromStdin(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	path := filepath.Join(t.TempDir(), "gate")
	var stderr bytes.Buffer

	code := runLogin([]string{"--gate=" + srv.URL, "--cookie-file=" + path, "--no-browser"}, strings.NewReader("good\n"), &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), srv.URL+"/login") {
		t.Errorf("expected the login URL in the instructions, got: %s", stderr.String())
	}
}

func TestRunLogin_RejectsBadCookie(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	path := filepath.Join(t.TempDir(), "gate")
	var stderr bytes.Buffer

	code := runLogin([]string{"--gate=" + srv.URL, "--cookie=bad", "--cookie-file=" + path, "--no-browser"}, strings.NewReader(""), &stderr)
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", code, stderr.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a rejected cookie must not be stored")
	}
	if !strings.Contains(stderr.String(), "login session") {
		t.Errorf("expected the login-required error, got: %s", stderr.String())
	}
}

func TestRunLogin_EmptyUserBodyIsNotSignedIn(t *testing.T) {
	// Some Gates answer an anonymous /auth/user with 200 and an empty body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	var stderr bytes.Buffer
	code := runLogin([]string{"--gate=" + srv.URL, "--cookie=whatever", "--cookie-file=" + filepath.Join(t.TempDir(), "c"), "--no-browser"}, strings.NewReader(""), &stderr)
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "did not recognise") {
		t.Errorf("unexpected error text: %s", stderr.String())
	}
}

func TestRunLogin_InvalidGateURL(t *testing.T) {
	var stderr bytes.Buffer
	if code := runLogin([]string{"--gate=ftp://nope", "--cookie=x", "--no-browser"}, strings.NewReader(""), &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
}

func TestRunLogin_OpensBrowserWhenPrompting(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	var opened []string
	orig := startCommand
	startCommand = func(name string, args ...string) error {
		opened = append(opened, name+" "+strings.Join(args, " "))
		return nil
	}
	defer func() { startCommand = orig }()

	var stderr bytes.Buffer
	code := runLogin([]string{"--gate=" + srv.URL, "--cookie-file=" + filepath.Join(t.TempDir(), "c")}, strings.NewReader("good\n"), &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr.String())
	}
	if len(opened) != 1 || !strings.HasSuffix(opened[0], srv.URL+"/login") {
		t.Errorf("expected one browser launch to the login page, got %v", opened)
	}
}

func TestOpenBrowser_ReportsLaunchFailure(t *testing.T) {
	orig := startCommand
	startCommand = func(string, ...string) error { return os.ErrNotExist }
	defer func() { startCommand = orig }()
	openBrowser("https://gate.example.com/login") // must not panic or exit; the error goes to stderr
}

func TestBrowserCommand(t *testing.T) {
	cases := map[string]string{"darwin": "open", "windows": "rundll32", "linux": "xdg-open", "freebsd": "xdg-open"}
	for goos, want := range cases {
		name, args := browserCommand(goos, "https://x/login")
		if name != want || args[len(args)-1] != "https://x/login" {
			t.Errorf("browserCommand(%s) = %s %v", goos, name, args)
		}
	}
}

func TestRunLogin_NoCookieOnStdin(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	for _, in := range []string{"", "   \n"} {
		var stderr bytes.Buffer
		code := runLogin([]string{"--gate=" + srv.URL, "--no-browser", "--cookie-file=" + filepath.Join(t.TempDir(), "c")}, strings.NewReader(in), &stderr)
		if code != 1 || !strings.Contains(stderr.String(), "no cookie given") {
			t.Errorf("stdin %q: exit %d, stderr %s", in, code, stderr.String())
		}
	}
}

func TestRunLogin_BadFlag(t *testing.T) {
	var stderr bytes.Buffer
	if code := runLogin([]string{"--definitely-not-a-flag"}, strings.NewReader(""), &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
}

func TestRunLogin_UnwritableCookieFile(t *testing.T) {
	srv := ssoGate(t)
	defer srv.Close()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	// the parent of the cookie path is a regular file, so MkdirAll fails
	code := runLogin([]string{"--gate=" + srv.URL, "--cookie=good", "--no-browser", "--cookie-file=" + filepath.Join(blocker, "cookie")}, strings.NewReader(""), &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "creating") {
		t.Errorf("exit %d, stderr %s", code, stderr.String())
	}
	// the cookie path itself is a directory, so WriteFile fails
	if err := writeCookieFile(dir, "SESSION=x"); err == nil || !strings.Contains(err.Error(), "writing") {
		t.Errorf("expected a write error for a directory path, got %v", err)
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("SPINNAKER_MCP_TEST_ENVOR", "")
	if got := envOr("SPINNAKER_MCP_TEST_ENVOR", "d"); got != "d" {
		t.Errorf("got %q, want default", got)
	}
	t.Setenv("SPINNAKER_MCP_TEST_ENVOR", "v")
	if got := envOr("SPINNAKER_MCP_TEST_ENVOR", "d"); got != "v" {
		t.Errorf("got %q, want v", got)
	}
}
