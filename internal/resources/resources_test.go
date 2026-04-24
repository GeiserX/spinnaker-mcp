package resources

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestTextResource_NonEmpty(t *testing.T) {
	rc := textResource("spinnaker://test", []byte(`{"key":"value"}`))
	if rc.URI != "spinnaker://test" {
		t.Errorf("URI = %q, want %q", rc.URI, "spinnaker://test")
	}
	if rc.MIMEType != "application/json" {
		t.Errorf("MIMEType = %q, want %q", rc.MIMEType, "application/json")
	}
	if rc.Text != `{"key":"value"}` {
		t.Errorf("Text = %q, want %q", rc.Text, `{"key":"value"}`)
	}
}

func TestTextResource_EmptyDefaultsToArray(t *testing.T) {
	rc := textResource("spinnaker://test", []byte(""))
	if rc.Text != "[]" {
		t.Errorf("Text = %q, want %q", rc.Text, "[]")
	}
}

func TestTextResource_NilDefaultsToArray(t *testing.T) {
	rc := textResource("spinnaker://test", nil)
	if rc.Text != "[]" {
		t.Errorf("Text = %q, want %q", rc.Text, "[]")
	}
}

func TestExtractParam_FromArguments(t *testing.T) {
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "spinnaker://application/myapp"
	req.Params.Arguments = map[string]any{"name": "fromargs"}

	got := extractParam(req, "name")
	if got != "fromargs" {
		t.Errorf("extractParam() = %q, want %q", got, "fromargs")
	}
}

func TestExtractParam_FallbackToURI(t *testing.T) {
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "spinnaker://application/myapp"

	got := extractParam(req, "name")
	if got != "myapp" {
		t.Errorf("extractParam() = %q, want %q", got, "myapp")
	}
}

func TestParseURIParam_ApplicationName(t *testing.T) {
	tests := []struct {
		uri  string
		key  string
		want string
	}{
		{"spinnaker://application/myapp", "name", "myapp"},
		{"spinnaker://application/myapp/pipelines", "name", "myapp"},
		{"spinnaker://account/prod", "name", "prod"},
		{"spinnaker://execution/abc123", "id", "abc123"},
		{"spinnaker://application/myapp", "id", ""},
		{"spinnaker://execution/abc123", "name", ""},
		{"spinnaker://", "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.uri+"_"+tt.key, func(t *testing.T) {
			got := parseURIParam(tt.uri, tt.key)
			if got != tt.want {
				t.Errorf("parseURIParam(%q, %q) = %q, want %q", tt.uri, tt.key, got, tt.want)
			}
		})
	}
}
