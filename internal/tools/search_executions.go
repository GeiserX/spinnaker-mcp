package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewSearchExecutions(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("search_executions",
		mcp.WithDescription("Search pipeline executions of an application by pipeline name, trigger type, status and trigger time. Use this instead of list_executions when you need executions matching specific criteria, for example every failed webhook run since yesterday. Returns a JSON array of executions, newest first, size per page (default 10)."),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("pipeline_name",
			mcp.Description("Only executions of this pipeline (exact, case-sensitive name as shown in the UI)"),
		),
		mcp.WithString("trigger_type",
			mcp.Description("Trigger types, comma-separated (e.g. manual, webhook, cron, docker, jenkins, pipeline)"),
		),
		mcp.WithString("statuses",
			mcp.Description("Comma-separated execution statuses (e.g. RUNNING,SUCCEEDED,TERMINAL,CANCELED,NOT_STARTED)"),
		),
		mcp.WithString("start_time",
			mcp.Description("Only executions triggered at or after this time: RFC 3339 (2026-07-28T09:00:00Z), a date (2026-07-28, UTC midnight) or epoch milliseconds"),
		),
		mcp.WithString("end_time",
			mcp.Description("Only executions triggered at or before this time, same formats as start_time"),
		),
		mcp.WithString("event_id",
			mcp.Description("Only executions started by this trigger event ID"),
		),
		mcp.WithInteger("size",
			mcp.Description("Executions per page (default 10)"),
			mcp.Min(1), mcp.Max(100),
		),
		mcp.WithInteger("start_index",
			mcp.Description("Zero-based index of the first execution to return, for paging (default 0)"),
			mcp.Min(0),
		),
		readOnly("Search executions"),
		jsonOutput(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Gate's ExecutionsController names: triggerTypes, and the trigger time
		// boundaries in epoch milliseconds. Unknown names are silently ignored by Gate.
		params := map[string]string{
			"pipelineName": req.GetString("pipeline_name", ""),
			"triggerTypes": req.GetString("trigger_type", ""),
			"statuses":     req.GetString("statuses", ""),
			"eventId":      req.GetString("event_id", ""),
		}
		for arg, gateName := range map[string]string{"start_time": "triggerTimeStartBoundary", "end_time": "triggerTimeEndBoundary"} {
			raw := req.GetString(arg, "")
			if raw == "" {
				continue
			}
			ms, err := epochMillis(raw)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("%s: %v", arg, err)), nil
			}
			params[gateName] = strconv.FormatInt(ms, 10)
		}
		if size := req.GetInt("size", 0); size > 0 {
			params["size"] = strconv.Itoa(min(size, 100))
		}
		if idx := req.GetInt("start_index", 0); idx > 0 {
			params["startIndex"] = strconv.Itoa(idx)
		}

		resp, err := gate.SearchExecutions(ctx, app, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Execution search results for %q:\n%s", app, string(resp)), resp), nil
	}

	return tool, handler
}

// epochMillis parses epoch milliseconds, an RFC 3339 timestamp or a plain date (UTC).
func epochMillis(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ms, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UnixMilli(), nil
	}
	if t, err := time.Parse(time.DateOnly, s); err == nil {
		return t.UnixMilli(), nil
	}
	return 0, fmt.Errorf("%q is not epoch milliseconds, RFC 3339 or YYYY-MM-DD", s)
}
