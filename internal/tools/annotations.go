package tools

import "github.com/mark3labs/mcp-go/mcp"

func boolPtr(b bool) *bool { return &b }

// readOnly marks tools that only read data (GET requests).
func readOnly() mcp.ToolOption {
	return mcp.WithToolAnnotation(mcp.ToolAnnotation{
		ReadOnlyHint:    boolPtr(true),
		DestructiveHint: boolPtr(false),
		IdempotentHint:  boolPtr(true),
		OpenWorldHint:   boolPtr(true),
	})
}

// mutating marks tools that create or modify resources (POST/PUT).
func mutating() mcp.ToolOption {
	return mcp.WithToolAnnotation(mcp.ToolAnnotation{
		ReadOnlyHint:    boolPtr(false),
		DestructiveHint: boolPtr(false),
		IdempotentHint:  boolPtr(false),
		OpenWorldHint:   boolPtr(true),
	})
}

// destructive marks tools that delete resources (DELETE).
func destructive() mcp.ToolOption {
	return mcp.WithToolAnnotation(mcp.ToolAnnotation{
		ReadOnlyHint:    boolPtr(false),
		DestructiveHint: boolPtr(true),
		IdempotentHint:  boolPtr(false),
		OpenWorldHint:   boolPtr(true),
	})
}
