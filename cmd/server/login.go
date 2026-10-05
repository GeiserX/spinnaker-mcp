package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
)

// runLogin implements `spinnaker-mcp login`, for a Gate that sits behind SSO (OAuth2, SAML).
// Gate only accepts the browser session there, so the user signs in once in a browser, pastes
// the SESSION cookie, and the MCP replays it. The cookie is verified against GET /auth/user
// before it is stored, with mode 0600, in the per-host file the server reads by default.
func runLogin(args []string, stdin io.Reader, stderr io.Writer) int {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	gateURL := fs.String("gate", envOr("GATE_URL", "http://localhost:8084"), "Gate URL (default: GATE_URL)")
	cookieFile := fs.String("cookie-file", "", "where to store the cookie (default: $XDG_CONFIG_HOME/spinnaker-mcp/cookies/<host>)")
	cookie := fs.String("cookie", "", "SESSION cookie value (default: read one line from stdin)")
	noBrowser := fs.Bool("no-browser", false, "do not open the Gate login page in a browser")
	insecure := fs.Bool("insecure", os.Getenv("GATE_INSECURE") == "true", "skip TLS certificate verification")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	base := strings.TrimRight(*gateURL, "/")
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		fmt.Fprintf(stderr, "login: invalid Gate URL %q\n", *gateURL)
		return 2
	}

	value := strings.TrimSpace(*cookie)
	if value == "" {
		loginURL := base + "/login"
		if !*noBrowser {
			openBrowser(loginURL)
		}
		fmt.Fprintf(stderr, "Sign in to Spinnaker in your browser:\n  %s\n\nThen copy the SESSION cookie for %s (DevTools, Application tab, Cookies) and paste it here.\nSESSION cookie: ", loginURL, u.Host)
		line, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil && line == "" {
			fmt.Fprintln(stderr, "\nlogin: no cookie given")
			return 1
		}
		value = strings.TrimSpace(line)
		if value == "" {
			fmt.Fprintln(stderr, "login: no cookie given")
			return 1
		}
	}
	value = client.NormalizeCookie(value)

	gate, err := client.NewGate(client.GateOptions{BaseURL: base, Cookie: value, Insecure: *insecure})
	if err != nil {
		fmt.Fprintf(stderr, "login: %v\n", err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	user, err := verifySession(ctx, gate)
	if err != nil {
		fmt.Fprintf(stderr, "login: %v\n", err)
		return 1
	}

	path := *cookieFile
	if path == "" {
		path = config.DefaultCookieFile(base)
	}
	if path == "" {
		fmt.Fprintln(stderr, "login: cannot derive a cookie file path; pass --cookie-file")
		return 1
	}
	if err := writeCookieFile(path, value); err != nil {
		fmt.Fprintf(stderr, "login: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "Signed in to %s as %s. Session cookie stored in %s\nspinnaker-mcp reads it automatically when GATE_URL is %s; run login again when the session expires.\n", u.Host, user, path, base)
	return 0
}

// verifySession asks Gate who the cookie belongs to. A Gate behind SSO answers an anonymous
// call to /auth/user with a redirect or an empty body, so both count as "not signed in".
func verifySession(ctx context.Context, gate *client.GateClient) (string, error) {
	body, err := gate.CurrentUser(ctx)
	if err != nil {
		return "", err
	}
	var who struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if len(bytes.TrimSpace(body)) == 0 || json.Unmarshal(body, &who) != nil || (who.Username == "" && who.Email == "") {
		return "", errors.New("gate did not recognise the session cookie (GET /auth/user returned no user); copy it again after signing in")
	}
	if who.Username != "" {
		return who.Username, nil
	}
	return who.Email, nil
}

func writeCookieFile(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(value+"\n"), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return os.Chmod(path, 0o600)
}

// startCommand launches a detached command; tests replace it.
var startCommand = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

// browserCommand returns the OS command that opens a URL in the default browser.
func browserCommand(goos, u string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{u}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", u}
	default:
		return "xdg-open", []string{u}
	}
}

func openBrowser(u string) {
	name, args := browserCommand(runtime.GOOS, u)
	if err := startCommand(name, args...); err != nil {
		fmt.Fprintf(os.Stderr, "login: could not open a browser (%v); open the URL yourself\n", err)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
