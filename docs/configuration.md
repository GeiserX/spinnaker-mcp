# Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GATE_URL` | `http://localhost:8084` | Spinnaker Gate API endpoint (without trailing /) |
| `GATE_TOKEN` | _(empty)_ | Bearer token for authentication |
| `GATE_USER` | _(empty)_ | Basic auth username (alternative to token) |
| `GATE_PASS` | _(empty)_ | Basic auth password |
| `GATE_CERT_FILE` | _(empty)_ | Path to x509 client certificate (PEM) |
| `GATE_KEY_FILE` | _(empty)_ | Path to x509 client key (PEM) |
| `GATE_INSECURE` | `false` | Skip TLS certificate verification |
| `TRANSPORT` | _(empty = HTTP)_ | Set to `stdio` for stdio transport |
| `MCP_PORT` | `8085` | HTTP transport port (ignored when TRANSPORT=stdio) |
| `MCP_BIND_ADDR` | `127.0.0.1` | HTTP transport bind address (set to `0.0.0.0` to listen on all interfaces) |

**Authentication priority**: Bearer token > Basic auth > x509 client cert > No auth.

Put them in a `.env` file (from `.env.example`) or set them in the environment.
