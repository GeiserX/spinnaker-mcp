package tools

import (
	"bytes"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// gateJSONSchema is the outputSchema of every tool that relays a Gate response. Gate's
// payloads are Spinnaker's own models and vary by cloud provider and version, so the
// schema pins the envelope rather than every field.
const gateJSONSchema = `{
  "type": "object",
  "description": "The Gate response. A JSON object from Gate comes back as is; a JSON array comes back as {items, count}.",
  "properties": {
    "items": {"type": "array", "description": "The records, when Gate answered with a list."},
    "count": {"type": "integer", "minimum": 0, "description": "Number of records in items."}
  }
}`

// jsonOutput declares gateJSONSchema as the tool's outputSchema.
func jsonOutput() mcp.ToolOption {
	return mcp.WithRawOutputSchema(json.RawMessage(gateJSONSchema))
}

// jsonResult returns text for clients that read only content, and the same Gate data as
// structuredContent matching gateJSONSchema.
func jsonResult(text string, body []byte) *mcp.CallToolResult {
	return mcp.NewToolResultStructured(structure(body), text)
}

// structure decodes a Gate body into an object: arrays become {items, count}, an empty
// body an empty list, and anything that is not JSON is kept as {"text": ...}.
func structure(body []byte) map[string]any {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return map[string]any{"items": []any{}, "count": 0}
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return map[string]any{"text": string(trimmed)}
	}
	switch t := v.(type) {
	case map[string]any:
		return t
	case []any:
		return map[string]any{"items": t, "count": len(t)}
	default:
		return map[string]any{"value": t}
	}
}
