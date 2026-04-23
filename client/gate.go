package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

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

func NewGate(base, token, user, pass, certFile, keyFile string, insecure bool) *GateClient {
	transport := &http.Transport{}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err == nil {
			transport.TLSClientConfig = &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: insecure,
			}
		}
	} else if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
		hc:   &http.Client{Transport: transport},
		auth: auth,
	}
}

func (c *GateClient) buildURL(path string, q url.Values) string {
	u, _ := url.Parse(c.base)
	u.Path = path
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
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Gate error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *GateClient) get(path string, q url.Values) ([]byte, error) {
	req, _ := http.NewRequest("GET", c.buildURL(path, q), nil)
	return c.do(req)
}

func (c *GateClient) post(path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequest("POST", c.buildURL(path, nil), body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

func (c *GateClient) put(path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequest("PUT", c.buildURL(path, nil), body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// --- Applications ---

func (c *GateClient) ListApplications() ([]byte, error) {
	return c.get("/applications", nil)
}

func (c *GateClient) GetApplication(app string) ([]byte, error) {
	return c.get(fmt.Sprintf("/applications/%s", url.PathEscape(app)), nil)
}

// --- Pipelines ---

func (c *GateClient) ListPipelines(app string) ([]byte, error) {
	return c.get(fmt.Sprintf("/applications/%s/pipelineConfigs", url.PathEscape(app)), nil)
}

func (c *GateClient) GetPipelineConfig(app, pipelineName string) ([]byte, error) {
	return c.get(fmt.Sprintf("/applications/%s/pipelineConfigs/%s",
		url.PathEscape(app), url.PathEscape(pipelineName)), nil)
}

func (c *GateClient) TriggerPipeline(app, pipelineName string, parameters map[string]any) ([]byte, error) {
	body := map[string]any{"type": "manual"}
	if len(parameters) > 0 {
		body["parameters"] = parameters
	}
	return c.post(fmt.Sprintf("/pipelines/v2/%s/%s",
		url.PathEscape(app), url.PathEscape(pipelineName)), body)
}

// --- Executions ---

func (c *GateClient) GetExecution(executionID string) ([]byte, error) {
	return c.get(fmt.Sprintf("/pipelines/%s", url.PathEscape(executionID)), nil)
}

func (c *GateClient) ListExecutions(app string, limit int, statuses string) ([]byte, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if statuses != "" {
		q.Set("statuses", statuses)
	}
	return c.get(fmt.Sprintf("/applications/%s/pipelines", url.PathEscape(app)), q)
}

func (c *GateClient) CancelExecution(executionID, reason string) ([]byte, error) {
	q := url.Values{}
	if reason != "" {
		q = url.Values{"reason": {reason}}
	}
	req, _ := http.NewRequest("PUT", c.buildURL(
		fmt.Sprintf("/pipelines/%s/cancel", url.PathEscape(executionID)), q), nil)
	return c.do(req)
}

func (c *GateClient) PauseExecution(executionID string) ([]byte, error) {
	return c.put(fmt.Sprintf("/pipelines/%s/pause", url.PathEscape(executionID)), nil)
}

func (c *GateClient) ResumeExecution(executionID string) ([]byte, error) {
	return c.put(fmt.Sprintf("/pipelines/%s/resume", url.PathEscape(executionID)), nil)
}

// --- Infrastructure ---

func (c *GateClient) ListServerGroups(app string) ([]byte, error) {
	return c.get(fmt.Sprintf("/applications/%s/serverGroups", url.PathEscape(app)), nil)
}

func (c *GateClient) ListLoadBalancers(app string) ([]byte, error) {
	return c.get(fmt.Sprintf("/applications/%s/loadBalancers", url.PathEscape(app)), nil)
}

// --- Tasks ---

func (c *GateClient) GetTask(taskID string) ([]byte, error) {
	return c.get(fmt.Sprintf("/tasks/%s", url.PathEscape(taskID)), nil)
}
