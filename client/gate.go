package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const maxResponseSize = 10 << 20 // 10 MB

type GateClient struct {
	base string
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

func NewGate(base, token, user, pass, certFile, keyFile string, insecure bool) (*GateClient, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("invalid GATE_URL %q: %w", base, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("GATE_URL must use http or https scheme, got %q", u.Scheme)
	}

	transport := &http.Transport{}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("loading x509 cert/key: %w", err)
		}
		transport.TLSClientConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: insecure,
		}
	} else if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if insecure {
		log.Println("WARNING: TLS certificate verification is disabled (GATE_INSECURE=true)")
	}

	var auth authMethod
	switch {
	case token != "":
		auth = &bearerAuth{token: token}
	case user != "":
		auth = &basicAuth{user: user, pass: pass}
	default:
		auth = &noAuth{}
	}

	return &GateClient{
		base: base,
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

func (c *GateClient) buildURL(path string, q url.Values) (string, error) {
	u, err := url.Parse(c.base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func (c *GateClient) do(req *http.Request) ([]byte, error) {
	c.auth.apply(req)
	req.Header.Set("Accept", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		const maxErrBody = 500
		errMsg := string(body)
		if len(errMsg) > maxErrBody {
			errMsg = errMsg[:maxErrBody] + "... (truncated)"
		}
		return nil, fmt.Errorf("Gate error %d: %s", resp.StatusCode, errMsg)
	}
	return body, nil
}

func (c *GateClient) get(ctx context.Context, path string, q url.Values) ([]byte, error) {
	u, err := c.buildURL(path, q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	return c.do(req)
}

func (c *GateClient) post(ctx context.Context, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		body = bytes.NewReader(b)
	}
	u, err := c.buildURL(path, nil)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u, body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

func (c *GateClient) put(ctx context.Context, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		body = bytes.NewReader(b)
	}
	u, err := c.buildURL(path, nil)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", u, body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
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
	u, err := c.buildURL(fmt.Sprintf("/pipelines/%s/cancel", url.PathEscape(executionID)), q)
	if err != nil {
		return nil, err
	}
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

// --- Infrastructure ---

func (c *GateClient) ListServerGroups(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/serverGroups", url.PathEscape(app)), nil)
}

func (c *GateClient) ListLoadBalancers(ctx context.Context, app string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/applications/%s/loadBalancers", url.PathEscape(app)), nil)
}

// --- Tasks ---

func (c *GateClient) GetTask(ctx context.Context, taskID string) ([]byte, error) {
	return c.get(ctx, fmt.Sprintf("/tasks/%s", url.PathEscape(taskID)), nil)
}
