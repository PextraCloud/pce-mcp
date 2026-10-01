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
	"sync"
	"time"

	"github.com/PextraCloud/pce-mcp/pkg/api"
)

// Default is the default instance of the cluster federation cache.
var Default = New()

// cacheExpiryDuration defines the duration for which a cache entry is considered valid.
const cacheExpiryDuration = 1 * time.Minute

// cacheEntry represents an individual entry in the cache.
type cacheEntry struct {
	Id     string
	Expiry time.Time
}

// makeCacheEntry creates a new cache entry with the given ID and sets its expiry time based on the cacheExpiryDuration.
func makeCacheEntry(id string) cacheEntry {
	return cacheEntry{
		Id:     id,
		Expiry: time.Now().Add(cacheExpiryDuration),
	}
}

// Cache represents the cluster federation cache.
type Cache struct {
	store map[string]cacheEntry
	mu    sync.RWMutex
}

// New creates and returns a new instance of Cache.
func New() *Cache {
	return &Cache{
		store: make(map[string]cacheEntry),
	}
}

// Reset clears the cluster federation cache.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]cacheEntry)
}

// GetNode retrieves the cluster federation ID for the given node ID, either from the cache or by querying the API.
func (c *Cache) GetNode(ctx context.Context, client *api.Client, id string) (string, bool, error) {
	c.mu.RLock()
	if entry, exists := c.store[id]; exists {
		if time.Now().Before(entry.Expiry) {
			return entry.Id, true, nil
		}
	}
	c.mu.RUnlock()

	// TODO: thundering herd issue here
	resp, err := api.GetNodeById(ctx, client, &api.GetNodeByIdArg{NodeId: id})
	if err != nil {
		return "", false, fmt.Errorf("failed to get cluster by ID for node %s: %w", id, err)
	}
	clusterId := resp.Node.ClusterId

	clusterFederationId, _, err2 := c.GetCluster(ctx, client, clusterId)
	if err2 != nil {
		return "", false, fmt.Errorf("failed to get cluster federation ID for node %s (in cluster %s): %w", id, clusterId, err2)
	}

	entry := makeCacheEntry(clusterFederationId)
	c.mu.Lock()
	c.store[id] = entry
	c.mu.Unlock()

	return entry.Id, true, nil
}

// GetCluster retrieves the cluster federation ID for the given cluster ID, either from the cache or by querying the API.
func (c *Cache) GetCluster(ctx context.Context, client *api.Client, id string) (string, bool, error) {
	c.mu.RLock()
	if entry, exists := c.store[id]; exists {
		if time.Now().Before(entry.Expiry) {
			return entry.Id, true, nil
		}
	}
	c.mu.RUnlock()

	// TODO: thundering herd issue here
	resp, err := api.GetClusterById(ctx, client, &api.GetClusterByIdArg{ClusterId: id})
	if err != nil {
		return "", false, fmt.Errorf("failed to get cluster federation ID for cluster %s: %w", id, err)
	}
	clusterFederationId := resp.Cluster.ClusterFederationId

	entry := makeCacheEntry(clusterFederationId)
	c.mu.Lock()
	c.store[id] = entry
	c.mu.Unlock()

	return entry.Id, true, nil
}
