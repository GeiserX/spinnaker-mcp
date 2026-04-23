package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewUpdatePipeline(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("update_pipeline",
		mcp.WithDescription("Update an existing pipeline configuration in Spinnaker"),
		mcp.WithString("pipeline_id",
			mcp.Required(),
			mcp.Description("Pipeline configuration ID to update"),
		),
		mcp.WithString("pipeline_json",
			mcp.Required(),
			mcp.Description("Updated pipeline JSON definition"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pipelineID, err := req.RequireString("pipeline_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		raw, err := req.RequireString("pipeline_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		const maxPipelineSize = 1 << 20 // 1 MB
		if len(raw) > maxPipelineSize {
			return mcp.NewToolResultError(fmt.Sprintf("pipeline JSON exceeds maximum size of %d bytes", maxPipelineSize)), nil
		}

		var pipeline map[string]any
		if err := json.Unmarshal([]byte(raw), &pipeline); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid pipeline JSON: %v", err)), nil
		}

		resp, err := gate.UpdatePipeline(ctx, pipelineID, pipeline)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Pipeline %s updated successfully.", pipelineID)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Pipeline %s updated. Response:\n%s", pipelineID, string(resp))), nil
	}

	return tool, handler
}
