package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewGetTask(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_task",
		mcp.WithDescription("Get the status and details of a Spinnaker orchestration task (e.g. deploy, resize, rollback)"),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("Orchestration task ID returned by Spinnaker when an operation is submitted"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := gate.GetTask(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task %s:\n%s", id, string(resp))), nil
	}

	return tool, handler
}
