package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetScalingActivities(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_scaling_activities",
		mcp.WithDescription("Get scaling activities for a server group in a Spinnaker cluster. Use this to audit recent auto-scaling events (scale-up, scale-down) and diagnose capacity issues. Returns JSON array of scaling activity records with timestamps, descriptions, and status."),
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
			mcp.Description("Cluster name as shown in the Spinnaker UI"),
		),
		mcp.WithString("server_group_name",
			mcp.Required(),
			mcp.Description("Server group name (e.g., myapp-v001)"),
		),
		mcp.WithString("provider",
			mcp.Description("Cloud provider (e.g. aws, gce, kubernetes)"),
		),
		readOnly(),
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
		serverGroup, err := req.RequireString("server_group_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		provider := req.GetString("provider", "")

		resp, err := gate.GetScalingActivities(ctx, app, account, cluster, serverGroup, provider)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Scaling activities for server group %q in cluster %q:\n%s", serverGroup, cluster, string(resp))), nil
	}

	return tool, handler
}
