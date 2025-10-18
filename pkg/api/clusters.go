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

import "context"

type GetClusterHardwareByIdArg struct {
	ClusterId string
}

type GetClusterHardwareByIdResponse struct {
	Partial   bool `json:"partial"`
	Vcpus     int  `json:"vcpus"`
	MemoryGB  int  `json:"memory_gb"`
	StorageGB int  `json:"storage_gb"`
}

func GetClusterHardwareById(ctx context.Context, c *Client, arg *GetClusterHardwareByIdArg) (*GetClusterHardwareByIdResponse, *APIError) {
	if arg == nil || arg.ClusterId == "" {
		return nil, NewAPIError(400, "cluster_id is required")
	}

	path := c.ExpandPath("/v1/clusters/{cluster_id}/hardware", map[string]string{"cluster_id": arg.ClusterId})

	var resp GetClusterHardwareByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetClusterLicensingByIdArg struct {
	ClusterId string
}
type GetClusterLicensingByIdResponse struct {
	Cluster struct {
		Status     string `json:"status"`
		NextExpiry string `json:"next_expiry"`
	}
	// `node_id` -> { ok: bool, expiry: string }
	Nodes map[string]struct {
		Ok     bool   `json:"ok"`
		Expiry string `json:"expiry"`
	}
}

func GetClusterLicensingById(ctx context.Context, c *Client, arg *GetClusterLicensingByIdArg) (*GetClusterLicensingByIdResponse, *APIError) {
	if arg == nil || arg.ClusterId == "" {
		return nil, NewAPIError(400, "cluster_id is required")
	}

	path := c.ExpandPath("/v1/clusters/{cluster_id}/licensing", map[string]string{"cluster_id": arg.ClusterId})

	var resp GetClusterLicensingByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
