package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewRestartStage(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("restart_stage",
		mcp.WithDescription("Restart a specific failed or completed stage within a pipeline execution. Use this to retry a stage that failed due to a transient error without re-running the entire pipeline. Requires the execution ID and stage ID. Returns a confirmation message."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID"),
		),
		mcp.WithString("stage_id",
			mcp.Required(),
			mcp.Description("Stage ID to restart"),
		),
		mutating(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executionID, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stageID, err := req.RequireString("stage_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.RestartStage(ctx, executionID, stageID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(resp) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Stage %s restarted successfully in execution %s.", stageID, executionID)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Stage %s restarted in execution %s. Response:\n%s", stageID, executionID, string(resp))), nil
	}

	return tool, handler
}
