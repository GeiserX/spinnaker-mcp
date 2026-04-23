package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewDeleteStrategy(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("delete_strategy",
		mcp.WithDescription("Delete a deployment strategy configuration from Spinnaker"),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("strategy_name",
			mcp.Required(),
			mcp.Description("Strategy name to delete"),
		),
		destructive(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := req.RequireString("strategy_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.DeleteStrategy(ctx, app, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Strategy %q deleted from %q successfully.", name, app)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Strategy %q deleted from %q. Response:\n%s", name, app, string(resp))), nil
	}

	return tool, handler
}
