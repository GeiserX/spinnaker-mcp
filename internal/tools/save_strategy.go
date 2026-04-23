package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewSaveStrategy(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("save_strategy",
		mcp.WithDescription("Save a new deployment strategy configuration to Spinnaker"),
		mcp.WithString("strategy_json",
			mcp.Required(),
			mcp.Description("Full strategy JSON definition"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		raw, err := req.RequireString("strategy_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		const maxStrategySize = 1 << 20 // 1 MB
		if len(raw) > maxStrategySize {
			return mcp.NewToolResultError(fmt.Sprintf("strategy JSON exceeds maximum size of %d bytes", maxStrategySize)), nil
		}

		var strategy map[string]any
		if err := json.Unmarshal([]byte(raw), &strategy); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid strategy JSON: %v", err)), nil
		}

		resp, err := gate.SaveStrategy(ctx, strategy)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText("Strategy saved successfully."), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Strategy saved. Response:\n%s", string(resp))), nil
	}

	return tool, handler
}
