package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
	"github.com/geiserx/spinnaker-mcp/internal/tools"
	"github.com/geiserx/spinnaker-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.Printf("Spinnaker MCP %s starting…", version.String())

	cfg := config.LoadGateConfig()
	gate := client.NewGate(cfg.BaseURL, cfg.Token, cfg.User, cfg.Pass, cfg.CertFile, cfg.KeyFile, cfg.Insecure)

	s := server.NewMCPServer(
		"Spinnaker MCP",
		"0.0.1",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Applications
	tool, handler := tools.NewListApplications(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetApplication(gate)
	s.AddTool(tool, handler)

	// Pipelines
	tool, handler = tools.NewListPipelines(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetPipeline(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewTriggerPipeline(gate)
	s.AddTool(tool, handler)

	// Executions
	tool, handler = tools.NewGetExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListExecutions(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewCancelExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewPauseExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewResumeExecution(gate)
	s.AddTool(tool, handler)

	// Infrastructure
	tool, handler = tools.NewListServerGroups(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListLoadBalancers(gate)
	s.AddTool(tool, handler)

	// Tasks
	tool, handler = tools.NewGetTask(gate)
	s.AddTool(tool, handler)

	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport == "stdio" {
		stdioSrv := server.NewStdioServer(s)
		log.Println("Spinnaker MCP running on stdio")
		if err := stdioSrv.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	} else {
		httpSrv := server.NewStreamableHTTPServer(s)
		log.Println("Spinnaker MCP listening on :8084")
		if err := httpSrv.Start(":8084"); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
