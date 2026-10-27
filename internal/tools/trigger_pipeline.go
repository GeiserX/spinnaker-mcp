package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const triggerOutputSchema = `{
  "type": "object",
  "properties": {
    "ref": {"type": "string", "description": "Gate's reference to the new execution, e.g. /pipelines/01J..."},
    "executionId": {"type": "string", "description": "The execution ID, for get_execution, cancel_execution and judge_stage"}
  }
}`

func NewTriggerPipeline(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("trigger_pipeline",
		mcp.WithDescription("Start a new execution of a pipeline, as a manual trigger, with optional parameters. This runs the real pipeline: it can deploy, resize or destroy infrastructure, so confirm the application, pipeline and parameters with the user first. Returns the execution ID; poll get_execution to follow it."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("pipeline_name",
			mcp.Required(),
			mcp.Description("Pipeline name as shown in the Spinnaker UI (exact, case-sensitive)"),
		),
		mcp.WithObject("parameters",
			mcp.Description(`Pipeline parameters as a JSON object, keys as in the pipeline's parameterConfig (e.g. {"tag":"v1.2.3","env":"staging"}). A JSON-encoded string is accepted too.`),
		),
		destructive("Trigger pipeline", false),
		mcp.WithRawOutputSchema(json.RawMessage(triggerOutputSchema)),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := req.RequireString("pipeline_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params, err := objectArg(req, "parameters")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid parameters: %v", err)), nil
		}

		resp, err := gate.TriggerPipeline(ctx, app, name, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var out struct {
			Ref         string `json:"ref"`
			ExecutionID string `json:"executionId,omitempty"`
		}
		_ = json.Unmarshal(resp, &out)
		if strings.HasPrefix(out.Ref, "/pipelines/") {
			out.ExecutionID = path.Base(out.Ref)
		}
		text := fmt.Sprintf("Pipeline %q triggered in %q. Response:\n%s", name, app, string(resp))
		if out.ExecutionID != "" {
			text = fmt.Sprintf("Pipeline %q triggered in %q. Execution ID: %s\nResponse:\n%s", name, app, out.ExecutionID, string(resp))
		}
		return mcp.NewToolResultStructured(out, text), nil
	}

	return tool, handler
}

// objectArg reads an optional object argument. Models often send JSON objects as
// JSON-encoded strings, so a string holding an object is accepted as well.
func objectArg(req mcp.CallToolRequest, key string) (map[string]any, error) {
	v, ok := req.GetArguments()[key]
	if !ok || v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case map[string]any:
		return t, nil
	case string:
		if strings.TrimSpace(t) == "" {
			return nil, nil
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(t), &m); err != nil {
			return nil, fmt.Errorf("expected a JSON object: %v", err)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("expected a JSON object, got %T", v)
	}
}
