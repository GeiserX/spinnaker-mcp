package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetTargetServerGroup(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_target_server_group",
		mcp.WithDescription("Get a target server group in a Spinnaker cluster by selection strategy"),
		mcp.WithString("application",
			mcp.Required(),
			mcp.Description("Application name as registered in Spinnaker"),
		),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("Spinnaker account name"),
		),
		mcp.WithString("cluster_name",
			mcp.Required(),
			mcp.Description("Cluster name"),
		),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target type: newest, oldest, largest, smallest, or ancestor"),
		),
		mcp.WithString("cloud_provider",
			mcp.Description("Cloud provider (e.g. aws, gce, kubernetes)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		app, err := req.RequireString("application")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		account, err := req.RequireString("account")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		cluster, err := req.RequireString("cluster_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		target, err := req.RequireString("target")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		cloudProvider := req.GetString("cloud_provider", "")

		resp, err := gate.GetTargetServerGroup(ctx, app, account, cluster, target, cloudProvider)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Target server group (%s) in cluster %q:\n%s", target, cluster, string(resp))), nil
	}

	return tool, handler
}
