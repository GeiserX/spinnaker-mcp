package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGateClient_Ping_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD, got %s", r.Method)
		}
		if r.URL.Path != "/applications" {
			t.Errorf("expected /applications, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	if err := gate.Ping(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_Ping_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	err := gate.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGateClient_Ping_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	gate := newTestGate(t, url)
	err := gate.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestGateClient_Ping_401Succeeds(t *testing.T) {
	// 401 is < 500, so Ping should succeed (Gate is reachable even if auth fails)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()

	gate := newTestGate(t, srv.URL)
	if err := gate.Ping(context.Background()); err != nil {
		t.Fatalf("Ping should succeed on 401 (gate is reachable): %v", err)
	}
}

func TestGateClient_Ping_WithAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-token" {
			t.Errorf("expected Bearer auth, got %q", auth)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gate, err := NewGate(GateOptions{BaseURL: srv.URL, Token: "my-token"})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	if err := gate.Ping(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateClient_BaseURL(t *testing.T) {
	gate := newTestGate(t, "http://localhost:8084")
	got := gate.BaseURL()
	if got != "http://localhost:8084" {
		t.Errorf("BaseURL() = %q, want %q", got, "http://localhost:8084")
	}
}

func TestGateClient_BaseURL_WithPath(t *testing.T) {
	gate := newTestGate(t, "http://localhost:8084/api/v1")
	got := gate.BaseURL()
	if got != "http://localhost:8084/api/v1" {
		t.Errorf("BaseURL() = %q, want %q", got, "http://localhost:8084/api/v1")
	}
}
