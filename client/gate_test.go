package client

import (
	"context"
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

	gate, err := NewGate(srv.URL, "test-token", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListApplications(context.Background())
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

	gate, err := NewGate(srv.URL, "", "admin", "secret", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.GetApplication(context.Background(), "myapp")
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListApplications(context.Background())
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.GetApplication(context.Background(), "nonexistent")
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListPipelines(context.Background(), "myapp")
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.TriggerPipeline(context.Background(), "myapp", "deploy", map[string]any{"tag": "v1.0"})
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.CancelExecution(context.Background(), "exec-123", "testing")
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

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListExecutions(context.Background(), "myapp", 10, "RUNNING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestNewGate_InvalidURL(t *testing.T) {
	_, err := NewGate("://bad-url", "", "", "", "", "", false)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNewGate_InvalidScheme(t *testing.T) {
	_, err := NewGate("ftp://example.com", "", "", "", "", "", false)
	if err == nil {
		t.Fatal("expected error for non-http scheme")
	}
}

func TestNewGate_InvalidCertPath(t *testing.T) {
	_, err := NewGate("http://localhost:8084", "", "", "", "/nonexistent/cert.pem", "/nonexistent/key.pem", false)
	if err == nil {
		t.Fatal("expected error for invalid cert path")
	}
}

func TestGateClient_PutMethod(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.PauseExecution(context.Background(), "exec-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_GetPipelineConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/pipelineConfigs/deploy" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"name":"deploy","stages":[]}`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.GetPipelineConfig(context.Background(), "myapp", "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"deploy","stages":[]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pipelines/exec-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"id":"exec-123","status":"SUCCEEDED"}`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.GetExecution(context.Background(), "exec-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"id":"exec-123","status":"SUCCEEDED"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ResumeExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines/exec-123/resume" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.ResumeExecution(context.Background(), "exec-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_ListServerGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/serverGroups" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"myapp-v001"}]`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListServerGroups(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"myapp-v001"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListLoadBalancers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/loadBalancers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"myapp-lb"}]`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.ListLoadBalancers(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"myapp-lb"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetTask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/task-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"id":"task-456","status":"SUCCEEDED"}`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.GetTask(context.Background(), "task-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"id":"task-456","status":"SUCCEEDED"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_CancelExecutionNoReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("reason") != "" {
			t.Errorf("expected no reason param, got %q", r.URL.Query().Get("reason"))
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.CancelExecution(context.Background(), "exec-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_ListExecutionsDefaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "" {
			t.Errorf("expected no limit param, got %q", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("statuses") != "" {
			t.Errorf("expected no statuses param, got %q", r.URL.Query().Get("statuses"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = gate.ListExecutions(context.Background(), "myapp", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_TriggerPipelineNoParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) == "" {
			t.Error("expected non-empty body even without params")
		}
		w.Write([]byte(`{"ref":"/pipelines/abc123"}`))
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := gate.TriggerPipeline(context.Background(), "myapp", "deploy", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"ref":"/pipelines/abc123"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_NoRedirectFollowing(t *testing.T) {
	var redirectFollowed bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirected" {
			redirectFollowed = true
			w.Write([]byte(`"redirected"`))
			return
		}
		http.Redirect(w, r, "/redirected", http.StatusFound)
	}))
	defer srv.Close()

	gate, err := NewGate(srv.URL, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The response is a 302 with a body — client should NOT follow it
	resp, _ := gate.ListApplications(context.Background())
	if redirectFollowed {
		t.Error("client followed redirect — CheckRedirect should prevent this")
	}
	// With ErrUseLastResponse, the 302 body is returned (status < 400)
	if string(resp) == `"redirected"` {
		t.Error("got redirected response — client should not follow redirects")
	}
}
