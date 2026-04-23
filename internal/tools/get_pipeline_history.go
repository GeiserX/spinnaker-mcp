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
		mcp.WithDescription("Get the version history of a pipeline configuration"),
		mcp.WithString("pipeline_config_id",
			mcp.Required(),
			mcp.Description("Pipeline configuration ID"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of history entries to return (default: 10)"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		configID, err := req.RequireString("pipeline_config_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := req.GetInt("limit", 10)

		resp, err := gate.GetPipelineHistory(ctx, configID, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Pipeline history for %s (limit %d):\n%s", configID, limit, string(resp))), nil
	}

	return tool, handler
}
