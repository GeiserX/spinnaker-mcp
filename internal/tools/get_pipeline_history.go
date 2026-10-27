package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetPipelineHistory(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_pipeline_history",
		mcp.WithDescription("Get the version history of a pipeline configuration showing past revisions. Use this to audit changes to a pipeline or compare previous configurations. Returns JSON array of historical pipeline versions ordered by update time."),
		mcp.WithString("pipeline_config_id",
			mcp.Required(),
			mcp.Description("Pipeline configuration ID"),
		),
		mcp.WithInteger("limit",
			mcp.Description("Maximum number of history entries to return (default 10)"),
			mcp.Min(1), mcp.Max(100),
		),
		readOnly("Pipeline config history"),
		jsonOutput(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		configID, err := req.RequireString("pipeline_config_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := clamp(req.GetInt("limit", 10), 1, 100)

		resp, err := gate.GetPipelineHistory(ctx, configID, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Pipeline history for %s (limit %d):\n%s", configID, limit, string(resp)), resp), nil
	}

	return tool, handler
}
