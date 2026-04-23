package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewDeletePipeline(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("delete_pipeline",
		mcp.WithDescription("Permanently delete a pipeline configuration from Spinnaker. Use this only when a pipeline is no longer needed — this action cannot be undone. Does not affect past executions. Returns a confirmation message on success."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("pipeline_name",
			mcp.Required(),
			mcp.Description("Pipeline name to delete"),
		),
		destructive(),
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

		resp, err := gate.DeletePipeline(ctx, app, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Pipeline %q deleted from %q successfully.", name, app)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Pipeline %q deleted from %q. Response:\n%s", name, app, string(resp))), nil
	}

	return tool, handler
}
