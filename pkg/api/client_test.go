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
	"testing"
)

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
