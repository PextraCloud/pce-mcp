/*
Copyright 2025 Pextra Inc.

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
)

type ListImagesByNodeArg struct {
	NodeId string
}
type ListImagesByNodeResponse = []ImageList

func ListImagesByNode(ctx context.Context, c *Client, arg *ListImagesByNodeArg) (*ListImagesByNodeResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}/images/images", map[string]string{"node_id": arg.NodeId})

	var resp ListImagesByNodeResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
