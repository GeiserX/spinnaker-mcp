package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetFirewall(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_firewall",
		mcp.WithDescription("Get details for a specific firewall rule (security group) by account, region, and name. Use this to inspect inbound/outbound rules and associated resources. Returns JSON with the firewall's full rule set and metadata."),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("Spinnaker account name"),
		),
		mcp.WithString("region",
			mcp.Required(),
			mcp.Description("Cloud region (e.g. us-east-1)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Firewall rule name"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account, err := req.RequireString("account")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		region, err := req.RequireString("region")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.GetFirewall(ctx, account, region, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Firewall %q in %s/%s:\n%s", name, account, region, string(resp))), nil
	}

	return tool, handler
}
