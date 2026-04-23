package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewTriggerPipeline(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("trigger_pipeline",
		mcp.WithDescription("Trigger a pipeline execution with optional parameters. Returns the execution reference ID."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("pipeline_name",
			mcp.Required(),
			mcp.Description("Pipeline name as shown in the Spinnaker UI"),
		),
		mcp.WithString("parameters",
			mcp.Description("JSON object of pipeline parameters (e.g. {\"tag\":\"v1.2.3\",\"env\":\"staging\"})"),
		),
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

		var params map[string]any
		if raw := req.GetString("parameters", ""); raw != "" {
			if err := json.Unmarshal([]byte(raw), &params); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid parameters JSON: %v", err)), nil
			}
		}

		resp, err := gate.TriggerPipeline(app, name, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Pipeline %q triggered in %q. Response:\n%s", name, app, string(resp))), nil
	}

	return tool, handler
}
