package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewListExecutions(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_executions",
		mcp.WithDescription("List recent pipeline executions for a Spinnaker application, optionally filtered by status"),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of executions to return (default: 25)"),
		),
		mcp.WithString("statuses",
			mcp.Description("Comma-separated execution statuses to filter by (e.g. RUNNING,SUCCEEDED,TERMINAL,CANCELED)"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := req.GetInt("limit", 25)
		statuses := req.GetString("statuses", "")

		resp, err := gate.ListExecutions(ctx, app, limit, statuses)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Executions for %q (limit %d):\n%s", app, limit, string(resp))), nil
	}

	return tool, handler
}
