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

// --- SavePipeline ---

func TestSavePipeline_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"name":"deploy","application":"myapp"}`,
	}))
	assertNoError(t, result, err)
}

func TestSavePipeline_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"name":"deploy","application":"myapp"}`,
	}))
	assertNoError(t, result, err)
}

func TestSavePipeline_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"name":"deploy"}`,
	}))
	assertToolError(t, result, err)
}

func TestSavePipeline_MissingName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"application":"myapp"}`,
	}))
	assertToolError(t, result, err)
}

func TestSavePipeline_OversizedJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSavePipeline(gate)
	bigJSON := `{"name":"x","application":"a","data":"` + string(make([]byte, 1<<20)) + `"}`
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": bigJSON,
	}))
	assertToolError(t, result, err)
}

func TestSavePipeline_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

func TestSavePipeline_InvalidJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{bad json}`,
	}))
	assertToolError(t, result, err)
}

// --- UpdatePipeline ---

func TestUpdatePipeline_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_id":   "pid-123",
		"pipeline_json": `{"name":"deploy","stages":[]}`,
	}))
	assertNoError(t, result, err)
}

func TestUpdatePipeline_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_id":   "pid-123",
		"pipeline_json": `{"name":"deploy"}`,
	}))
	assertNoError(t, result, err)
}

func TestUpdatePipeline_MissingPipelineID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"name":"deploy"}`,
	}))
	assertToolError(t, result, err)
}

func TestUpdatePipeline_MissingPipelineJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_id": "pid-123",
	}))
	assertToolError(t, result, err)
}

func TestUpdatePipeline_InvalidJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_id":   "pid-123",
		"pipeline_json": `{bad json}`,
	}))
	assertToolError(t, result, err)
}

// --- DeletePipeline ---

func TestDeletePipeline_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewDeletePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertNoError(t, result, err)
}

func TestDeletePipeline_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewDeletePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertNoError(t, result, err)
}

func TestDeletePipeline_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewDeletePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_name": "deploy",
	}))
	assertToolError(t, result, err)
}

func TestDeletePipeline_MissingPipelineName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewDeletePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

// --- GetPipelineHistory ---

func TestGetPipelineHistory_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"version":1}]`))
	})
	_, handler := tools.NewGetPipelineHistory(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_config_id": "cfg-1",
	}))
	assertNoError(t, result, err)
}

func TestGetPipelineHistory_WithLimit(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewGetPipelineHistory(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_config_id": "cfg-1",
		"limit":              json.Number("5"),
	}))
	assertNoError(t, result, err)
}

func TestGetPipelineHistory_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetPipelineHistory(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- RestartStage ---

func TestRestartStage_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewRestartStage(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"stage_id":     "s1",
	}))
	assertNoError(t, result, err)
}

func TestRestartStage_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewRestartStage(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"stage_id":     "s1",
	}))
	assertNoError(t, result, err)
}

func TestRestartStage_MissingExecutionID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewRestartStage(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"stage_id": "s1",
	}))
	assertToolError(t, result, err)
}

func TestRestartStage_MissingStageID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewRestartStage(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
	}))
	assertToolError(t, result, err)
}

// --- SearchExecutions ---

func TestSearchExecutions_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"e1"}]`))
	})
	_, handler := tools.NewSearchExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertNoError(t, result, err)
}

func TestSearchExecutions_WithFilters(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewSearchExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"trigger_type": "manual",
		"statuses":     "RUNNING",
	}))
	assertNoError(t, result, err)
}

func TestSearchExecutions_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSearchExecutions(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- EvaluateExpression ---

func TestEvaluateExpression_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result":"value"}`))
	})
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"expression":   "${trigger.type}",
	}))
	assertNoError(t, result, err)
}

func TestEvaluateExpression_MissingExecutionID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"expression": "${trigger.type}",
	}))
	assertToolError(t, result, err)
}

func TestEvaluateExpression_MissingExpression(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
	}))
	assertToolError(t, result, err)
}

func TestEvaluateExpression_UnsafeSpEL(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"expression":   "T(java.lang.Runtime).getRuntime().exec('whoami')",
	}))
	assertToolError(t, result, err)
}

func TestEvaluateExpression_OversizedExpression(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "e1",
		"expression":   string(make([]byte, 4097)),
	}))
	assertToolError(t, result, err)
}

// --- ListStrategies ---

func TestListStrategies_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"canary"}]`))
	})
	_, handler := tools.NewListStrategies(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertNoError(t, result, err)
}

func TestListStrategies_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListStrategies(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- SaveStrategy ---

func TestSaveStrategy_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewSaveStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"strategy_json": `{"name":"canary","application":"myapp"}`,
	}))
	assertNoError(t, result, err)
}

func TestSaveStrategy_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewSaveStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"strategy_json": `{"name":"canary"}`,
	}))
	assertNoError(t, result, err)
}

func TestSaveStrategy_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSaveStrategy(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

func TestSaveStrategy_InvalidJSON(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewSaveStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"strategy_json": `{bad json}`,
	}))
	assertToolError(t, result, err)
}

// --- DeleteStrategy ---

func TestDeleteStrategy_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	_, handler := tools.NewDeleteStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"strategy_name": "canary",
	}))
	assertNoError(t, result, err)
}

func TestDeleteStrategy_EmptyResponse(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, handler := tools.NewDeleteStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"strategy_name": "canary",
	}))
	assertNoError(t, result, err)
}

func TestDeleteStrategy_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewDeleteStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"strategy_name": "canary",
	}))
	assertToolError(t, result, err)
}

func TestDeleteStrategy_MissingStrategyName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewDeleteStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

// --- ListClusters ---

func TestListClusters_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"cluster-1"}]`))
	})
	_, handler := tools.NewListClusters(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertNoError(t, result, err)
}

func TestListClusters_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListClusters(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetCluster ---

func TestGetCluster_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"cluster-1","serverGroups":[]}`))
	})
	_, handler := tools.NewGetCluster(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"account":      "prod",
		"cluster_name": "cluster-1",
	}))
	assertNoError(t, result, err)
}

func TestGetCluster_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetCluster(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":      "prod",
		"cluster_name": "cluster-1",
	}))
	assertToolError(t, result, err)
}

func TestGetCluster_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetCluster(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"cluster_name": "cluster-1",
	}))
	assertToolError(t, result, err)
}

func TestGetCluster_MissingClusterName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetCluster(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
		"account":     "prod",
	}))
	assertToolError(t, result, err)
}

// --- GetScalingActivities ---

func TestGetScalingActivities_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"activity":"scale-up"}]`))
	})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":       "myapp",
		"account":           "prod",
		"cluster_name":      "cluster-1",
		"server_group_name": "sg-001",
	}))
	assertNoError(t, result, err)
}

func TestGetScalingActivities_WithProvider(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":       "myapp",
		"account":           "prod",
		"cluster_name":      "cluster-1",
		"server_group_name": "sg-001",
		"provider":          "aws",
	}))
	assertNoError(t, result, err)
}

func TestGetScalingActivities_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":           "prod",
		"cluster_name":      "cluster-1",
		"server_group_name": "sg-001",
	}))
	assertToolError(t, result, err)
}

func TestGetScalingActivities_MissingServerGroup(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"account":      "prod",
		"cluster_name": "cluster-1",
	}))
	assertToolError(t, result, err)
}

// --- GetTargetServerGroup ---

func TestGetTargetServerGroup_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"sg-newest"}`))
	})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":    "myapp",
		"account":        "prod",
		"cluster_name":   "cluster-1",
		"cloud_provider": "aws",
		"scope":          "us-east-1",
		"target":         "newest",
	}))
	assertNoError(t, result, err)
}

func TestGetTargetServerGroup_InvalidTarget(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":    "myapp",
		"account":        "prod",
		"cluster_name":   "cluster-1",
		"cloud_provider": "aws",
		"scope":          "us-east-1",
		"target":         "invalid",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingApplication(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":        "prod",
		"cluster_name":   "cluster-1",
		"cloud_provider": "aws",
		"scope":          "us-east-1",
		"target":         "newest",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingTarget(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":    "myapp",
		"account":        "prod",
		"cluster_name":   "cluster-1",
		"cloud_provider": "aws",
		"scope":          "us-east-1",
	}))
	assertToolError(t, result, err)
}

// --- ListFirewalls ---

func TestListFirewalls_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"sg-default"}]`))
	})
	_, handler := tools.NewListFirewalls(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertNoError(t, result, err)
}

func TestListFirewalls_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	_, handler := tools.NewListFirewalls(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetFirewall ---

func TestGetFirewall_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"sg-web","inboundRules":[]}`))
	})
	_, handler := tools.NewGetFirewall(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"region":  "us-east-1",
		"name":    "sg-web",
	}))
	assertNoError(t, result, err)
}

func TestGetFirewall_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetFirewall(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"region": "us-east-1",
		"name":   "sg-web",
	}))
	assertToolError(t, result, err)
}

func TestGetFirewall_MissingRegion(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetFirewall(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"name":    "sg-web",
	}))
	assertToolError(t, result, err)
}

func TestGetFirewall_MissingName(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetFirewall(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"region":  "us-east-1",
	}))
	assertToolError(t, result, err)
}

// --- GetInstance ---

func TestGetInstance_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"instanceId":"i-abc123"}`))
	})
	_, handler := tools.NewGetInstance(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"region":      "us-east-1",
		"instance_id": "i-abc123",
	}))
	assertNoError(t, result, err)
}

func TestGetInstance_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetInstance(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"region":      "us-east-1",
		"instance_id": "i-abc123",
	}))
	assertToolError(t, result, err)
}

func TestGetInstance_MissingRegion(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetInstance(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"instance_id": "i-abc123",
	}))
	assertToolError(t, result, err)
}

func TestGetInstance_MissingInstanceID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetInstance(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"region":  "us-east-1",
	}))
	assertToolError(t, result, err)
}

// --- GetConsoleOutput ---

func TestGetConsoleOutput_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"output":"boot log..."}`))
	})
	_, handler := tools.NewGetConsoleOutput(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"region":      "us-east-1",
		"instance_id": "i-abc123",
	}))
	assertNoError(t, result, err)
}

func TestGetConsoleOutput_WithProvider(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"output":"log"}`))
	})
	_, handler := tools.NewGetConsoleOutput(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"region":      "us-east-1",
		"instance_id": "i-abc123",
		"provider":    "aws",
	}))
	assertNoError(t, result, err)
}

func TestGetConsoleOutput_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetConsoleOutput(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"region":      "us-east-1",
		"instance_id": "i-abc123",
	}))
	assertToolError(t, result, err)
}

func TestGetConsoleOutput_MissingInstanceID(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetConsoleOutput(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"region":  "us-east-1",
	}))
	assertToolError(t, result, err)
}

// --- FindImages ---

func TestFindImages_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"imageName":"ami-123"}]`))
	})
	_, handler := tools.NewFindImages(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"provider": "aws",
	}))
	assertNoError(t, result, err)
}

func TestFindImages_WithOptionals(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewFindImages(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"provider": "aws",
		"query":    "base-image",
		"region":   "us-east-1",
		"account":  "prod",
	}))
	assertNoError(t, result, err)
}

func TestFindImages_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewFindImages(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetImageTags ---

func TestGetImageTags_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`["latest","v1.0","v2.0"]`))
	})
	_, handler := tools.NewGetImageTags(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":    "dockerhub",
		"repository": "library/nginx",
	}))
	assertNoError(t, result, err)
}

func TestGetImageTags_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetImageTags(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"repository": "library/nginx",
	}))
	assertToolError(t, result, err)
}

func TestGetImageTags_MissingRepository(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetImageTags(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "dockerhub",
	}))
	assertToolError(t, result, err)
}

// --- ListNetworks ---

func TestListNetworks_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"vpc-1"}]`))
	})
	_, handler := tools.NewListNetworks(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertNoError(t, result, err)
}

func TestListNetworks_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	_, handler := tools.NewListNetworks(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- ListSubnets ---

func TestListSubnets_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"subnet-1"}]`))
	})
	_, handler := tools.NewListSubnets(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"cloud_provider": "aws",
	}))
	assertNoError(t, result, err)
}

func TestListSubnets_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewListSubnets(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

func TestListSubnets_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	_, handler := tools.NewListSubnets(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"cloud_provider": "aws",
	}))
	assertToolError(t, result, err)
}

// --- ListAccounts ---

func TestListAccounts_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"prod"}]`))
	})
	_, handler := tools.NewListAccounts(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertNoError(t, result, err)
}

func TestListAccounts_GateError(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	_, handler := tools.NewListAccounts(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}

// --- GetAccount ---

func TestGetAccount_Happy(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"prod","type":"aws"}`))
	})
	_, handler := tools.NewGetAccount(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
	}))
	assertNoError(t, result, err)
}

func TestGetAccount_MissingParam(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("gate should not be called")
	})
	_, handler := tools.NewGetAccount(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}
