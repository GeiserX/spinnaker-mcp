package main

import (
	"context"
	"fmt"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("spinnaker-mcp " + version.String())
			return
		case "--help", "-h":
			fmt.Println("spinnaker-mcp " + version.String())
			fmt.Println()
			fmt.Println("Usage: spinnaker-mcp [flags]")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  --version, -v  Print version and exit")
			fmt.Println("  --help, -h     Show this help message")
			fmt.Println()
			fmt.Println("Environment variables:")
			fmt.Println("  GATE_URL       Spinnaker Gate API URL (default: http://localhost:8084)")
			fmt.Println("  GATE_TOKEN     Bearer token for Gate authentication")
			fmt.Println("  GATE_USER      Basic auth username (alternative to token)")
			fmt.Println("  GATE_PASS      Basic auth password")
			fmt.Println("  TRANSPORT      'stdio' for stdio transport (default: HTTP)")
			fmt.Println("  MCP_PORT       HTTP listen port (default: 8085)")
			fmt.Println("  MCP_BIND_ADDR  HTTP bind address (default: 127.0.0.1)")
			return
		}
	}

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
		bindAddr := os.Getenv("MCP_BIND_ADDR")
		if bindAddr == "" {
			bindAddr = "127.0.0.1"
		}
		httpSrv := server.NewStreamableHTTPServer(s)
		log.Printf("Spinnaker MCP listening on %s:%s", bindAddr, portStr)
		if err := httpSrv.Start(bindAddr + ":" + portStr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
