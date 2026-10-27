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
		mcp.WithDescription("List recent pipeline executions for a Spinnaker application, optionally filtered by status. Use this to monitor deployment activity or find a specific execution ID. Returns JSON array of execution summaries with status, timing, and trigger info. Use get_execution for full details of a specific run."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithInteger("limit",
			mcp.Description("Maximum number of executions to return (default 25)"),
			mcp.Min(1), mcp.Max(100),
		),
		mcp.WithString("statuses",
			mcp.Description("Comma-separated execution statuses to filter by (e.g. RUNNING,SUCCEEDED,TERMINAL,CANCELED)"),
		),
		readOnly("List executions"),
		jsonOutput(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := clamp(req.GetInt("limit", 25), 1, 100)
		statuses := req.GetString("statuses", "")

		resp, err := gate.ListExecutions(ctx, app, limit, statuses)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Executions for %q (limit %d):\n%s", app, limit, string(resp)), resp), nil
	}

	return tool, handler
}
