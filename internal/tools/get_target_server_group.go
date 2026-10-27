package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// targetSelectors are the values Clouddriver's TargetServerGroup accepts.
var targetSelectors = []string{"current", "newest", "previous", "ancestor", "oldest", "largest", "fail"}

func NewGetTargetServerGroup(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_target_server_group",
		mcp.WithDescription("Get one server group of a cluster by a target selector instead of its name: current (or newest), previous (or ancestor), oldest, largest, or fail (errors unless exactly one server group exists). Use this for rollback planning or to find what is live now. Returns JSON with the matched server group."),
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
		mcp.WithString("cloud_provider",
			mcp.Required(),
			mcp.Description("Cloud provider (e.g. aws, gce, kubernetes)"),
		),
		mcp.WithString("scope",
			mcp.Required(),
			mcp.Description("Scope for target lookup (e.g. a region like us-east-1)"),
		),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target selector, as Clouddriver names it: current and newest are the latest server group, previous and ancestor the one before it"),
			mcp.Enum(targetSelectors...),
		),
		readOnly("Find target server group"),
		jsonOutput(),
	)

	validTargets := map[string]bool{}
	for _, t := range targetSelectors {
		validTargets[t] = true
	}

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
		cloudProvider, err := req.RequireString("cloud_provider")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		scope, err := req.RequireString("scope")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		target, err := req.RequireString("target")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if !validTargets[target] {
			return mcp.NewToolResultError(fmt.Sprintf("invalid target %q: must be one of %s", target, strings.Join(targetSelectors, ", "))), nil
		}

		resp, err := gate.GetTargetServerGroup(ctx, app, account, cluster, cloudProvider, scope, target)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Target server group (%s) in cluster %q:\n%s", target, cluster, string(resp)), resp), nil
	}

	return tool, handler
}
