package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGate_BearerAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer auth, got %q", auth)
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`[{"name":"myapp"}]`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "test-token", "", "", "", "", false)
	resp, err := gate.ListApplications()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"myapp"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestNewGate_BasicAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			t.Errorf("expected basic auth admin:secret, got %q:%q ok=%v", user, pass, ok)
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"name":"myapp"}`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "admin", "secret", "", "", false)
	resp, err := gate.GetApplication("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"myapp"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestNewGate_NoAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no auth header")
		}
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	resp, err := gate.ListApplications()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Not Found"}`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	_, err := gate.GetApplication("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestGateClient_ListPipelines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/pipelineConfigs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"deploy"}]`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	resp, err := gate.ListPipelines("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"deploy"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_TriggerPipeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected non-empty body")
		}
		w.Write([]byte(`{"ref":"/pipelines/abc123"}`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	resp, err := gate.TriggerPipeline("myapp", "deploy", map[string]any{"tag": "v1.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"ref":"/pipelines/abc123"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_CancelExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Query().Get("reason") != "testing" {
			t.Errorf("expected reason=testing, got %s", r.URL.Query().Get("reason"))
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	_, err := gate.CancelExecution("exec-123", "testing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_ListExecutions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("statuses") != "RUNNING" {
			t.Errorf("expected statuses=RUNNING, got %s", r.URL.Query().Get("statuses"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := NewGate(srv.URL, "", "", "", "", "", false)
	resp, err := gate.ListExecutions("myapp", 10, "RUNNING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}
