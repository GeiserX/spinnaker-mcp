package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewResumeExecution(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("resume_execution",
		mcp.WithDescription("Resume a paused pipeline execution"),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to resume"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.ResumeExecution(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Execution %s resumed successfully.", id)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Execution %s resumed. Response:\n%s", id, string(resp))), nil
	}

	return tool, handler
}
