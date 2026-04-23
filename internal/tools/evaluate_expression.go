package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// dangerousSpEL blocks Java reflection and class instantiation patterns
var dangerousSpEL = regexp.MustCompile(`(?i)(T\s*\(|\.class\b|\.getClass\s*\(|Runtime|ProcessBuilder|Thread|ClassLoader|URLClassLoader|ScriptEngine|MethodHandle)`)

func NewEvaluateExpression(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("evaluate_expression",
		mcp.WithDescription("Evaluate a SpEL expression against a pipeline execution context"),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to evaluate against"),
		),
		mcp.WithString("expression",
			mcp.Required(),
			mcp.Description("SpEL expression string to evaluate"),
		),
		readOnly(),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executionID, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		expression, err := req.RequireString("expression")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(expression) > 4096 {
			return mcp.NewToolResultError("expression exceeds maximum length of 4096 characters"), nil
		}
		if dangerousSpEL.MatchString(expression) {
			return mcp.NewToolResultError("expression contains disallowed patterns (type references, reflection, or dangerous classes are not permitted via MCP)"), nil
		}
		if strings.Contains(expression, "new ") {
			return mcp.NewToolResultError("expression contains disallowed 'new' keyword (object instantiation is not permitted via MCP)"), nil
		}

		resp, err := gate.EvaluateExpression(ctx, executionID, expression)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Expression result for execution %s:\n%s", executionID, string(resp))), nil
	}

	return tool, handler
}
