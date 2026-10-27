package tools

import (
	"context"
	"fmt"
	"regexp"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// dangerousSpEL rejects type references, object construction, reflection and JDK classes
// that reach the OS, network or class loading. Orca evaluates expressions in its own
// restricted context, which is the real boundary; this check stops the obvious attempts
// before they reach Gate. It is case-sensitive, like Java identifiers, so properties such
// as parameters.runtimeVersion pass, and it matches "new" followed by any whitespace.
var dangerousSpEL = regexp.MustCompile(`\bT\s*\(|(?i:\bnew\s)|\.class\b|\.getClass\s*\(|\bjava\.(io|net|nio|lang\.reflect|lang\.invoke)\.|\b(Runtime|ProcessBuilder|Thread|ClassLoader|URLClassLoader|ScriptEngine|ScriptEngineManager|MethodHandles?|System)\b`)

func NewEvaluateExpression(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("evaluate_expression",
		mcp.WithDescription("Evaluate a Spring Expression Language (SpEL) expression against a pipeline execution context. Use this to inspect pipeline variables, stage outputs, or computed values during debugging. Dangerous patterns (reflection, class instantiation) are blocked for security. Returns the evaluated expression result as JSON."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID to evaluate against"),
		),
		mcp.WithString("expression",
			mcp.Required(),
			mcp.Description("SpEL expression in ${...}, e.g. ${trigger.parameters.tag} or ${#stage('Deploy').status}"),
			mcp.MaxLength(4096),
		),
		readOnly("Evaluate SpEL expression"),
		jsonOutput(),
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

		resp, err := gate.EvaluateExpression(ctx, executionID, expression)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(fmt.Sprintf("Expression result for execution %s:\n%s", executionID, string(resp)), resp), nil
	}

	return tool, handler
}
