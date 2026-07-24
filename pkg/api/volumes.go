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
	"net/url"
)

type ListVolumesByNodeOrStoragePoolArg struct {
	NodeId        string
	StoragePoolId string
}
type ListVolumesByNodeOrStoragePoolResponse []VolumeList

func ListVolumesByNodeOrStoragePool(ctx context.Context, c *Client, arg *ListVolumesByNodeOrStoragePoolArg) (*ListVolumesByNodeOrStoragePoolResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "either node_id or storage_pool_id is required")
	}
	if arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	query := make(url.Values)
	query.Set("node_id", arg.NodeId)

	if arg.StoragePoolId != "" {
		query.Set("storage_pool_id", arg.StoragePoolId)
	}

	var resp ListVolumesByNodeOrStoragePoolResponse
	if apiErr := c.Get(ctx, "/v1/volumes", query, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
