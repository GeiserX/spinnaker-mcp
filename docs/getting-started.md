# Getting started

Needs a reachable Spinnaker Gate URL and, for the npm package, Node 18 or newer.

## npm (stdio transport)

```sh
npx spinnaker-mcp
```

Or install globally:

```sh
npm install -g spinnaker-mcp
spinnaker-mcp
```

The first run downloads the pre-built Go binary for your platform from the GitHub release, refuses it unless its SHA-256 matches the `checksums.txt` shipped in the package, and runs it with stdio transport. Install scripts are not needed, so it also works with `--ignore-scripts`.

## Docker

```sh
docker run -i --rm -e GATE_URL=http://spin-gate:8084 -e TRANSPORT=stdio drumsergio/spinnaker-mcp:v0.4.0
```

Image tags carry the `v` prefix (`v0.4.0`). `-i` keeps stdin open, which the stdio transport needs.

## Local build

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
cd spinnaker-mcp

cp .env.example .env && $EDITOR .env

SPINNAKER_MCP_ENV_FILE=.env go run ./cmd/server
```

## Kubernetes (Helm)

The chart is in the repository, not in a Helm repository, so install it from a checkout:

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
kubectl create secret generic spinnaker-mcp-auth --from-literal=MCP_AUTH_TOKEN="$(openssl rand -hex 32)"
helm install spinnaker-mcp ./spinnaker-mcp/helm/spinnaker-mcp \
  --set gate.url=http://spin-gate:8084 \
  --set gate.auth.existingSecret=spinnaker-mcp-gate \
  --set mcpAuth.existingSecret=spinnaker-mcp-auth
```

`spinnaker-mcp-gate` is a Secret you create with a `GATE_TOKEN` key, or `GATE_USER` and `GATE_PASS`. Without
it, `gate.auth.type` (`token` or `basic`) with `gate.auth.token` or `gate.auth.user` and `gate.auth.pass`
makes the chart create one. For an x509 client certificate, `gate.tls.certSecret` names a Secret with
`tls.crt` and `tls.key`. The pod runs the [HTTP transport](configuration.md#http-transport) on port 8085
behind a `ClusterIP` Service, with liveness on `/healthz` and readiness on `/readyz`, and the image tag
defaults to `v` plus the chart's `appVersion` (`v0.4.0`). It listens on `0.0.0.0` inside the pod, so the chart
refuses to render without `mcpAuth.existingSecret` (a Secret with an `MCP_AUTH_TOKEN` key) or
`mcpAuth.token`: every `/mcp` request must carry that token as a bearer token. The chart registers only the
read tools; set `toolsets=all` only if the clients must change things, and still limit who reaches the Service.

## Claude Code / Claude Desktop configuration

```json
{
  "mcpServers": {
    "spinnaker": {
      "command": "npx",
      "args": ["-y", "spinnaker-mcp"],
      "env": {
        "GATE_URL": "https://spin-gate.example.com",
        "GATE_TOKEN": "your-token-here"
      }
    }
  }
}
```

See [Configuration](configuration.md) for the environment variables. Behind a single sign-on login Gate takes no
token: leave `GATE_TOKEN` out and run `npx -y spinnaker-mcp login --gate=https://spin-gate.example.com` once, see
[Gate behind SSO](configuration.md#gate-behind-sso).
