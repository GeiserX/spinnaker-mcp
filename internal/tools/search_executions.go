package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewSearchExecutions(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("search_executions",
		mcp.WithDescription("Search pipeline executions across an application with rich filters such as status, trigger type, and time range. Use this instead of list_executions when you need to find executions matching specific criteria. Returns JSON array of matching execution summaries."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("trigger_type",
			mcp.Description("Filter by trigger type (e.g. manual, webhook, cron)"),
		),
		mcp.WithString("statuses",
			mcp.Description("Comma-separated execution statuses to filter by (e.g. RUNNING,SUCCEEDED,TERMINAL)"),
		),
		mcp.WithString("start_time",
			mcp.Description("Filter executions started after this time (ISO 8601)"),
		),
		mcp.WithString("end_time",
			mcp.Description("Filter executions started before this time (ISO 8601)"),
		),
		mcp.WithString("event_id",
			mcp.Description("Filter by event ID"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]string)
		if v := req.GetString("trigger_type", ""); v != "" {
			params["triggerType"] = v
		}
		if v := req.GetString("statuses", ""); v != "" {
			params["statuses"] = v
		}
		if v := req.GetString("start_time", ""); v != "" {
			params["startTime"] = v
		}
		if v := req.GetString("end_time", ""); v != "" {
			params["endTime"] = v
		}
		if v := req.GetString("event_id", ""); v != "" {
			params["eventId"] = v
		}

		resp, err := gate.SearchExecutions(ctx, app, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Execution search results for %q:\n%s", app, string(resp))), nil
	}

	return tool, handler
}
