package main

import (
	"context"
	"log"
	"os"
	"strconv"
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
	gate, err := client.NewGate(client.GateOptions{
		BaseURL:  cfg.BaseURL,
		Token:    cfg.Token,
		User:     cfg.User,
		Pass:     cfg.Pass,
		CertFile: cfg.CertFile,
		KeyFile:  cfg.KeyFile,
		Insecure: cfg.Insecure,
	})
	if err != nil {
		log.Fatalf("Failed to create Gate client: %v", err)
	}

	s := server.NewMCPServer(
		"Spinnaker MCP",
		version.Version,
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

	tool, handler = tools.NewSavePipeline(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewUpdatePipeline(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewDeletePipeline(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetPipelineHistory(gate)
	s.AddTool(tool, handler)

	// Executions
	tool, handler = tools.NewGetExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListExecutions(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewSearchExecutions(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewCancelExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewPauseExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewResumeExecution(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewRestartStage(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewEvaluateExpression(gate)
	s.AddTool(tool, handler)

	// Strategies
	tool, handler = tools.NewListStrategies(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewSaveStrategy(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewDeleteStrategy(gate)
	s.AddTool(tool, handler)

	// Infrastructure
	tool, handler = tools.NewListServerGroups(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListLoadBalancers(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListClusters(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetCluster(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetScalingActivities(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetTargetServerGroup(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListFirewalls(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetFirewall(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetInstance(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetConsoleOutput(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewFindImages(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetImageTags(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListNetworks(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListSubnets(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewListAccounts(gate)
	s.AddTool(tool, handler)

	tool, handler = tools.NewGetAccount(gate)
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
		portStr := os.Getenv("MCP_PORT")
		if portStr == "" {
			portStr = "8085"
		}
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			log.Fatalf("Invalid MCP_PORT %q: must be 1-65535", portStr)
		}
		httpSrv := server.NewStreamableHTTPServer(s)
		log.Printf("Spinnaker MCP listening on :%s", portStr)
		if err := httpSrv.Start(":" + portStr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
