package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewListLoadBalancers(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_load_balancers",
		mcp.WithDescription("List all load balancers for a Spinnaker application across all accounts and regions"),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.ListLoadBalancers(ctx, app)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Load balancers for %q:\n%s", app, string(resp))), nil
	}

	return tool, handler
}
