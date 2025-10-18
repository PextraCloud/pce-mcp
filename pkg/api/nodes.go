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

type GetNodeByIdArg struct {
	NodeId string
}
type GetNodeByIdResponse struct {
	Node      NodeDetail     `json:"node"`
	Instances []InstanceList `json:"instances"`
	Hostname  string         `json:"hostname"`
	Os        struct {
		Kernel       string `json:"kernel"`
		Architecture string `json:"architecture"`
		Uefi         bool   `json:"uefi"`
	} `json:"os"`
	Time struct {
		Time     int64  `json:"time"`
		Timezone string `json:"timezone"`
		Uptime   int64  `json:"uptime"`
	} `json:"time"`
	PceVersion string `json:"pce_version"`
}

func GetNodeById(ctx context.Context, c *Client, arg *GetNodeByIdArg) (*GetNodeByIdResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}", map[string]string{"node_id": arg.NodeId})

	var resp GetNodeByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetNodeHardwareByIdArg struct {
	NodeId string
}

type GetNodeHardwareByIdResponse struct {
	Vcpus  int                  `json:"vcpus"`
	CPU    NodeHardwareCpu      `json:"cpu"`
	Memory []NodeHardwareMemory `json:"memory"`
	Disks  []NodeHardwareDisk   `json:"disks"`
	Usb    []NodeHardwareUsb    `json:"usb"`
}

func GetNodeHardwareById(ctx context.Context, c *Client, arg *GetNodeHardwareByIdArg) (*GetNodeHardwareByIdResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}/hardware", map[string]string{"node_id": arg.NodeId})

	var resp GetNodeHardwareByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetNodeLicenseByIdArg struct {
	NodeId string
}
type GetNodeLicenseByIdResponse struct {
	Key    string `json:"key"`
	Expiry string `json:"expiry"`
	Valid  bool   `json:"valid"`
}

func GetNodeLicenseById(ctx context.Context, c *Client, arg *GetNodeByIdArg) (*GetNodeLicenseByIdResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}/license", map[string]string{"node_id": arg.NodeId})

	var resp GetNodeLicenseByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetNodeStoragePoolsByIdArg struct {
	NodeId string
}
type GetNodeStoragePoolsByIdResponse = []StoragePoolDetail

func GetNodeStoragePoolsById(ctx context.Context, c *Client, arg *GetNodeStoragePoolsByIdArg) (*GetNodeStoragePoolsByIdResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}/storage/pools", map[string]string{"node_id": arg.NodeId})

	var resp GetNodeStoragePoolsByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetNodePciDevicesByIdArg struct {
	NodeId string
}
type GetNodePciDevicesByIdResponse = []NodePciDevice

func GetNodePciDevicesById(ctx context.Context, c *Client, arg *GetNodePciDevicesByIdArg) (*GetNodePciDevicesByIdResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/nodes/{node_id}/hardware/pci", map[string]string{"node_id": arg.NodeId})

	var resp GetNodePciDevicesByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
