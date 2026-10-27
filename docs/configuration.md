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
| `GATE_COOKIE` | _(empty)_ | Session cookie for a Gate behind SSO, `SESSION=...` or the bare value, see [Gate behind SSO](#gate-behind-sso) |
| `GATE_COOKIE_FILE` | _(per host)_ | File holding that cookie; default `$XDG_CONFIG_HOME/spinnaker-mcp/cookies/<host>`, which `spinnaker-mcp login` writes |
| `TRANSPORT` | _(empty = HTTP)_ | Set to `stdio` for stdio transport |
| `MCP_PORT` | `8085` | HTTP transport port (ignored when TRANSPORT=stdio) |
| `MCP_BIND_ADDR` | `127.0.0.1` | HTTP transport bind address; anything but loopback needs `MCP_AUTH_TOKEN` |
| `MCP_AUTH_TOKEN` | _(empty)_ | Bearer token every `/mcp` request must carry, see [HTTP transport](#http-transport) |
| `MCP_ALLOWED_ORIGINS` | _(empty)_ | Browser origins allowed to call `/mcp` cross-origin, comma-separated (`https://host[:port]`) |
| `TOOLSETS` | `readonly` | Tool groups to register, see [Toolsets](#toolsets) |
| `SPINNAKER_MCP_ENV_FILE` | _(empty)_ | A dotenv file to load at start |

**Authentication priority**: Bearer token > Basic auth > Session cookie > No auth. An x509 client certificate
(`GATE_CERT_FILE` and `GATE_KEY_FILE`) is presented on top of whichever of those applies.

Set them in the environment, or put them in a file (from `.env.example`) and name it in
`SPINNAKER_MCP_ENV_FILE`. The server never reads a `.env` from the working directory on its own: MCP clients
start it inside whatever project is open, and a stray `.env` there could redirect it to another Gate. The npm package always runs
with `TRANSPORT=stdio`, and the Docker image sets `TRANSPORT=stdio` unless you override it.

A value that is nothing but an unexpanded placeholder, such as `${GATE_TOKEN}`, counts as unset. Some MCP
clients write that placeholder when an optional field is left blank; the server ignores it rather than send it
to Gate as a token or reject it as a toolset.

## Gate behind SSO

A Gate that sits behind a single sign-on login (OAuth2, SAML) accepts no token and no password from an API
client: every call without its browser session is answered with a redirect to `/login`, and the server
reports it as `gate requires a login session`. What Gate does accept is the `SESSION` cookie it sets in the
browser after you sign in, so the server replays that cookie.

Sign in once per Gate:

```sh
spinnaker-mcp login --gate=https://spin-gate.example.com      # or: npx -y spinnaker-mcp login --gate=...
```

It opens the Gate login page in your browser. Sign in, then copy the `SESSION` cookie for the Gate host
(browser DevTools, Application tab, Cookies) and paste it at the prompt. The command checks the cookie
against `GET /auth/user`, prints the user it belongs to, and stores it with mode `0600` in
`$XDG_CONFIG_HOME/spinnaker-mcp/cookies/<host>` (`~/.config/...` by default). From then on the server
picks it up on its own whenever `GATE_URL` points at that host, so the client configuration needs only
`GATE_URL`. When the session expires the tools fail with the same `gate requires a login session` message:
run `login` again.

The prompt does not echo what you paste. `--cookie=VALUE` skips the prompt but leaves the cookie in your
shell history and the process list, so prefer the prompt or piping it on stdin; `--cookie-file=PATH` stores it elsewhere, `--no-browser` only prints
the URL. `GATE_COOKIE` and `GATE_COOKIE_FILE` set the cookie directly for a server that cannot run
`login`, for example in a container. The cookie is a credential: whoever holds it is you on that Gate
until it expires, so keep it out of shared configuration and version control.

## Toolsets

`TOOLSETS`, or the `--toolsets=` flag (the flag wins), chooses which tools the server registers. A tool that
is not registered does not exist for the client. Resources and prompts are registered whatever it says.

| Value | Tools |
|-------|-------|
| `readonly` (default) | the 27 that only read |
| `all` | all 38 |
| `mutating` | the 11 that change Spinnaker |
| `applications` | 2 |
| `pipelines` | 7 |
| `executions` | 9 |
| `strategies` | 3 |
| `infrastructure` | 16 |
| `tasks` | 1 |

Groups combine with commas: `TOOLSETS=pipelines,executions`. `all`, `readonly` and `mutating` stand alone;
combined with each other or with a group, the server exits at start with `Invalid toolsets`.

With no `TOOLSETS` the server is read-only: an assistant cannot trigger, cancel, edit, delete or judge
anything, even if text inside a Gate response tells it to. To let it, add `"TOOLSETS": "all"` (or the groups
it needs) to the `env` block of your client configuration. [Usage](usage.md#tools) lists which tools change
Spinnaker.

## HTTP transport

When `TRANSPORT` is anything but `stdio`, the server listens on `MCP_BIND_ADDR:MCP_PORT`
(`127.0.0.1:8085` by default) and serves:

| Path | Answers |
|------|---------|
| `/mcp` | MCP over stateless Streamable HTTP, both protocol eras |
| `/healthz` | `200` with `{"status":"ok","version":"..."}` |
| `/readyz` | `200` when Gate answers a `HEAD /applications` within 5 s with any status below 500, otherwise `503`; the body is `{"status":...,"gate_reachable":...}` |

`/readyz` proves Gate is reachable, not that the credentials work: a `401` from Gate still counts as ready.

Anyone who reaches `/mcp` acts with the Gate credentials the server holds, so it is guarded:

- **Bearer token.** With `MCP_AUTH_TOKEN` set, every `/mcp` request must send
  `Authorization: Bearer <token>`; anything else gets `401`. The server refuses to start on an address
  that is not loopback without it. Generate one with `openssl rand -hex 32`. `/healthz` and `/readyz` stay
  open for probes and reveal nothing about Gate.
- **Origin.** A browser request from another site gets `403`, so a web page cannot drive a server on your
  machine. Clients that are not browsers send no `Origin` and are unaffected. `MCP_ALLOWED_ORIGINS` lets
  named browser origins through.
- **Size and time.** Request bodies are capped at 4 MiB; headers must arrive within 10 s.

In Docker, `127.0.0.1` is the container's own loopback, so the HTTP transport needs `0.0.0.0` and therefore
a token: `-e TRANSPORT=http -e MCP_BIND_ADDR=0.0.0.0 -e MCP_AUTH_TOKEN -p 127.0.0.1:8085:8085`.
