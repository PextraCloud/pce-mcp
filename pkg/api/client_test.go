/*
Copyright 2026 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testBaseURL = "http://test.invalid"

func TestClusterFederationIdContextRoundTrip(t *testing.T) {
	t.Run("empty id is not stored", func(t *testing.T) {
		ctx := WithClusterFederationId(context.Background(), "")
		if got := ClusterFederationIdFromContext(ctx); got != "" {
			t.Errorf("expected empty cluster federation id, got %q", got)
		}
	})

	t.Run("id round-trips", func(t *testing.T) {
		ctx := WithClusterFederationId(context.Background(), "fed-abc")
		if got := ClusterFederationIdFromContext(ctx); got != "fed-abc" {
			t.Errorf("expected fed-abc, got %q", got)
		}
	})
}

func TestNewRequestClusterFederationHeader(t *testing.T) {
	newTestClusterFederationClient := func(t *testing.T, staticHeaders http.Header) *Client {
		t.Helper()
		c, err := NewClient(testBaseURL, false, 0, "", staticHeaders)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		return c
	}

	t.Run("no id means no header", func(t *testing.T) {
		c := newTestClusterFederationClient(t, nil)
		ctx := WithClusterFederationId(context.Background(), "")
		req, apiErr := c.newRequest(ctx, http.MethodGet, "/v1/nodes/node-1", nil, nil)
		if apiErr != nil {
			t.Fatalf("newRequest: %v", apiErr)
		}
		if got := req.Header.Get(ClusterFederationIdHeader); got != "" {
			t.Errorf("expected no cluster federation header, got %q", got)
		}
	})

	t.Run("context id is set on request", func(t *testing.T) {
		c := newTestClusterFederationClient(t, nil)
		ctx := WithClusterFederationId(context.Background(), "fed-xyz")
		req, apiErr := c.newRequest(ctx, http.MethodGet, "/v1/nodes/node-1", nil, nil)
		if apiErr != nil {
			t.Fatalf("newRequest: %v", apiErr)
		}
		if got := req.Header.Get(ClusterFederationIdHeader); got != "fed-xyz" {
			t.Errorf("expected %q header, got %q", "fed-xyz", got)
		}
	})

	t.Run("context id overrides static header", func(t *testing.T) {
		static := make(http.Header)
		static.Set(ClusterFederationIdHeader, "static-fed")
		c := newTestClusterFederationClient(t, static)
		ctx := WithClusterFederationId(context.Background(), "ctx-fed")
		req, apiErr := c.newRequest(ctx, http.MethodGet, "/v1/nodes/node-1", nil, nil)
		if apiErr != nil {
			t.Fatalf("newRequest: %v", apiErr)
		}
		if got := req.Header.Get(ClusterFederationIdHeader); got != "ctx-fed" {
			t.Errorf("expected context header to win with %q, got %q", "ctx-fed", got)
		}
	})
}

func TestClusterFederationHeaderOnWire(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(ClusterFederationIdHeader)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, false, 0, "", nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := WithClusterFederationId(context.Background(), "fed-wire")
	var out map[string]any
	if apiErr := client.Get(ctx, "/v1/clusters/cls-1", nil, &out); apiErr != nil {
		t.Fatalf("Get: %v", apiErr)
	}
	if gotHeader != "fed-wire" {
		t.Errorf("expected %q header on the wire, got %q", "fed-wire", gotHeader)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		params   map[string]string
		expected string
	}{
		{
			name:     "single placeholder",
			path:     "/nodes/{node_id}/images",
			params:   map[string]string{"node_id": "123"},
			expected: "/nodes/123/images",
		},
		{
			name:     "multiple placeholders",
			path:     "/orgs/{org_id}/nodes/{node_id}/images/{image_id}",
			params:   map[string]string{"org_id": "abc", "node_id": "xyz", "image_id": "img123"},
			expected: "/orgs/abc/nodes/xyz/images/img123",
		},
		{
			name:     "placeholder with special characters in value",
			path:     "/nodes/{node_id}/status",
			params:   map[string]string{"node_id": "node-123_test"},
			expected: "/nodes/node-123_test/status",
		},
		{
			name:     "placeholder with URL special characters",
			path:     "/search/{query}",
			params:   map[string]string{"query": "hello world"},
			expected: "/search/hello%20world",
		},
		{
			name:     "no placeholders",
			path:     "/nodes/list",
			params:   map[string]string{},
			expected: "/nodes/list",
		},
		{
			name:     "empty params map",
			path:     "/nodes/{node_id}",
			params:   nil,
			expected: "/nodes/{node_id}",
		},
		{
			name:     "placeholder with slash in value",
			path:     "/path/{segment}/end",
			params:   map[string]string{"segment": "foo/bar"},
			expected: "/path/foo%2Fbar/end",
		},
		{
			name:     "multiple same placeholders",
			path:     "/{type}/{type}/end",
			params:   map[string]string{"type": "value"},
			expected: "/value/value/end",
		},
		{
			name:     "placeholder with empty value",
			path:     "/nodes/{node_id}/info",
			params:   map[string]string{"node_id": ""},
			expected: "/nodes//info",
		},
		{
			name:     "path with query string format",
			path:     "/api/v1/resources/{resource_id}",
			params:   map[string]string{"resource_id": "res-456"},
			expected: "/api/v1/resources/res-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			result := c.ExpandPath(tt.path, tt.params)
			if result != tt.expected {
				t.Errorf("ExpandPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}
