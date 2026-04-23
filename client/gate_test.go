package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestGate(t *testing.T, url string) *GateClient {
	t.Helper()
	g, err := NewGate(GateOptions{BaseURL: url})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	return g
}

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

	gate, err := NewGate(GateOptions{BaseURL: srv.URL, Token: "test-token"})
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

	gate, err := NewGate(GateOptions{BaseURL: srv.URL, User: "admin", Pass: "secret"})
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.GetApplication(context.Background(), "nonexistent")
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.CancelExecution(context.Background(), "exec-123", "testing")
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

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListExecutions(context.Background(), "myapp", 10, "RUNNING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestNewGate_InvalidURL(t *testing.T) {
	_, err := NewGate(GateOptions{BaseURL: "://bad-url"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNewGate_InvalidScheme(t *testing.T) {
	_, err := NewGate(GateOptions{BaseURL: "ftp://example.com"})
	if err == nil {
		t.Fatal("expected error for non-http scheme")
	}
}

func TestNewGate_InvalidCertPath(t *testing.T) {
	_, err := NewGate(GateOptions{
		BaseURL:  "http://localhost:8084",
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.PauseExecution(context.Background(), "exec-123")
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.ResumeExecution(context.Background(), "exec-123")
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.CancelExecution(context.Background(), "exec-123", "")
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

	gate := newTestGate(t, srv.URL)
	_, err := gate.ListExecutions(context.Background(), "myapp", 0, "")
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

	gate := newTestGate(t, srv.URL)
	resp, err := gate.TriggerPipeline(context.Background(), "myapp", "deploy", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"ref":"/pipelines/abc123"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_RedirectReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirected", http.StatusFound)
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.ListApplications(context.Background())
	if err == nil {
		t.Fatal("expected error for redirect response")
	}
	if !strings.Contains(err.Error(), "redirect") {
		t.Errorf("expected redirect error, got: %v", err)
	}
}

func TestGateClient_ErrorBodyTruncated(t *testing.T) {
	bigBody := strings.Repeat("x", 1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(bigBody))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.ListApplications(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "(truncated)") {
		t.Errorf("expected truncated error, got: %v", err)
	}
	if len(err.Error()) > 600 {
		t.Errorf("error too long (%d chars), truncation may have failed", len(err.Error()))
	}
}

func TestGateClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := gate.ListApplications(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestGateClient_BasePathPreserved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/applications" {
			t.Errorf("expected /api/v1/applications, got %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := NewGate(GateOptions{BaseURL: srv.URL + "/api/v1"})
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

func TestGateClient_BasePathTrailingSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/applications" {
			t.Errorf("expected /api/v1/applications, got %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := NewGate(GateOptions{BaseURL: srv.URL + "/api/v1/"})
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

func TestNewGate_InsecureNoWarnOnHTTP(t *testing.T) {
	_, err := NewGate(GateOptions{BaseURL: "http://localhost:8084", Insecure: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_SavePipeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"name":"my-pipeline"`) {
			t.Errorf("expected pipeline name in body, got %s", string(body))
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.SavePipeline(context.Background(), map[string]any{"name": "my-pipeline", "application": "myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_UpdatePipeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines/pipe-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"name":"updated"`) {
			t.Errorf("expected updated name in body, got %s", string(body))
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.UpdatePipeline(context.Background(), "pipe-123", map[string]any{"name": "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_DeletePipeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines/myapp/deploy" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.DeletePipeline(context.Background(), "myapp", "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetPipelineHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pipelineConfigs/config-abc/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("expected limit=5, got %s", r.URL.Query().Get("limit"))
		}
		w.Write([]byte(`[{"id":"v1"},{"id":"v2"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetPipelineHistory(context.Background(), "config-abc", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"id":"v1"},{"id":"v2"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetPipelineHistoryNoLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "" {
			t.Errorf("expected no limit param, got %q", r.URL.Query().Get("limit"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.GetPipelineHistory(context.Background(), "config-abc", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_RestartStage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines/exec-123/stages/stage-1/restart" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected non-empty body")
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.RestartStage(context.Background(), "exec-123", "stage-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_SearchExecutions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/executions/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("pipelineName") != "deploy" {
			t.Errorf("expected pipelineName=deploy, got %s", r.URL.Query().Get("pipelineName"))
		}
		if r.URL.Query().Get("statuses") != "RUNNING" {
			t.Errorf("expected statuses=RUNNING, got %s", r.URL.Query().Get("statuses"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.SearchExecutions(context.Background(), "myapp", map[string]string{
		"pipelineName": "deploy",
		"statuses":     "RUNNING",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_SearchExecutionsEmptyValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("statuses") != "" {
			t.Errorf("expected no statuses param, got %q", r.URL.Query().Get("statuses"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.SearchExecutions(context.Background(), "myapp", map[string]string{"statuses": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_EvaluateExpression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pipelines/exec-123/evaluateExpression" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"expression":"${trigger.buildNumber}"`) {
			t.Errorf("expected expression in body, got %s", string(body))
		}
		w.Write([]byte(`{"result":"42"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.EvaluateExpression(context.Background(), "exec-123", "${trigger.buildNumber}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"result":"42"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListStrategies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/strategyConfigs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"highlander"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListStrategies(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"highlander"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_SaveStrategy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/strategies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"name":"highlander"`) {
			t.Errorf("expected strategy name in body, got %s", string(body))
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.SaveStrategy(context.Background(), map[string]any{"name": "highlander", "application": "myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_DeleteStrategy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/strategies/myapp/highlander" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.DeleteStrategy(context.Background(), "myapp", "highlander")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListClusters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/clusters" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"prod":["myapp-prod"]}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListClusters(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"prod":["myapp-prod"]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetCluster(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/clusters/prod/myapp-prod" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"name":"myapp-prod"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetCluster(context.Background(), "myapp", "prod", "myapp-prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"myapp-prod"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetScalingActivities(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/clusters/prod/myapp-prod/serverGroups/myapp-v001/scalingActivities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("provider") != "aws" {
			t.Errorf("expected provider=aws, got %s", r.URL.Query().Get("provider"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetScalingActivities(context.Background(), "myapp", "prod", "myapp-prod", "myapp-v001", "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetScalingActivitiesNoProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("provider") != "" {
			t.Errorf("expected no provider param, got %q", r.URL.Query().Get("provider"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.GetScalingActivities(context.Background(), "myapp", "prod", "myapp-prod", "myapp-v001", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_GetTargetServerGroup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/myapp/clusters/prod/myapp-prod/aws/us-east-1/serverGroups/target/newest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"name":"myapp-v002"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetTargetServerGroup(context.Background(), "myapp", "prod", "myapp-prod", "aws", "us-east-1", "newest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"myapp-v002"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListFirewalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/securityGroups" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"sg-default"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListFirewalls(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"sg-default"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetFirewall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/securityGroups/prod/us-east-1/sg-default" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"name":"sg-default"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetFirewall(context.Background(), "prod", "us-east-1", "sg-default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"sg-default"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetInstance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instances/prod/us-east-1/i-abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"instanceId":"i-abc123"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetInstance(context.Background(), "prod", "us-east-1", "i-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"instanceId":"i-abc123"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetConsoleOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instances/prod/us-east-1/i-abc123/console" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("provider") != "aws" {
			t.Errorf("expected provider=aws, got %s", r.URL.Query().Get("provider"))
		}
		w.Write([]byte(`{"output":"boot log"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetConsoleOutput(context.Background(), "prod", "us-east-1", "i-abc123", "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"output":"boot log"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetConsoleOutputNoProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("provider") != "" {
			t.Errorf("expected no provider param, got %q", r.URL.Query().Get("provider"))
		}
		w.Write([]byte(`{"output":"boot log"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.GetConsoleOutput(context.Background(), "prod", "us-east-1", "i-abc123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_FindImages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/find" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("provider") != "aws" {
			t.Errorf("expected provider=aws, got %s", r.URL.Query().Get("provider"))
		}
		if r.URL.Query().Get("q") != "my-image" {
			t.Errorf("expected q=my-image, got %s", r.URL.Query().Get("q"))
		}
		w.Write([]byte(`[{"imageName":"my-image-v1"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.FindImages(context.Background(), map[string]string{
		"provider": "aws",
		"q":        "my-image",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"imageName":"my-image-v1"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_FindImagesEmptyValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("provider") != "" {
			t.Errorf("expected no provider param, got %q", r.URL.Query().Get("provider"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	_, err := gate.FindImages(context.Background(), map[string]string{"provider": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_GetImageTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("account") != "my-registry" {
			t.Errorf("expected account=my-registry, got %s", r.URL.Query().Get("account"))
		}
		if r.URL.Query().Get("repository") != "myapp" {
			t.Errorf("expected repository=myapp, got %s", r.URL.Query().Get("repository"))
		}
		w.Write([]byte(`["v1.0","v1.1"]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetImageTags(context.Background(), "my-registry", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `["v1.0","v1.1"]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListNetworks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/networks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"aws":[{"id":"vpc-123"}]}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListNetworks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"aws":[{"id":"vpc-123"}]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListSubnets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subnets/aws" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"id":"subnet-123"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListSubnets(context.Background(), "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"id":"subnet-123"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_ListAccounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/credentials" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"name":"prod","type":"aws"}]`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.ListAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `[{"name":"prod","type":"aws"}]` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestGateClient_GetAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/credentials/prod" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"name":"prod","type":"aws"}`))
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	resp, err := gate.GetAccount(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"name":"prod","type":"aws"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}
