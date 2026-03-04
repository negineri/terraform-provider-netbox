// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(serverURL string) *NetboxClient {
	return NewNetboxClient(serverURL, "testkey", "testtoken")
}

func TestNewNetboxClient(t *testing.T) {
	c := NewNetboxClient("http://localhost:8080", "mykey", "mytoken")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("unexpected baseURL: %s", c.baseURL)
	}
}

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/status/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("expected Bearer token, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	body, err := c.Get(context.Background(), "api/status/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body == nil {
		t.Fatal("expected non-nil body")
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(*body), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("unexpected status: %s", result["status"])
	}
}

func TestPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		bodyBytes, _ := io.ReadAll(r.Body)
		var payload map[string]string
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if payload["prefix"] != "10.0.0.0/24" {
			t.Errorf("unexpected prefix: %s", payload["prefix"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"prefix":"10.0.0.0/24"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	payload := `{"prefix":"10.0.0.0/24"}`
	body, err := c.Post(context.Background(), "api/ipam/prefixes/", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body == nil {
		t.Fatal("expected non-nil body")
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(*body), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	idVal, ok := result["id"].(float64)
	if !ok || idVal != 1 {
		t.Errorf("unexpected id: %v", result["id"])
	}
}

func TestPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/ipam/prefixes/1/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"status":{"value":"reserved"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	payload := `{"status":"reserved"}`
	body, err := c.Patch(context.Background(), "api/ipam/prefixes/1/", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/ipam/prefixes/1/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.Delete(context.Background(), "api/ipam/prefixes/1/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRequest_HTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Not found."}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Post(context.Background(), "api/ipam/prefixes/", strings.NewReader("{}"))
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to contain status code 404, got: %v", err)
	}
}

func TestGet_RequestHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer nbt_testkey.testtoken" {
			t.Errorf("unexpected Authorization header: %s", authHeader)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected Accept: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Get(context.Background(), "api/status/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// newFastRetryClient は短いリトライ待機時間を持つテスト用クライアントを生成する。
func newFastRetryClient(serverURL string) *NetboxClient {
	c := NewNetboxClient(serverURL, "testkey", "testtoken")
	c.retryClient.RetryWaitMin = 10 * time.Millisecond
	c.retryClient.RetryWaitMax = 50 * time.Millisecond
	c.retryClient.HTTPClient.Timeout = 5 * time.Second
	return c
}

func TestDoRequest_RateLimitRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	c := newFastRetryClient(srv.URL)
	body, err := c.Post(context.Background(), "api/ipam/prefixes/", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}
	if body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestDoRequest_RateLimitExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"detail":"rate limit exceeded"}`))
	}))
	defer srv.Close()

	c := newFastRetryClient(srv.URL)
	_, err := c.Post(context.Background(), "api/ipam/prefixes/", strings.NewReader("{}"))
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if !strings.Contains(err.Error(), "giving up after") {
		t.Errorf("expected error to indicate retry exhaustion, got: %v", err)
	}
}
