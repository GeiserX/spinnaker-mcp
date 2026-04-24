package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestDeployReview(t *testing.T) {
	prompt, handler := deployReview()

	if prompt.Name != "deploy-review" {
		t.Errorf("Name = %q, want %q", prompt.Name, "deploy-review")
	}
	if len(prompt.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(prompt.Arguments))
	}

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
		"pipeline":    "deploy-prod",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "myapp") {
		t.Error("result should contain application name")
	}
	if !strings.Contains(text, "deploy-prod") {
		t.Error("result should contain pipeline name")
	}
	if !strings.Contains(text, "get_pipeline") {
		t.Error("result should reference get_pipeline tool")
	}
}

func TestIncidentResponse_WithExecutionID(t *testing.T) {
	prompt, handler := incidentResponse()

	if prompt.Name != "incident-response" {
		t.Errorf("Name = %q, want %q", prompt.Name, "incident-response")
	}

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application":  "myapp",
		"execution_id": "exec-123",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "exec-123") {
		t.Error("result should contain execution_id")
	}
	if !strings.Contains(text, "get_execution") {
		t.Error("result should reference get_execution tool")
	}
}

func TestIncidentResponse_WithoutExecutionID(t *testing.T) {
	_, handler := incidentResponse()

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "search_executions") {
		t.Error("result should reference search_executions when no execution_id")
	}
}

func TestPipelineAudit(t *testing.T) {
	prompt, handler := pipelineAudit()

	if prompt.Name != "pipeline-audit" {
		t.Errorf("Name = %q, want %q", prompt.Name, "pipeline-audit")
	}

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
		"pipeline":    "build-and-deploy",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "CRITICAL") {
		t.Error("result should mention severity ratings")
	}
}

func TestInfraOverview_WithAccount(t *testing.T) {
	prompt, handler := infraOverview()

	if prompt.Name != "infra-overview" {
		t.Errorf("Name = %q, want %q", prompt.Name, "infra-overview")
	}

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
		"account":     "prod-aws",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Description, "prod-aws") {
		t.Error("description should contain account name")
	}
}

func TestInfraOverview_WithoutAccount(t *testing.T) {
	_, handler := infraOverview()

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Description, "in account") {
		t.Error("description should not contain account filter when omitted")
	}
}

func TestRollbackPlan(t *testing.T) {
	prompt, handler := rollbackPlan()

	if prompt.Name != "rollback-plan" {
		t.Errorf("Name = %q, want %q", prompt.Name, "rollback-plan")
	}
	if len(prompt.Arguments) != 4 {
		t.Fatalf("expected 4 arguments, got %d", len(prompt.Arguments))
	}

	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"application": "myapp",
		"cluster":     "myapp-prod",
		"account":     "prod-aws",
		"region":      "us-east-1",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "myapp-prod") {
		t.Error("result should contain cluster name")
	}
	if !strings.Contains(text, "get_target_server_group") {
		t.Error("result should reference get_target_server_group tool")
	}
	if !strings.Contains(text, "ancestor") {
		t.Error("result should mention ancestor target for rollback")
	}
}

func TestRegister_DoesNotPanic(t *testing.T) {
	s := server.NewMCPServer("test", "0.0.0",
		server.WithPromptCapabilities(false),
	)
	Register(s)
}

func TestAllPromptNames(t *testing.T) {
	expected := map[string]bool{
		"deploy-review":    false,
		"incident-response": false,
		"pipeline-audit":   false,
		"infra-overview":   false,
		"rollback-plan":    false,
	}

	prompts := []func() (mcp.Prompt, server.PromptHandlerFunc){
		deployReview, incidentResponse, pipelineAudit, infraOverview, rollbackPlan,
	}

	for _, fn := range prompts {
		p, _ := fn()
		if _, ok := expected[p.Name]; !ok {
			t.Errorf("unexpected prompt: %s", p.Name)
		}
		expected[p.Name] = true
	}

	for name, found := range expected {
		if !found {
			t.Errorf("missing prompt: %s", name)
		}
	}
}
