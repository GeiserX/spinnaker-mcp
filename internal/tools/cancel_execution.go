package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewCancelExecution(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("cancel_execution",
		mcp.WithDescription("Cancel a running pipeline execution by its ID, optionally providing a reason. Use this to stop a deployment in progress — for example, when a bad build was triggered. Does not roll back completed stages. Returns a confirmation message."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to cancel"),
		),
		mcp.WithString("reason",
			mcp.Description("Human-readable reason for cancellation, shown in the Spinnaker UI"),
		),
		mcp.WithBoolean("force",
			mcp.Description("Cancel even if a stage is not cancellable (default false). Use only when a normal cancel left the execution running."),
			mcp.DefaultBool(false),
		),
		destructive("Cancel execution", true),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		reason := req.GetString("reason", "")

		resp, err := gate.CancelExecution(ctx, id, reason, req.GetBool("force", false))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Execution %s cancelled successfully.", id)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Execution %s cancelled. Response:\n%s", id, string(resp))), nil
	}

	return tool, handler
}
