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

This downloads the pre-built Go binary for your platform and runs it with stdio transport.

## Docker

```sh
docker run -i --rm -e GATE_URL=http://spin-gate:8084 -e TRANSPORT=stdio drumsergio/spinnaker-mcp:v0.3.4
```

Image tags carry the `v` prefix (`v0.3.4`). `-i` keeps stdin open, which the stdio transport needs.

## Local build

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
cd spinnaker-mcp

cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

## Kubernetes (Helm)

The chart is in the repository, not in a Helm repository, so install it from a checkout:

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
helm install spinnaker-mcp ./spinnaker-mcp/helm/spinnaker-mcp \
  --set gate.url=http://spin-gate:8084 \
  --set gate.auth.existingSecret=spinnaker-mcp-gate \
  --set toolsets=readonly
```

`spinnaker-mcp-gate` is a Secret you create with a `GATE_TOKEN` key, or `GATE_USER` and `GATE_PASS`. Without
it, `gate.auth.type` (`token` or `basic`) with `gate.auth.token` or `gate.auth.user` and `gate.auth.pass`
makes the chart create one. For an x509 client certificate, `gate.tls.certSecret` names a Secret with
`tls.crt` and `tls.key`. The pod runs the [HTTP transport](configuration.md#http-transport) on port 8085
behind a `ClusterIP` Service, with liveness on `/healthz` and readiness on `/readyz`, and the image tag
defaults to `v` plus the chart's `appVersion` (`v0.3.4`). It listens on `0.0.0.0` inside the pod and `/mcp`
has no login, so anything that reaches the Service acts with the Gate identity: drop `toolsets=readonly`
only if the clients must change things, and limit who reaches the Service.

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
