package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseSize int64 = 10 << 20 // 10 MB

type GateClient struct {
	base *url.URL
	hc   *http.Client
	auth authMethod
}

type authMethod interface {
	apply(req *http.Request)
}

type bearerAuth struct{ token string }

func (a *bearerAuth) apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.token)
}

type basicAuth struct{ user, pass string }

func (a *basicAuth) apply(req *http.Request) {
	req.SetBasicAuth(a.user, a.pass)
}

type noAuth struct{}

func (a *noAuth) apply(_ *http.Request) {}

// cookieAuth sends a browser session cookie. It is how a Gate behind SSO (OAuth2, SAML)
// authenticates API calls: the user signs in once in a browser and the SESSION cookie
// Gate sets there is replayed on every request.
type cookieAuth struct{ cookie string }

func (a *cookieAuth) apply(req *http.Request) {
	req.Header.Set("Cookie", a.cookie)
}

// ErrLoginRequired is returned when Gate answers an API call with a redirect to its login
// page, which is how a Gate behind SSO reports a missing or expired session.
var ErrLoginRequired = errors.New("gate requires a login session: set GATE_COOKIE or GATE_COOKIE_FILE to a valid SESSION cookie, or run 'spinnaker-mcp login'")

// NormalizeCookie accepts either a full Cookie header value ("SESSION=abc") or a bare
// SESSION value ("abc") and returns the header value to send.
func NormalizeCookie(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	if !strings.Contains(v, "=") {
		return "SESSION=" + v
	}
	return v
}

func isLoginRedirect(location string) bool {
	u, err := url.Parse(location)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/login")
}

type GateOptions struct {
	BaseURL  string
	Token    string
	User     string
	Pass     string
	CertFile string
	KeyFile  string
	Insecure bool
	// Cookie is a Cookie header value (or bare SESSION value) for a Gate behind SSO.
	Cookie string
}

func NewGate(opts GateOptions) (*GateClient, error) {
	u, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GATE_URL %q: %w", opts.BaseURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("GATE_URL must use http or https scheme, got %q", u.Scheme)
	}
	u.Path = strings.TrimRight(u.Path, "/")

	transport := http.DefaultTransport.(*http.Transport).Clone()

	if opts.CertFile != "" && opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading x509 cert/key: %w", err)
		}
		transport.TLSClientConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: opts.Insecure,
			MinVersion:         tls.VersionTLS12,
		}
	} else if opts.Insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
	}

	if opts.Insecure && u.Scheme == "https" {
		log.Println("WARNING: TLS certificate verification is disabled (GATE_INSECURE=true)")
	}

	var auth authMethod
	switch {
	case opts.Token != "":
		auth = &bearerAuth{token: opts.Token}
	case opts.User != "":
		auth = &basicAuth{user: opts.User, pass: opts.Pass}
	case strings.TrimSpace(opts.Cookie) != "":
		auth = &cookieAuth{cookie: NormalizeCookie(opts.Cookie)}
	default:
		auth = &noAuth{}
	}

	return &GateClient{
		base: u,
		hc: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		auth: auth,
	}, nil
}

func (c *GateClient) buildURL(path string, q url.Values) string {
	u := *c.base
	u.Path += path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func (c *GateClient) do(req *http.Request) ([]byte, error) {
	c.auth.apply(req)
	req.Header.Set("Accept", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if int64(len(body)) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds %d bytes", maxResponseSize)
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		if isLoginRedirect(loc) {
			return nil, fmt.Errorf("%w (gate redirected to %q)", ErrLoginRequired, loc)
		}
		return nil, fmt.Errorf("gate returned unexpected redirect %d to %q", resp.StatusCode, loc)
	}
	if resp.StatusCode >= 400 {
		const maxErrBody = 500
		errMsg := string(body)
		if len(errMsg) > maxErrBody {
			errMsg = errMsg[:maxErrBody] + "... (truncated)"
		}
		return nil, fmt.Errorf("gate error %d: %s", resp.StatusCode, errMsg)
	}
	return body, nil
}

func (c *GateClient) get(ctx context.Context, path string, q url.Values) ([]byte, error) {
	u := c.buildURL(path, q)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	return c.do(req)
}

func (c *GateClient) doWithBody(ctx context.Context, method, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		body = bytes.NewReader(b)
	}
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

func (c *GateClient) post(ctx context.Context, path string, payload any) ([]byte, error) {
	return c.doWithBody(ctx, "POST", path, payload)
}

func (c *GateClient) put(ctx context.Context, path string, payload any) ([]byte, error) {
	return c.doWithBody(ctx, "PUT", path, payload)
}

func (c *GateClient) del(ctx context.Context, path string) ([]byte, error) {
	return c.doWithBody(ctx, "DELETE", path, nil)
}

// --- Applications ---

func (c *GateClient) ListApplications(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/applications", nil)
}

func (c *GateClient) GetApplication(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s", url.PathEscape(app)), nil)
}

// --- Pipelines ---

func (c *GateClient) ListPipelines(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/pipelineConfigs", url.PathEscape(app)), nil)
}

func (c *GateClient) GetPipelineConfig(ctx context.Context, app, pipelineName string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/pipelineConfigs/%s",
		url.PathEscape(app), url.PathEscape(pipelineName)), nil)
}

func (c *GateClient) TriggerPipeline(ctx context.Context, app, pipelineName string, parameters map[string]any) ([]byte, error) {
	body := map[string]any{"type": "manual"}
	if len(parameters) > 0 {
		body["parameters"] = parameters
	}
	return c.post(ctx, fmt.Sprintf("/pipelines/v2/%s/%s",
		url.PathEscape(app), url.PathEscape(pipelineName)), body)
}

func (c *GateClient) SavePipeline(ctx context.Context, pipeline map[string]any) ([]byte, error) {
	return c.post(ctx, "/pipelines", pipeline)
}

func (c *GateClient) UpdatePipeline(ctx context.Context, pipelineID string, pipeline map[string]any) ([]byte, error) {
	return c.put(ctx, fmt.Sprintf("/pipelines/%s", url.PathEscape(pipelineID)), pipeline)
}

func (c *GateClient) DeletePipeline(ctx context.Context, app, pipelineName string) ([]byte, error) {
	return c.del(ctx, fmt.Sprintf("/pipelines/%s/%s",
		url.PathEscape(app), url.PathEscape(pipelineName)))
}

func (c *GateClient) GetPipelineHistory(ctx context.Context, pipelineConfigID string, limit int) ([]byte, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	return c.get(ctx, fmt.Sprintf("/pipelineConfigs/%s/history", url.PathEscape(pipelineConfigID)), q)
}

// --- Executions ---

func (c *GateClient) GetExecution(ctx context.Context, executionID string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/pipelines/%s", url.PathEscape(executionID)), nil)
}

func (c *GateClient) ListExecutions(ctx context.Context, app string, limit int, statuses string) ([]byte, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if statuses != "" {
		q.Set("statuses", statuses)
	}
	return c.get(ctx, fmt.Sprintf("/applications/%s/pipelines", url.PathEscape(app)), q)
}

func (c *GateClient) CancelExecution(ctx context.Context, executionID, reason string) ([]byte, error) {
	q := url.Values{}
	if reason != "" {
		q = url.Values{"reason": {reason}}
	}
	u := c.buildURL(fmt.Sprintf("/pipelines/%s/cancel", url.PathEscape(executionID)), q)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	return c.do(req)
}

func (c *GateClient) PauseExecution(ctx context.Context, executionID string) ([]byte, error) {
	return c.put(ctx, fmt.Sprintf("/pipelines/%s/pause", url.PathEscape(executionID)), nil)
}

func (c *GateClient) ResumeExecution(ctx context.Context, executionID string) ([]byte, error) {
	return c.put(ctx, fmt.Sprintf("/pipelines/%s/resume", url.PathEscape(executionID)), nil)
}

func (c *GateClient) RestartStage(ctx context.Context, executionID, stageID string) ([]byte, error) {
	return c.put(ctx, fmt.Sprintf("/pipelines/%s/stages/%s/restart",
		url.PathEscape(executionID), url.PathEscape(stageID)), map[string]any{})
}

func (c *GateClient) SearchExecutions(ctx context.Context, app string, params map[string]string) ([]byte, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	return c.get(ctx, fmt.Sprintf("/applications/%s/executions/search", url.PathEscape(app)), q)
}

func (c *GateClient) EvaluateExpression(ctx context.Context, executionID, expression string) ([]byte, error) {
	return c.post(ctx, fmt.Sprintf("/pipelines/%s/evaluateExpression",
		url.PathEscape(executionID)), map[string]any{"expression": expression})
}

// --- Strategies ---

func (c *GateClient) ListStrategies(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/strategyConfigs", url.PathEscape(app)), nil)
}

func (c *GateClient) SaveStrategy(ctx context.Context, strategy map[string]any) ([]byte, error) {
	return c.post(ctx, "/strategies", strategy)
}

func (c *GateClient) DeleteStrategy(ctx context.Context, app, strategyName string) ([]byte, error) {
	return c.del(ctx, fmt.Sprintf("/strategies/%s/%s",
		url.PathEscape(app), url.PathEscape(strategyName)))
}

// --- Infrastructure ---

func (c *GateClient) ListServerGroups(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/serverGroups", url.PathEscape(app)), nil)
}

func (c *GateClient) ListLoadBalancers(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/loadBalancers", url.PathEscape(app)), nil)
}

// --- Clusters ---

func (c *GateClient) ListClusters(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/clusters", url.PathEscape(app)), nil)
}

func (c *GateClient) GetCluster(ctx context.Context, app, account, cluster string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/clusters/%s/%s",
		url.PathEscape(app), url.PathEscape(account), url.PathEscape(cluster)), nil)
}

func (c *GateClient) GetScalingActivities(ctx context.Context, app, account, cluster, serverGroupName, provider string) ([]byte, error) {
	q := url.Values{}
	if provider != "" {
		q.Set("provider", provider)
	}
	return c.get(ctx, fmt.Sprintf("/applications/%s/clusters/%s/%s/serverGroups/%s/scalingActivities",
		url.PathEscape(app), url.PathEscape(account), url.PathEscape(cluster), url.PathEscape(serverGroupName)), q)
}

func (c *GateClient) GetTargetServerGroup(ctx context.Context, app, account, cluster, cloudProvider, scope, target string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/clusters/%s/%s/%s/%s/serverGroups/target/%s",
		url.PathEscape(app), url.PathEscape(account), url.PathEscape(cluster),
		url.PathEscape(cloudProvider), url.PathEscape(scope), url.PathEscape(target)), nil)
}

// --- Security Groups / Firewalls ---

func (c *GateClient) ListFirewalls(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/securityGroups", nil)
}

func (c *GateClient) GetFirewall(ctx context.Context, account, region, name string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/securityGroups/%s/%s/%s",
		url.PathEscape(account), url.PathEscape(region), url.PathEscape(name)), nil)
}

// --- Instances ---

func (c *GateClient) GetInstance(ctx context.Context, account, region, instanceID string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/instances/%s/%s/%s",
		url.PathEscape(account), url.PathEscape(region), url.PathEscape(instanceID)), nil)
}

func (c *GateClient) GetConsoleOutput(ctx context.Context, account, region, instanceID, provider string) ([]byte, error) {
	q := url.Values{}
	if provider != "" {
		q.Set("provider", provider)
	}
	return c.get(ctx, fmt.Sprintf("/instances/%s/%s/%s/console",
		url.PathEscape(account), url.PathEscape(region), url.PathEscape(instanceID)), q)
}

// --- Images ---

func (c *GateClient) FindImages(ctx context.Context, params map[string]string) ([]byte, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	return c.get(ctx, "/images/find", q)
}

func (c *GateClient) GetImageTags(ctx context.Context, account, repository string) ([]byte, error) {
	q := url.Values{}
	q.Set("account", account)
	q.Set("repository", repository)
	return c.get(ctx, "/images/tags", q)
}

// --- Networks and Subnets ---

func (c *GateClient) ListNetworks(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/networks", nil)
}

func (c *GateClient) ListSubnets(ctx context.Context, cloudProvider string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/subnets/%s", url.PathEscape(cloudProvider)), nil)
}

// --- Credentials / Accounts ---

func (c *GateClient) ListAccounts(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/credentials", nil)
}

func (c *GateClient) GetAccount(ctx context.Context, account string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/credentials/%s", url.PathEscape(account)), nil)
}

// --- Tasks ---

func (c *GateClient) GetTask(ctx context.Context, taskID string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/tasks/%s", url.PathEscape(taskID)), nil)
}

// --- Auth ---

// CurrentUser returns Gate's view of the caller (GET /auth/user). Behind SSO it is the
// cheapest check that a session cookie is still valid; Gate answers an expired or
// missing session with a redirect to /login, which surfaces as ErrLoginRequired.
func (c *GateClient) CurrentUser(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/auth/user", nil)
}

// --- Health ---

// Ping checks if the Gate API is reachable by issuing a HEAD request to the base URL.
func (c *GateClient) Ping(ctx context.Context) error {
	u := c.buildURL("/applications", nil)
	req, err := http.NewRequestWithContext(ctx, "HEAD", u, nil)
	if err != nil {
		return fmt.Errorf("building health check request: %w", err)
	}
	c.auth.apply(req)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("gate returned %d", resp.StatusCode)
	}
	return nil
}

// BaseURL returns the configured Gate API base URL string.
func (c *GateClient) BaseURL() string {
	return c.base.String()
}
