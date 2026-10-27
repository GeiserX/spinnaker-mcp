package tools

import "github.com/mark3labs/mcp-go/mcp"

func boolPtr(b bool) *bool { return &b }

// annotate gives a tool its display title and all four behaviour hints, set explicitly so
// no client falls back to the spec defaults (destructive, open world). Every tool talks to
// Gate, an external system, so openWorldHint is always true.
func annotate(title string, readOnly, destructive, idempotent bool) mcp.ToolOption {
	return func(t *mcp.Tool) {
		t.Title = title
		t.Annotations = mcp.ToolAnnotation{
			Title:           title,
			ReadOnlyHint:    boolPtr(readOnly),
			DestructiveHint: boolPtr(destructive),
			IdempotentHint:  boolPtr(idempotent),
			OpenWorldHint:   boolPtr(true),
		}
	}
}

// readOnly marks a tool that only reads from Gate. evaluate_expression POSTs, but Orca
// evaluates the expression without changing anything.
func readOnly(title string) mcp.ToolOption {
	return annotate(title, true, false, true)
}

// mutating marks a tool that changes Spinnaker state in a way that is easy to undo and
// destroys nothing (pause and resume).
func mutating(title string, idempotent bool) mcp.ToolOption {
	return annotate(title, false, false, idempotent)
}

// destructive marks a tool that can deploy, destroy, stop or overwrite something:
// triggering a pipeline can run Destroy Server Group stages, saving a pipeline with an
// existing id overwrites it, cancelling stops a deploy half way. Clients should confirm.
func destructive(title string, idempotent bool) mcp.ToolOption {
	return annotate(title, false, true, idempotent)
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
