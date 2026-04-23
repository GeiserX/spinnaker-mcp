package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewSavePipeline(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("save_pipeline",
		mcp.WithDescription("Save a new pipeline configuration to Spinnaker"),
		mcp.WithString("pipeline_json",
			mcp.Required(),
			mcp.Description("Full pipeline JSON definition"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		if _, ok := pipeline["application"].(string); !ok {
			return mcp.NewToolResultError("pipeline JSON must include an 'application' string field"), nil
		}
		if _, ok := pipeline["name"].(string); !ok {
			return mcp.NewToolResultError("pipeline JSON must include a 'name' string field"), nil
		}

		resp, err := gate.SavePipeline(ctx, pipeline)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText("Pipeline saved successfully."), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Pipeline saved. Response:\n%s", string(resp))), nil
	}

	return tool, handler
}
