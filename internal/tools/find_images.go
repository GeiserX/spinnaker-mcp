package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewFindImages(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("find_images",
		mcp.WithDescription("Search for machine images (AMIs, GCE images, Docker images) available in Spinnaker for a given cloud provider. Use this to find images for a deployment or verify which image versions are available. Supports filtering by query string, region, and account. Returns JSON array of matching image objects."),
		mcp.WithString("provider",
			mcp.Required(),
			mcp.Description("Cloud provider (e.g. aws, gce, docker)"),
		),
		mcp.WithString("query",
			mcp.Description("Search query to filter images"),
		),
		mcp.WithString("region",
			mcp.Description("Cloud region to filter images"),
		),
		mcp.WithString("account",
			mcp.Description("Spinnaker account to filter images"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		provider, err := req.RequireString("provider")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{"provider": provider}
		if q := req.GetString("query", ""); q != "" {
			params["q"] = q
		}
		if region := req.GetString("region", ""); region != "" {
			params["region"] = region
		}
		if account := req.GetString("account", ""); account != "" {
			params["account"] = account
		}

		resp, err := gate.FindImages(ctx, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Images for provider %q:\n%s", provider, string(resp))), nil
	}

	return tool, handler
}
