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
		mcp.WithDescription("Cancel a running pipeline execution"),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to cancel"),
		),
		mcp.WithString("reason",
			mcp.Description("Human-readable reason for cancellation"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		reason := req.GetString("reason", "")

		resp, err := gate.CancelExecution(ctx, id, reason)
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
