package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const executionOutputSchema = `{
  "type": "object",
  "description": "By default the compact execution summary; with summary=false the full execution object from Gate, which has id, status and stages too.",
  "properties": {
    "id": {"type": "string"},
    "status": {"type": "string"},
    "stages": {"type": "array", "description": "Per stage summary, or Gate's full stage objects with summary=false."},
    "running": {"type": "array", "description": "Stages running now (summary only)."},
    "awaitingJudgment": {"type": "array", "description": "Manual judgment stages waiting for input (summary only)."}
  }
}`

func NewGetExecution(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_execution",
		mcp.WithDescription("Get one pipeline execution by ID. By default returns a compact summary: overall status, trigger, timing, and per stage its id, name, type, status, timing and first error, plus the stages running now and any manual judgment waiting for input. Use it to answer \"why did my deploy fail\" or \"is it waiting on me\". Pass summary=false for the full execution JSON with stage contexts and outputs, which is often hundreds of KB. Get the execution_id from trigger_pipeline, list_executions or search_executions."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID (returned by trigger_pipeline, list_executions or search_executions)"),
		),
		mcp.WithBoolean("summary",
			mcp.Description("Return the compact summary (default true). false returns the full execution JSON."),
			mcp.DefaultBool(true),
		),
		readOnly("Get execution"),
		mcp.WithRawOutputSchema(json.RawMessage(executionOutputSchema)),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.GetExecution(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if !req.GetBool("summary", true) {
			return jsonResult(fmt.Sprintf("Execution %s:\n%s", id, string(resp)), resp), nil
		}
		sum, err := summarizeExecution(resp)
		if err != nil {
			return jsonResult(fmt.Sprintf("Execution %s (not summarizable: %v):\n%s", id, err, string(resp)), resp), nil
		}
		b, _ := json.Marshal(sum)
		return mcp.NewToolResultStructured(sum, fmt.Sprintf("Execution %s summary:\n%s", id, b)), nil
	}

	return tool, handler
}

// ExecutionSummary is the compact view of an Orca pipeline execution.
type ExecutionSummary struct {
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	Application string         `json:"application,omitempty"`
	Status      string         `json:"status"`
	TriggerType string         `json:"triggerType,omitempty"`
	TriggeredBy string         `json:"triggeredBy,omitempty"`
	StartTime   int64          `json:"startTime,omitempty"`
	EndTime     int64          `json:"endTime,omitempty"`
	DurationMs  int64          `json:"durationMs,omitempty"`
	Error       string         `json:"error,omitempty"`
	Running     []StageSummary `json:"running"`
	Judgments   []StageSummary `json:"awaitingJudgment"`
	Stages      []StageSummary `json:"stages"`
}

// StageSummary is one stage of an ExecutionSummary.
type StageSummary struct {
	ID         string `json:"id"`
	RefID      string `json:"refId,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	StartTime  int64  `json:"startTime,omitempty"`
	EndTime    int64  `json:"endTime,omitempty"`
	DurationMs int64  `json:"durationMs,omitempty"`
	Error      string `json:"error,omitempty"`
}

type rawStage struct {
	ID        string         `json:"id"`
	RefID     string         `json:"refId"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	StartTime int64          `json:"startTime"`
	EndTime   int64          `json:"endTime"`
	Context   map[string]any `json:"context"`
}

type rawExecution struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Application string `json:"application"`
	Status      string `json:"status"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Trigger     struct {
		Type string `json:"type"`
		User string `json:"user"`
	} `json:"trigger"`
	Stages []rawStage `json:"stages"`
}

// maxStageError bounds each error message in a summary.
const maxStageError = 500

func summarizeExecution(body []byte) (ExecutionSummary, error) {
	var ex rawExecution
	if err := json.Unmarshal(body, &ex); err != nil {
		return ExecutionSummary{}, err
	}
	if ex.ID == "" && ex.Status == "" {
		return ExecutionSummary{}, fmt.Errorf("no id or status in the response")
	}
	s := ExecutionSummary{
		ID: ex.ID, Name: ex.Name, Application: ex.Application, Status: ex.Status,
		TriggerType: ex.Trigger.Type, TriggeredBy: ex.Trigger.User,
		StartTime: ex.StartTime, EndTime: ex.EndTime, DurationMs: duration(ex.StartTime, ex.EndTime),
		Running: []StageSummary{}, Judgments: []StageSummary{}, Stages: []StageSummary{},
	}
	for _, st := range ex.Stages {
		ss := StageSummary{
			ID: st.ID, RefID: st.RefID, Name: st.Name, Type: st.Type, Status: st.Status,
			StartTime: st.StartTime, EndTime: st.EndTime, DurationMs: duration(st.StartTime, st.EndTime),
			Error: stageError(st.Context),
		}
		s.Stages = append(s.Stages, ss)
		if st.Status == "RUNNING" {
			s.Running = append(s.Running, ss)
			if st.Type == "manualJudgment" {
				s.Judgments = append(s.Judgments, ss)
			}
		}
		if s.Error == "" && ss.Error != "" {
			s.Error = fmt.Sprintf("%s: %s", ss.Name, ss.Error)
		}
	}
	return s, nil
}

func duration(start, end int64) int64 {
	if start > 0 && end >= start {
		return end - start
	}
	return 0
}

// stageError digs the first failure message out of a stage context: Orca's
// exception.details (errors or error), else the last failed Clouddriver task.
func stageError(ctx map[string]any) string {
	if exc, ok := ctx["exception"].(map[string]any); ok {
		if d, ok := exc["details"].(map[string]any); ok {
			if errs, ok := d["errors"].([]any); ok && len(errs) > 0 {
				var parts []string
				for _, e := range errs {
					if s, ok := e.(string); ok && s != "" {
						parts = append(parts, s)
					}
				}
				if len(parts) > 0 {
					return truncate(strings.Join(parts, "; "), maxStageError)
				}
			}
			if s, ok := d["error"].(string); ok && s != "" {
				return truncate(s, maxStageError)
			}
		}
	}
	if tasks, ok := ctx["kato.tasks"].([]any); ok {
		for i := len(tasks) - 1; i >= 0; i-- {
			t, _ := tasks[i].(map[string]any)
			if e, ok := t["exception"].(map[string]any); ok {
				if s, ok := e["message"].(string); ok && s != "" {
					return truncate(s, maxStageError)
				}
			}
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}
