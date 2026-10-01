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
package clusterfederation

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PextraCloud/pce-mcp/pkg/api"
)

// newTestFederationServer is a helper to create a test server and client for federation tests.
// It simulates the following endpoints:
//   - GET /api/v1/nodes/{node_id}
//   - GET /api/v1/clusters/{cluster_id}
//
// The server responds with JSON objects containing the information set in nodeCluster and clusterFederation.
func newTestFederationServer(t *testing.T, nodeCluster map[string]string, clusterFederation map[string]string) (*api.Client, *int32, *int32) {
	t.Helper()

	var nodeHits int32
	var clusterHits int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/nodes/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&nodeHits, 1)

		// Path looks like /api/v1/nodes/{node_id}
		nodeId := strings.TrimPrefix(r.URL.Path, "/api/v1/nodes/")
		clusterId := nodeCluster[nodeId]
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"node":{"id":%q,"cluster_id":%q}}`, nodeId, clusterId)
	})
	mux.HandleFunc("/api/v1/clusters/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&clusterHits, 1)

		// Path looks like /api/v1/clusters/{cluster_id}
		clusterId := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")
		federationId := clusterFederation[clusterId]
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"cluster":{"id":%q,"cluster_federation_id":%q}}`, clusterId, federationId)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client, err := api.NewClient(srv.URL, false, 0, "", nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, &nodeHits, &clusterHits
}

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func TestGetCluster(t *testing.T) {
	tests := []struct {
		name           string
		clusterId      string
		nodeCluster    map[string]string
		clusterFed     map[string]string
		wantFed        string
		wantExists     bool
		wantErr        bool
		wantClusterHit int32
	}{
		{
			name:           "happy path",
			clusterId:      "cls-1",
			nodeCluster:    map[string]string{},
			clusterFed:     map[string]string{"cls-1": "fed-1"},
			wantFed:        "fed-1",
			wantExists:     true,
			wantClusterHit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, clusterHits := newTestFederationServer(t, tt.nodeCluster, tt.clusterFed)
			c := New()
			ctx := context.Background()

			gotFed, gotExists, err := c.GetCluster(ctx, client, tt.clusterId)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotFed != tt.wantFed {
				t.Errorf("GetCluster() federation = %q, want %q", gotFed, tt.wantFed)
			}
			if gotExists != tt.wantExists {
				t.Errorf("GetCluster() exists = %v, want %v", gotExists, tt.wantExists)
			}
			if gotHits := atomic.LoadInt32(clusterHits); gotHits != tt.wantClusterHit {
				t.Errorf("GetCluster() cluster endpoint hits = %d, want %d", gotHits, tt.wantClusterHit)
			}
		})
	}
}

func TestGetCluster_SecondCallUsesCache(t *testing.T) {
	client, _, clusterHits := newTestFederationServer(t, map[string]string{}, map[string]string{"cls-1": "fed-1"})
	c := New()
	ctx := context.Background()

	// First call populates the cache.
	fed1, exists1, err := c.GetCluster(ctx, client, "cls-1")
	if err != nil {
		t.Fatalf("GetCluster() first call error: %v", err)
	}
	if !exists1 || fed1 != "fed-1" {
		t.Fatalf("GetCluster() first call = (%q, %v), want (fed-1, true)", fed1, exists1)
	}

	// Second call should hit the cache, not the network.
	fed2, exists2, err := c.GetCluster(ctx, client, "cls-1")
	if err != nil {
		t.Fatalf("GetCluster() second call error: %v", err)
	}
	if !exists2 || fed2 != "fed-1" {
		t.Fatalf("GetCluster() second call = (%q, %v), want (fed-1, true)", fed2, exists2)
	}

	if got := atomic.LoadInt32(clusterHits); got != 1 {
		t.Errorf("cluster endpoint hits = %d, want 1 (second call should use cache)", got)
	}
}

func TestGetNode(t *testing.T) {
	// Node node-1 belongs to cluster cls-1 which belongs to federation fed-9.
	nodeCluster := map[string]string{"node-1": "cls-1"}
	clusterFed := map[string]string{"cls-1": "fed-9"}

	client, nodeHits, _ := newTestFederationServer(t, nodeCluster, clusterFed)
	c := New()
	ctx := context.Background()

	fed, exists, err := c.GetNode(ctx, client, "node-1")
	if err != nil {
		t.Fatalf("GetNode() error: %v", err)
	}
	if !exists {
		t.Fatal("GetNode() exists = false, want true")
	}
	if fed != "fed-9" {
		t.Errorf("GetNode() federation = %q, want %q", fed, "fed-9")
	}
	if got := atomic.LoadInt32(nodeHits); got != 1 {
		t.Errorf("node endpoint hits = %d, want 1", got)
	}
}

func TestGetNode_SecondCallUsesCache(t *testing.T) {
	nodeCluster := map[string]string{"node-1": "cls-1"}
	clusterFed := map[string]string{"cls-1": "fed-9"}

	client, nodeHits, _ := newTestFederationServer(t, nodeCluster, clusterFed)
	c := New()
	ctx := context.Background()

	if _, _, err := c.GetNode(ctx, client, "node-1"); err != nil {
		t.Fatalf("GetNode() first call error: %v", err)
	}
	if _, _, err := c.GetNode(ctx, client, "node-1"); err != nil {
		t.Fatalf("GetNode() second call error: %v", err)
	}

	if got := atomic.LoadInt32(nodeHits); got != 1 {
		t.Errorf("node endpoint hits = %d, want 1 (second call should use cache)", got)
	}
}

func TestReset(t *testing.T) {
	nodeCluster := map[string]string{"node-1": "cls-1"}
	clusterFed := map[string]string{"cls-1": "fed-9"}

	client, nodeHits, _ := newTestFederationServer(t, nodeCluster, clusterFed)
	c := New()
	ctx := context.Background()

	if _, _, err := c.GetNode(ctx, client, "node-1"); err != nil {
		t.Fatalf("GetNode() warmup error: %v", err)
	}

	// Reset should clear the cache so the next call hits the network again.
	c.Reset()

	if _, _, err := c.GetNode(ctx, client, "node-1"); err != nil {
		t.Fatalf("GetNode() after reset error: %v", err)
	}

	if got := atomic.LoadInt32(nodeHits); got != 2 {
		t.Errorf("node endpoint hits after reset = %d, want 2 (reset should invalidate cache)", got)
	}
}

func TestMakeCacheEntry(t *testing.T) {
	before := makeCacheEntry("some-id")
	if before.Id != "some-id" {
		t.Errorf("makeCacheEntry().Id = %q, want %q", before.Id, "some-id")
	}
	// Expiry should be set to now + cacheExpiryDuration, i.e. in the future.
	if !before.Expiry.After(time.Now()) {
		t.Errorf("makeCacheEntry().Expiry = %v, expected to be in the future (~now+%v)", before.Expiry, cacheExpiryDuration)
	}
}
