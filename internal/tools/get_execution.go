package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetExecution(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_execution",
		mcp.WithDescription("Get the full details of a pipeline execution including stage statuses, outputs, and timing"),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID (UUID returned when triggering or from list_executions)"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.GetExecution(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Execution %s:\n%s", id, string(resp))), nil
	}

	return tool, handler
}
