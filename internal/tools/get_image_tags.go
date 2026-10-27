package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetImageTags(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_image_tags",
		mcp.WithDescription("Get available tags for a Docker image repository registered in Spinnaker. Use this to list version tags before triggering a pipeline with a specific image tag. Returns the tag strings for the specified repository as {items, count}."),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("Spinnaker account name"),
		),
		mcp.WithString("repository",
			mcp.Required(),
			mcp.Description("Docker image repository (e.g. library/nginx)"),
		),
		mcp.WithString("provider",
			mcp.Description("Image provider of the account (default dockerRegistry; Gate itself would assume aws)"),
			mcp.DefaultString("dockerRegistry"),
		),
		readOnly("List image tags"),
		jsonOutput(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account, err := req.RequireString("account")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repository, err := req.RequireString("repository")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		provider := req.GetString("provider", "dockerRegistry")
		resp, err := gate.GetImageTags(ctx, account, repository, provider)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Tags for %q (account %q):\n%s", repository, account, string(resp)), resp), nil
	}

	return tool, handler
}
