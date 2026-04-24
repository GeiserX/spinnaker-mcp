package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
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
