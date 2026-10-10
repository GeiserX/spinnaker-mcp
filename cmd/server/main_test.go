package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
	"github.com/geiserx/spinnaker-mcp/internal/toolsets"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/healthz", nil)
	w := httptest.NewRecorder()

	healthzHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var data map[string]string
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %q, want %q", data["status"], "ok")
	}
}

func TestReadyzHandler_GateUp(t *testing.T) {
	gateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer gateSrv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: gateSrv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	handler := readyzHandler(gate)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/readyz", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data["status"] != "ready" {
		t.Errorf("status = %q, want %q", data["status"], "ready")
	}
	if data["gate_reachable"] != true {
		t.Errorf("gate_reachable = %v, want true", data["gate_reachable"])
	}
}

func TestReadyzHandler_GateDown(t *testing.T) {
	// Use a closed server so connection is refused
	gateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	gateURL := gateSrv.URL
	gateSrv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: gateURL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	handler := readyzHandler(gate)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/readyz", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 503 {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data["status"] != "unavailable" {
		t.Errorf("status = %q, want %q", data["status"], "unavailable")
	}
	if data["gate_reachable"] != false {
		t.Errorf("gate_reachable = %v, want false", data["gate_reachable"])
	}
}

func TestReadyzHandler_Gate500(t *testing.T) {
	gateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer gateSrv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: gateSrv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	handler := readyzHandler(gate)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/readyz", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 503 {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}

func TestPrintHelp_DoesNotPanic(t *testing.T) {
	// printHelp just writes to stdout; verify it doesn't panic
	printHelp()
}

func TestToolsetsLabel(t *testing.T) {
	if got := toolsetsLabel(""); got != "all" {
		t.Errorf("toolsetsLabel(\"\") = %q, want %q", got, "all")
	}
	if got := toolsetsLabel("pipelines,executions"); got != "pipelines,executions" {
		t.Errorf("toolsetsLabel(\"pipelines,executions\") = %q, want %q", got, "pipelines,executions")
	}
}

// An MCP client that leaves GATE_TOKEN blank may still pass the literal "${GATE_TOKEN}".
// It must not reach Gate as a bearer token.
func TestPlaceholderToken_SendsNoAuthorization(t *testing.T) {
	var auth string
	gateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer gateSrv.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GATE_URL", gateSrv.URL)
	t.Setenv("GATE_TOKEN", "${GATE_TOKEN}")
	t.Setenv("GATE_USER", "")
	t.Setenv("GATE_COOKIE", "")
	t.Setenv("GATE_COOKIE_FILE", "")

	gate, err := newGate(config.LoadGateConfig())
	if err != nil {
		t.Fatalf("newGate: %v", err)
	}
	if err := gate.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if auth != "" {
		t.Errorf("Authorization = %q, want none", auth)
	}
}

// The same for TOOLSETS: "${TOOLSETS}" means the default, all tools, not an unknown group.
func TestPlaceholderToolsets_RegistersAllTools(t *testing.T) {
	gate, err := client.NewGate(client.GateOptions{BaseURL: "http://localhost:8084"})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	all := toolsets.BuildTools(gate)
	want, _ := toolsets.Resolve("", all)

	// The literal value is what made the server exit; prove it still would.
	if _, err := toolsets.Resolve("${TOOLSETS}", all); err == nil {
		t.Fatal("Resolve accepted the literal placeholder; this test no longer covers the exit")
	}

	t.Setenv("TOOLSETS", "${TOOLSETS}")
	got, err := toolsets.Resolve(toolsetsSpec(""), all)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("registered %d tools, want all %d", len(got), len(want))
	}

	t.Setenv("TOOLSETS", "readonly")
	if got := toolsetsSpec(""); got != "readonly" {
		t.Errorf("toolsetsSpec = %q, want readonly", got)
	}
	if got := toolsetsSpec("pipelines"); got != "pipelines" {
		t.Errorf("toolsetsSpec with flag = %q, want pipelines", got)
	}
}
