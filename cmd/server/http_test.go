package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/internal/toolsets"
)

func newTestHTTP(t *testing.T, c httpConfig) *httptest.Server {
	t.Helper()
	gate, err := client.NewGate(client.GateOptions{BaseURL: "http://127.0.0.1:1"})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	selected, _ := toolsets.Resolve("", toolsets.BuildTools(gate))
	h, err := newHTTPHandler(newMCPServer(gate, selected), gate, c)
	if err != nil {
		t.Fatalf("newHTTPHandler: %v", err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

const discoverBody = `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{` + modernMeta + `}}`

func postMCP(t *testing.T, srv *httptest.Server, body string, hdr map[string]string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest("POST", srv.URL+"/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

var modernHeaders = map[string]string{"MCP-Protocol-Version": "2026-07-28", "Mcp-Method": "server/discover"}

func with(h map[string]string, kv ...string) map[string]string {
	out := map[string]string{}
	for k, v := range h {
		out[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		out[kv[i]] = kv[i+1]
	}
	return out
}

func TestHTTP_ModernDiscover(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	resp, body := postMCP(t, srv, discoverBody, modernHeaders)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"2026-07-28"`) {
		t.Errorf("server/discover over HTTP does not offer 2026-07-28: %s", body)
	}
	if resp.Header.Get("Mcp-Session-Id") != "" {
		t.Error("stateless server minted a session id")
	}
}

func TestHTTP_LegacyInitializeThenToolsList(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	resp, body := postMCP(t, srv, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("legacy initialize: %d %s", resp.StatusCode, body)
	}
	// Stateless: no session to remember, so any replica can serve the next request.
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		t.Errorf("legacy initialize minted session %q on a stateless server", sid)
	}
	hdr := map[string]string{"MCP-Protocol-Version": "2025-06-18"}
	resp, body = postMCP(t, srv, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, hdr)
	if resp.StatusCode != 200 || !strings.Contains(body, `"list_applications"`) {
		t.Fatalf("legacy tools/list: %d %s", resp.StatusCode, body)
	}
}

// The spec requires a 403 for an invalid Origin; browsers on another site must not be
// able to drive the server, even when it only listens on loopback.
func TestHTTP_CrossOriginRejected(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{allowedOrigins: []string{"https://console.example.com"}})

	resp, _ := postMCP(t, srv, discoverBody, with(modernHeaders, "Origin", "https://evil.example", "Sec-Fetch-Site", "cross-site"))
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-site POST status = %d, want 403", resp.StatusCode)
	}
	resp, _ = postMCP(t, srv, discoverBody, with(modernHeaders, "Origin", "https://evil.example"))
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign Origin POST status = %d, want 403", resp.StatusCode)
	}
	// Positive controls: no Origin (every non-browser client) and an allow-listed origin.
	if resp, body := postMCP(t, srv, discoverBody, modernHeaders); resp.StatusCode != 200 {
		t.Errorf("POST without Origin = %d %s, want 200", resp.StatusCode, body)
	}
	if resp, body := postMCP(t, srv, discoverBody, with(modernHeaders, "Origin", "https://console.example.com", "Sec-Fetch-Site", "cross-site")); resp.StatusCode != 200 {
		t.Errorf("trusted Origin POST = %d %s, want 200", resp.StatusCode, body)
	}
}

func TestHTTP_BodyCapped(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	pad := strings.Repeat("x", maxRequestBody+1024)
	big := `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{` + modernMeta + `,"pad":"` + pad + `"}}`
	resp, body := postMCP(t, srv, big, modernHeaders)
	if resp.StatusCode == 200 && strings.Contains(body, "supportedVersions") {
		t.Errorf("a %d-byte body was accepted", len(big))
	}
	// Positive control: the same request just under the cap is served.
	small := `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{` + modernMeta + `,"pad":"` + strings.Repeat("x", 1<<20) + `"}}`
	if resp, body := postMCP(t, srv, small, modernHeaders); resp.StatusCode != 200 {
		t.Errorf("1 MiB body = %d %s, want 200", resp.StatusCode, body)
	}
}

func TestHTTP_BearerRequired(t *testing.T) {
	const tok = "s3cr3t-token-for-tests"
	srv := newTestHTTP(t, httpConfig{token: tok})
	cases := map[string]string{
		"missing":        "",
		"no Bearer word": tok,
		"wrong token":    "Bearer nope",
		"empty bearer":   "Bearer ",
		"basic scheme":   "Basic " + tok,
	}
	for name, auth := range cases {
		hdr := modernHeaders
		if auth != "" {
			hdr = with(modernHeaders, "Authorization", auth)
		}
		if resp, _ := postMCP(t, srv, discoverBody, hdr); resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", name, resp.StatusCode)
		}
	}
	if resp, body := postMCP(t, srv, discoverBody, with(modernHeaders, "Authorization", "Bearer "+tok)); resp.StatusCode != 200 {
		t.Errorf("valid token: %d %s, want 200", resp.StatusCode, body)
	}
	// Probes stay open for Kubernetes and Docker health checks.
	resp, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Errorf("/healthz without token: %v %v", resp, err)
	}
}

func TestLoadHTTPConfig(t *testing.T) {
	for _, tc := range []struct {
		bind, token string
		wantErr     bool
	}{
		{"", "", false},
		{"127.0.0.1", "", false},
		{"::1", "", false},
		{"localhost", "", false},
		{"0.0.0.0", "", true},
		{"10.1.2.3", "", true},
		{"0.0.0.0", "${MCP_AUTH_TOKEN}", true},
		{"0.0.0.0", "tok", false},
	} {
		t.Setenv("MCP_PORT", "")
		t.Setenv("MCP_ALLOWED_ORIGINS", "")
		t.Setenv("MCP_BIND_ADDR", tc.bind)
		t.Setenv("MCP_AUTH_TOKEN", tc.token)
		_, err := loadHTTPConfig()
		if (err != nil) != tc.wantErr {
			t.Errorf("bind %q token %q: err = %v, wantErr %v", tc.bind, tc.token, err, tc.wantErr)
		}
	}
	t.Setenv("MCP_BIND_ADDR", "")
	t.Setenv("MCP_PORT", "70000")
	if _, err := loadHTTPConfig(); err == nil {
		t.Error("MCP_PORT=70000 accepted")
	}
}

func TestHTTP_ReadyzHidesGateURL(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	resp, err := srv.Client().Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var data map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&data)
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if !slices.Equal(keys, []string{"gate_reachable", "status"}) {
		t.Errorf("/readyz keys = %v", keys)
	}
}
