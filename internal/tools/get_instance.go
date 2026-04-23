package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetInstance(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_instance",
		mcp.WithDescription("Get details for a specific compute instance by account, region, and instance ID. Use this to inspect instance health, metadata, launch time, and cloud-specific attributes. Returns JSON with full instance details including health indicators."),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("Spinnaker account name"),
		),
		mcp.WithString("region",
			mcp.Required(),
			mcp.Description("Cloud region (e.g. us-east-1)"),
		),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("Cloud provider instance ID (e.g., i-0abc1234def56789 for AWS)"),
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
		instanceID, err := req.RequireString("instance_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.GetInstance(ctx, account, region, instanceID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Instance %q in %s/%s:\n%s", instanceID, account, region, string(resp))), nil
	}

	return tool, handler
}
