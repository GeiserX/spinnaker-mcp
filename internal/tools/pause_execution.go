package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewPauseExecution(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("pause_execution",
		mcp.WithDescription("Pause a running pipeline execution at the current stage boundary. Use this to temporarily halt a deployment for manual review before it continues to the next stage. The execution can be resumed later with resume_execution. Returns a confirmation message."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to pause"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.PauseExecution(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Execution %s paused successfully.", id)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Execution %s paused. Response:\n%s", id, string(resp))), nil
	}

	return tool, handler
}
