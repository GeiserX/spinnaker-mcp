package tools_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/geiserx/spinnaker-mcp/internal/tools"
)

// gateError returns a gate mock that always returns 500.
func gateError(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`internal error`))
	}
}

func TestCancelExecution_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewCancelExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
	}))
	assertToolError(t, result, err)
}

func TestCancelExecution_WithReason(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"cancelled":true}`))
	})
	_, handler := tools.NewCancelExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
		"reason":       "bad build",
	}))
	assertNoError(t, result, err)
}

func TestResumeExecution_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewResumeExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
	}))
	assertToolError(t, result, err)
}

func TestResumeExecution_WithBody(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"resumed":true}`))
	})
	_, handler := tools.NewResumeExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
	}))
	assertNoError(t, result, err)
}

func TestGetExecution_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
	}))
	assertToolError(t, result, err)
}

func TestSearchExecutions_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewSearchExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestSearchExecutions_AllFilters(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	_, handler := tools.NewSearchExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"trigger_type": "webhook",
		"statuses":     "TERMINAL",
		"start_time":   "2024-01-01",
		"end_time":     "2024-12-31",
		"event_id":     "evt-1",
	}))
	assertNoError(t, result, err)
}

func TestGetTargetServerGroup_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":    "myapp",
		"account":        "prod",
		"cluster_name":   "myapp-prod",
		"cloud_provider": "aws",
		"scope":          "us-east-1",
		"target":         "newest",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingCluster(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
		"account":     "prod",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingCloudProvider(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"account":      "prod",
		"cluster_name": "myapp-prod",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingScope(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":    "myapp",
		"account":        "prod",
		"cluster_name":   "myapp-prod",
		"cloud_provider": "aws",
	}))
	assertToolError(t, result, err)
}

func TestGetTargetServerGroup_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetTargetServerGroup(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestGetScalingActivities_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":       "myapp",
		"account":           "prod",
		"cluster_name":      "myapp-prod",
		"server_group_name": "myapp-v001",
	}))
	assertToolError(t, result, err)
}

func TestGetScalingActivities_MissingCluster(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
		"account":     "prod",
	}))
	assertToolError(t, result, err)
}

func TestGetScalingActivities_MissingAccount(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewGetScalingActivities(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestGetConsoleOutput_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetConsoleOutput(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"instance_id": "i-123",
	}))
	assertToolError(t, result, err)
}

func TestGetPipeline_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertToolError(t, result, err)
}

func TestListPipelines_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListPipelines(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestListExecutions_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListExecutions(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestListServerGroups_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListServerGroups(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestListLoadBalancers_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListLoadBalancers(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestListClusters_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListClusters(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestGetCluster_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetCluster(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":  "myapp",
		"account":      "prod",
		"cluster_name": "myapp-prod",
	}))
	assertToolError(t, result, err)
}

func TestGetFirewall_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetFirewall(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
		"region":  "us-east-1",
		"name":    "sg-web",
	}))
	assertToolError(t, result, err)
}

func TestGetInstance_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetInstance(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":     "prod",
		"region":      "us-east-1",
		"instance_id": "i-123",
	}))
	assertToolError(t, result, err)
}

func TestGetImageTags_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetImageTags(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account":    "prod",
		"repository": "myrepo",
	}))
	assertToolError(t, result, err)
}

func TestFindImages_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewFindImages(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"provider": "aws",
		"query":    "myimage",
	}))
	assertToolError(t, result, err)
}

func TestDeletePipeline_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewDeletePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertToolError(t, result, err)
}

func TestDeleteStrategy_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewDeleteStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"strategy_name": "canary",
	}))
	assertToolError(t, result, err)
}

func TestGetPipelineHistory_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetPipelineHistory(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_config_id": "abc-123",
	}))
	assertToolError(t, result, err)
}

func TestListStrategies_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewListStrategies(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application": "myapp",
	}))
	assertToolError(t, result, err)
}

func TestGetAccount_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewGetAccount(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"account": "prod",
	}))
	assertToolError(t, result, err)
}

func TestEvaluateExpression_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewEvaluateExpression(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
		"expression":   "${trigger.type}",
	}))
	assertToolError(t, result, err)
}

func TestPauseExecution_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewPauseExecution(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
	}))
	assertToolError(t, result, err)
}

func TestRestartStage_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewRestartStage(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"execution_id": "exec-1",
		"stage_id":     "stage-1",
	}))
	assertToolError(t, result, err)
}

func TestTriggerPipeline_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewTriggerPipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"application":   "myapp",
		"pipeline_name": "deploy",
	}))
	assertToolError(t, result, err)
}

func TestSavePipeline_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewSavePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_json": `{"name":"test","application":"myapp","stages":[]}`,
	}))
	assertToolError(t, result, err)
}

func TestUpdatePipeline_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewUpdatePipeline(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pipeline_id":   "pipe-1",
		"pipeline_json": `{"name":"test","application":"myapp","id":"pipe-1","stages":[]}`,
	}))
	assertToolError(t, result, err)
}

func TestSaveStrategy_GateError(t *testing.T) {
	gate := newGate(t, gateError(t))
	_, handler := tools.NewSaveStrategy(gate)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"strategy_json": `{"name":"canary","application":"myapp"}`,
	}))
	assertToolError(t, result, err)
}

func TestListSubnets_MissingProvider(t *testing.T) {
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {})
	_, handler := tools.NewListSubnets(gate)
	result, err := handler(context.Background(), makeRequest(nil))
	assertToolError(t, result, err)
}
