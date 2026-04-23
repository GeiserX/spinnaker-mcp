package tools

import (
	"context"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewListNetworks(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_networks",
		mcp.WithDescription("List all networks (VPCs) across all Spinnaker accounts and cloud providers. Use this to discover available networks when configuring deployment targets or security groups. Returns JSON array of network objects with provider, account, and CIDR details."),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := gate.ListNetworks(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(resp)), nil
	}

	return tool, handler
}
