package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetConsoleOutput(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_console_output",
		mcp.WithDescription("Get the console output (boot log) for a specific compute instance. Use this to debug instance startup failures or inspect early boot messages. Requires account, region, and instance ID. Returns the raw console output text."),
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
		mcp.WithString("provider",
			mcp.Description("Cloud provider (e.g. aws, gce, kubernetes)"),
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
		provider := req.GetString("provider", "")

		resp, err := gate.GetConsoleOutput(ctx, account, region, instanceID, provider)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Console output for instance %q in %s/%s:\n%s", instanceID, account, region, string(resp))), nil
	}

	return tool, handler
}
