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
		mcp.WithDescription("Get available tags for a Docker image repository registered in Spinnaker. Use this to list version tags before triggering a pipeline with a specific image tag. Returns JSON array of tag strings for the specified repository."),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("Spinnaker account name"),
		),
		mcp.WithString("repository",
			mcp.Required(),
			mcp.Description("Docker image repository (e.g. library/nginx)"),
		),
		readOnly(),
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

		resp, err := gate.GetImageTags(ctx, account, repository)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Tags for %q (account %q):\n%s", repository, account, string(resp))), nil
	}

	return tool, handler
}
