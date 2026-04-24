package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
	"github.com/geiserx/spinnaker-mcp/internal/prompts"
	"github.com/geiserx/spinnaker-mcp/internal/resources"
	"github.com/geiserx/spinnaker-mcp/internal/toolsets"
	"github.com/geiserx/spinnaker-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var toolsetsFlag string

	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Println("spinnaker-mcp " + version.String())
			return
		case arg == "--help" || arg == "-h":
			printHelp()
			return
		case strings.HasPrefix(arg, "--toolsets="):
			toolsetsFlag = strings.TrimPrefix(arg, "--toolsets=")
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\nRun 'spinnaker-mcp --help' for usage.\n", arg)
				os.Exit(1)
			}
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

	// Resolve toolsets
	if toolsetsFlag == "" {
		toolsetsFlag = os.Getenv("TOOLSETS")
	}

	allTools := toolsets.BuildTools(gate)
	selectedTools, err := toolsets.Resolve(toolsetsFlag, allTools)
	if err != nil {
		log.Fatalf("Invalid toolsets: %v", err)
	}

	s := server.NewMCPServer(
		"Spinnaker MCP",
		version.Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithRecovery(),
	)

	// Register selected tools
	for _, entry := range selectedTools {
		s.AddTool(entry.Tool, entry.Handler)
	}
	log.Printf("Registered %d tools (toolsets: %s)", len(selectedTools), toolsetsLabel(toolsetsFlag))

	// Register resources and prompts (always, regardless of toolsets)
	resources.Register(s, gate)
	prompts.Register(s)

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

		mux := http.NewServeMux()
		mux.Handle("/mcp", httpSrv)
		mux.HandleFunc("/healthz", healthzHandler)
		mux.HandleFunc("/readyz", readyzHandler(gate))

		addr := bindAddr + ":" + portStr
		log.Printf("Spinnaker MCP listening on %s (health: /healthz, /readyz)", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

func toolsetsLabel(raw string) string {
	if raw == "" {
		return "all"
	}
	return raw
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": version.Version,
	})
}

func readyzHandler(gate *client.GateClient) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := gate.Ping(ctx)
		reachable := err == nil

		status := "ready"
		code := http.StatusOK
		if !reachable {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}

		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"status":         status,
			"gate_url":       gate.BaseURL(),
			"gate_reachable": reachable,
		})
	}
}

func printHelp() {
	fmt.Println("spinnaker-mcp " + version.String())
	fmt.Println()
	fmt.Println("Usage: spinnaker-mcp [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v          Print version and exit")
	fmt.Println("  --help, -h             Show this help message")
	fmt.Println("  --toolsets=GROUPS      Comma-separated tool groups to enable (default: all)")
	fmt.Println()
	fmt.Println("Toolset groups:")
	fmt.Println("  applications           Application listing and details (2 tools)")
	fmt.Println("  pipelines              Pipeline CRUD, trigger, history (7 tools)")
	fmt.Println("  executions             Execution lifecycle and search (8 tools)")
	fmt.Println("  strategies             Deployment strategy management (3 tools)")
	fmt.Println("  infrastructure         Server groups, clusters, LBs, firewalls, images, networks, accounts (16 tools)")
	fmt.Println("  tasks                  Task status (1 tool)")
	fmt.Println()
	fmt.Println("Meta-groups:")
	fmt.Println("  all                    All tool groups (default)")
	fmt.Println("  readonly               Only read-only tools (24 tools)")
	fmt.Println("  mutating               Only mutating and destructive tools (13 tools)")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  GATE_URL               Spinnaker Gate API URL (default: http://localhost:8084)")
	fmt.Println("  GATE_TOKEN             Bearer token for Gate authentication")
	fmt.Println("  GATE_USER              Basic auth username (alternative to token)")
	fmt.Println("  GATE_PASS              Basic auth password")
	fmt.Println("  GATE_CERT_FILE         Path to x509 client certificate (PEM)")
	fmt.Println("  GATE_KEY_FILE          Path to x509 client key (PEM)")
	fmt.Println("  GATE_INSECURE          Skip TLS certificate verification (default: false)")
	fmt.Println("  TRANSPORT              'stdio' for stdio transport (default: HTTP)")
	fmt.Println("  MCP_PORT               HTTP listen port (default: 8085)")
	fmt.Println("  MCP_BIND_ADDR          HTTP bind address (default: 127.0.0.1)")
	fmt.Println("  TOOLSETS               Same as --toolsets flag (flag takes precedence)")
}
