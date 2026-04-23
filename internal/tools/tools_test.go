package tools_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

func newGate(t *testing.T, handler http.HandlerFunc) *client.GateClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	g, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	return g
}

func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func assertNoError(t *testing.T, result *mcp.CallToolResult, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("handler returned Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned tool error: %v", result.Content)
	}
}

func assertToolError(t *testing.T, result *mcp.CallToolResult, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("handler returned Go error (expected tool error): %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error, got success")
	}
}

// --- ListApplications ---

func TestListApplications_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"app1"}]`))
	})
	_, handler := tools.NewListApplications(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertNoError(t, result, err)
}

func TestListApplications_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"fail"}`))
	})
	_, handler := tools.NewListApplications(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetApplication ---

func TestGetApplication_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"myapp"}`))
	})
	_, handler := tools.NewGetApplication(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertNoError(t, result, err)
}

func TestGetApplication_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetApplication(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

func TestGetApplication_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	_, handler := tools.NewGetApplication(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "x"}))
	assertToolError(t, result, err)
}

// --- ListPipelines ---

func TestListPipelines_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"deploy"}]`))
	})
	_, handler := tools.NewListPipelines(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertNoError(t, result, err)
}

func TestListPipelines_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListPipelines(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetPipeline ---

func TestGetPipeline_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"deploy","stages":[]}`))
	})
	_, handler := tools.NewGetPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertNoError(t, result, err)
}

func TestGetPipeline_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"pipeline_name": "deploy"}))
	assertToolError(t, result, err)
}

func TestGetPipeline_MissingPipelineName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertToolError(t, result, err)
}

// --- TriggerPipeline ---

func TestTriggerPipeline_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ref":"/pipelines/abc"}`))
	})
	_, handler := tools.NewTriggerPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
		"parameters":    `{"tag":"v1.0"}`,
	}))
	assertNoError(t, result, err)
}

func TestTriggerPipeline_NoParams(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ref":"/pipelines/abc"}`))
	})
	_, handler := tools.NewTriggerPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertNoError(t, result, err)
}

func TestTriggerPipeline_InvalidJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewTriggerPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
		"parameters":    `{bad json}`,
	}))
	assertToolError(t, result, err)
}

func TestTriggerPipeline_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewTriggerPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"pipeline_name": "deploy"}))
	assertToolError(t, result, err)
}

// --- GetExecution ---

func TestGetExecution_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"e1","status":"SUCCEEDED"}`))
	})
	_, handler := tools.NewGetExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"execution_id": "e1"}))
	assertNoError(t, result, err)
}

func TestGetExecution_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetExecution(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- ListExecutions ---

func TestListExecutions_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewListExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
		"limit":       json.Number("10"),
		"statuses":    "RUNNING",
	}))
	assertNoError(t, result, err)
}

func TestListExecutions_Defaults(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewListExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertNoError(t, result, err)
}

func TestListExecutions_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListExecutions(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- CancelExecution ---

func TestCancelExecution_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewCancelExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"reason":       "testing",
	}))
	assertNoError(t, result, err)
}

func TestCancelExecution_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewCancelExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"execution_id": "e1"}))
	assertNoError(t, result, err)
}

func TestCancelExecution_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewCancelExecution(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- PauseExecution ---

func TestPauseExecution_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewPauseExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"execution_id": "e1"}))
	assertNoError(t, result, err)
}

func TestPauseExecution_WithBody(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"paused"}`))
	})
	_, handler := tools.NewPauseExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"execution_id": "e1"}))
	assertNoError(t, result, err)
}

func TestPauseExecution_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewPauseExecution(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- ResumeExecution ---

func TestResumeExecution_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewResumeExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"execution_id": "e1"}))
	assertNoError(t, result, err)
}

func TestResumeExecution_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewResumeExecution(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- ListServerGroups ---

func TestListServerGroups_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"sg-001"}]`))
	})
	_, handler := tools.NewListServerGroups(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertNoError(t, result, err)
}

func TestListServerGroups_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListServerGroups(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- ListLoadBalancers ---

func TestListLoadBalancers_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"lb-001"}]`))
	})
	_, handler := tools.NewListLoadBalancers(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"application": "myapp"}))
	assertNoError(t, result, err)
}

func TestListLoadBalancers_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListLoadBalancers(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetTask ---

func TestGetTask_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"t1","status":"SUCCEEDED"}`))
	})
	_, handler := tools.NewGetTask(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"task_id": "t1"}))
	assertNoError(t, result, err)
}

func TestGetTask_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetTask(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

func TestGetTask_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	_, handler := tools.NewGetTask(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{"task_id": "t1"}))
	assertToolError(t, result, err)
}
