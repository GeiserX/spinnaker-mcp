package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewListSubnets(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_subnets",
		mcp.WithDescription("List all subnets for a given cloud provider across Spinnaker accounts. Use this to discover available subnets when configuring server groups or deployment targets. Returns JSON array of subnet objects with CIDR, availability zone, and VPC association."),
		mcp.WithString("cloud_provider",
			mcp.Required(),
			mcp.Description("Cloud provider (e.g. aws, gce, kubernetes)"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cloudProvider, err := req.RequireString("cloud_provider")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.ListSubnets(ctx, cloudProvider)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Subnets for %q:\n%s", cloudProvider, string(resp))), nil
	}

	return tool, handler
}
