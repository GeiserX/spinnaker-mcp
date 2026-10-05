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
