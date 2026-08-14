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
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
)

type GetInstancesByIdArg struct {
	// Either `NodeId` or `ClusterId` must be provided
	NodeId    string
	ClusterId string
}
type GetInstancesByIdResponse = []InstanceList

func GetInstancesById(ctx context.Context, c *Client, arg *GetInstancesByIdArg) (*GetInstancesByIdResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "either node_id or cluster_id is required")
	}
	if arg.NodeId != "" && arg.ClusterId != "" {
		return nil, NewAPIError(400, "only one of node_id or cluster_id should be provided")
	}

	query := make(url.Values)
	if arg.NodeId != "" {
		query.Set("node_id", arg.NodeId)
	} else if arg.ClusterId != "" {
		query.Set("cluster_id", arg.ClusterId)
	} else {
		return nil, NewAPIError(400, "either node_id or cluster_id is required")
	}

	path := "/v1/instances"

	var resp GetInstancesByIdResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type PowerInstanceArg struct {
	InstanceId string
	NodeId     string
	Action     enum.InstancePowerAction
}
type PowerInstanceResponse struct {
	TaskId string `json:"task_id"`
}

func PowerInstance(ctx context.Context, c *Client, arg *PowerInstanceArg) (*PowerInstanceResponse, *APIError) {
	if arg == nil || arg.NodeId == "" || arg.InstanceId == "" {
		return nil, NewAPIError(400, "node_id and instance_id are required")
	}
	if !enum.InstancePowerAction.IsValid(arg.Action) {
		return nil, NewAPIError(400, "invalid action")
	}

	path := c.ExpandPath("/v1/instances/{instance_id}/power", map[string]string{"instance_id": arg.InstanceId})

	query := make(url.Values)
	query.Set("node_id", arg.NodeId)

	payload, err := json.Marshal(map[string]string{"action": string(arg.Action)})
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp PowerInstanceResponse
	if apiErr := c.Post(ctx, path, query, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type DeployInstanceV2CPUArg struct {
	Cores    int    `json:"cores"`
	Sockets  int    `json:"sockets"`
	Threads  int    `json:"threads"`
	Affinity string `json:"affinity,omitempty"`
}

type DeployInstanceV2BandwidthRateArg struct {
	Average string `json:"average,omitempty"`
	Burst   string `json:"burst,omitempty"`
}

type DeployInstanceV2BandwidthArg struct {
	Inbound  DeployInstanceV2BandwidthRateArg `json:"inbound"`
	Outbound DeployInstanceV2BandwidthRateArg `json:"outbound"`
}

type DeployInstanceV2IPAddressArg struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type DeployInstanceV2NetworkArg struct {
	Mac           string                        `json:"mac,omitempty"`
	VSwitchId     string                        `json:"vswitch_id"`
	PortGroupName string                        `json:"port_group_name"`
	Bandwidth     *DeployInstanceV2BandwidthArg `json:"bandwidth,omitempty"`
	Model         string                        `json:"model,omitempty"`
	IPv4          *DeployInstanceV2IPAddressArg `json:"ipv4,omitempty"`
	IPv6          *DeployInstanceV2IPAddressArg `json:"ipv6,omitempty"`
}

type deployInstanceV2Destination struct {
	NodeId string `json:"node_id"`
}

type deployInstanceV2Payload struct {
	Architecture string                       `json:"architecture"`
	CPU          DeployInstanceV2CPUArg       `json:"cpu"`
	Destination  deployInstanceV2Destination  `json:"destination"`
	Memory       int                          `json:"memory"`
	Name         string                       `json:"name"`
	Networks     []DeployInstanceV2NetworkArg `json:"networks"`
	Metadata     any                          `json:"metadata"`
	Type         enum.InstanceTypeEnum        `json:"type"`
	Autostart    bool                         `json:"autostart"`
	Description  string                       `json:"description,omitempty"`
	Image        string                       `json:"image,omitempty"`
	IMDSEnabled  bool                         `json:"imds_enabled"`
}

type DeployInstanceV2LXCMetadata struct {
	Init string `json:"init,omitempty"`
}

type DeployInstanceV2QEMUSecureBoot struct {
	Enabled            bool `json:"enabled"`
	EnrollStandardKeys bool `json:"enroll_standard_keys"`
}

type DeployInstanceV2QEMUVolumeBase struct {
	Bus      enum.QEMUDiskBusEnum       `json:"bus"`
	Dev      string                     `json:"dev"`
	Driver   enum.LibvirtDiskDriverEnum `json:"driver"`
	Backup   *bool                      `json:"backup,omitempty"`
	Readonly *bool                      `json:"readonly,omitempty"`
}

type DeployInstanceV2QEMUExistingVolume struct {
	DeployInstanceV2QEMUVolumeBase
	VolumeId string `json:"volume_id"`
}
type DeployInstanceV2QEMUNewVolume struct {
	DeployInstanceV2QEMUVolumeBase
	Size          float64 `json:"size"`
	StoragePoolId string  `json:"storage_pool_id"`
}

type DeployInstanceV2QEMUInstanceMetadata struct {
	CPUModel        string                               `json:"cpu_model"`
	MachineType     string                               `json:"machine_type"`
	Firmware        string                               `json:"firmware"`
	SecureBoot      DeployInstanceV2QEMUSecureBoot       `json:"secure_boot"`
	NewVolumes      []DeployInstanceV2QEMUNewVolume      `json:"new_volumes"`
	ExistingVolumes []DeployInstanceV2QEMUExistingVolume `json:"existing_volumes"`
}

type DeployInstanceV2Arg struct {
	Type         enum.InstanceTypeEnum
	NodeId       string
	Name         string
	Description  string
	Architecture string
	Image        string
	CPU          DeployInstanceV2CPUArg
	Memory       int
	IMDSEnabled  bool
	Networks     []DeployInstanceV2NetworkArg
	Autostart    bool
	LXCMetadata  *DeployInstanceV2LXCMetadata
	QEMUMetadata *DeployInstanceV2QEMUInstanceMetadata
}
type DeployInstanceV2Result struct {
	TaskId string `json:"task_id"`
}

func DeployInstanceV2(ctx context.Context, c *Client, arg *DeployInstanceV2Arg) (*DeployInstanceV2Result, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}
	if arg.Name == "" || arg.Architecture == "" {
		return nil, NewAPIError(400, "name and architecture are required")
	}
	if arg.CPU.Sockets <= 0 || arg.CPU.Cores <= 0 || arg.CPU.Threads <= 0 {
		return nil, NewAPIError(400, "CPU sockets, cores, and threads must be greater than zero")
	}
	if arg.Memory < 64 {
		return nil, NewAPIError(400, "memory must be at least 64")
	}

	metadata := make(map[string]any)
	switch arg.Type {
	case enum.InstanceTypeEnumLXC:
		if arg.LXCMetadata == nil {
			return nil, NewAPIError(400, "LXC metadata is required for LXC instances")
		}
		metadata["_type"] = "lxc"
		metadata["init"] = arg.LXCMetadata.Init
	case enum.InstanceTypeEnumQEMU:
		if arg.QEMUMetadata == nil {
			return nil, NewAPIError(400, "QEMU metadata is required for QEMU instances")
		}
		metadata["_type"] = "qemu"
		metadata["cpu_model"] = arg.QEMUMetadata.CPUModel
		metadata["existing_volumes"] = arg.QEMUMetadata.ExistingVolumes
		metadata["firmware"] = arg.QEMUMetadata.Firmware
		metadata["machine_type"] = arg.QEMUMetadata.MachineType
		metadata["new_volumes"] = arg.QEMUMetadata.NewVolumes
		metadata["secure_boot"] = arg.QEMUMetadata.SecureBoot
	default:
		return nil, NewAPIError(400, "invalid instance type")
	}

	payload := &deployInstanceV2Payload{
		Destination: deployInstanceV2Destination{
			NodeId: arg.NodeId,
		},
		Name:         arg.Name,
		Description:  arg.Description,
		Architecture: arg.Architecture,
		Image:        arg.Image,
		Type:         arg.Type,
		CPU:          arg.CPU,
		Memory:       arg.Memory,
		IMDSEnabled:  arg.IMDSEnabled,
		Networks:     arg.Networks,
		Metadata:     metadata,
		Autostart:    arg.Autostart,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp DeployInstanceV2Result
	if apiError := c.Post(ctx, "/v2/instances", nil, bytes.NewReader(data), &resp); apiError != nil {
		return nil, apiError
	}
	return &resp, nil
}
