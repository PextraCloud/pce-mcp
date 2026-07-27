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
package pce

import (
	"context"
"fmt"

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	deployInstanceDefaultCPUCores        = 1
	deployInstanceDefaultMemoryMB        = 2048
		deployInstanceDefaultQEMUFirmware    = "efi"
	deployInstanceDefaultQEMUMachineType = "q35"
)

type getDeployInstanceContextStoragePool struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	AvailableGB float64 `json:"available_gb"`
}

type getDeployInstanceContextNetwork struct {
	VswitchId              string `json:"vswitch_id"`
	VswitchName            string `json:"vswitch_name"`
	VswitchHasUplinks      bool   `json:"vswitch_has_uplinks"`
	PortGroupName          string `json:"port_group_name"`
	PortGroupConfigSummary string `json:"port_group_config_summary"`
}

type getDeployInstanceContextVolume struct {
	Id     string  `json:"id"`
	Name   string  `json:"name"`
	SizeGB float64 `json:"size_gb"`
}

type getDeployInstanceContextResult struct {
	NodeId       string                `json:"node_id"`
NodeVcpus    int                   `json:"node_vcpus"`
	NodeMemoryMB int                   `json:"node_memory_mb"`
	InstanceType enum.InstanceTypeEnum `json:"instance_type"`
	// Only return image names (unique)
	Images []string `json:"images"`
	// Available storage pools
	AvailableStoragePools   []getDeployInstanceContextStoragePool `json:"available_storage_pools"`
	Networks                []getDeployInstanceContextNetwork     `json:"networks"`
	AvailableVolumes        []getDeployInstanceContextVolume      `json:"available_volumes"`
	RecommendedArchitecture string                                `json:"recommended_architecture"`
}

func GetDeployInstanceContext() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_deploy_instance_context",
		mcp.WithDescription("Gather context for deploying an instance; this tool returns information about valid images, storage pools, networks, and other necessary details for deploying an instance. After gathering this context, show a summary and ask for explicit confirmation before proceeding with the deployment."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Deploy Instance Context",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Description("Node to get deployment context on (format: node-<xxx>); if not specified, the current node is used"),
		),
		mcp.WithString("instance_type",
			mcp.Required(),
			mcp.Description("Type of instance to deploy (e.g., qemu, lxc); images will be filtered based on this type. LXC instances currently do not support attaching volumes, so the storage pool list will be empty for LXC."),
			mcp.Enum("qemu", "lxc"),
		),
		mcp.WithOutputSchema[getDeployInstanceContextResult](),
	), handleGetDeployInstanceContext
}

func handleGetDeployInstanceContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instanceTypeString, _ := requiredParam[string](req, "instance_type")
	instanceType, ok := enum.InstanceTypeEnumFromString(instanceTypeString)
	if !ok {
		return mcp.NewToolResultError("instance_type not valid"), nil
	}

	nodeId, _ := optionalParam[string](req, "node_id")

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// If no node ID is provided, use current node ID
	if nodeId == "" {
		currentIds, err := getCurrentTreeIds(ctx, client)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		nodeId = currentIds.NodeId
	}

hardware, apiErr := api.GetNodeHardwareById(ctx, client, &api.GetNodeHardwareByIdArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	images, apiErr := api.ListImagesByNode(ctx, client, &api.ListImagesByNodeArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	pools, apiErr := api.GetNodeStoragePoolsById(ctx, client, &api.GetNodeStoragePoolsByIdArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	volumes, apiErr := api.ListVolumesByNodeOrStoragePool(ctx, client, &api.ListVolumesByNodeOrStoragePoolArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	vswitches, apiErr := api.ListStandaloneVswitches(ctx, client, &api.ListStandaloneVswitchesArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	portGroups, apiErr := api.ListStandalonePortGroups(ctx, client, &api.ListStandalonePortGroupsArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	capabilities, apiErr := api.GetNodeVirtualizationCapabilities(ctx, client, &api.GetNodeVirtualizationCapabilitiesArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}

// Accumulate total memory in MB from all memory banks (each bank size is in GiB)
	nodeMemoryMb := 0
	for _, bank := range hardware.Memory {
		nodeMemoryMb += bank.Data.Size * 1024
	}

	result := &getDeployInstanceContextResult{
		NodeId:       nodeId,
		NodeVcpus:    hardware.Vcpus,
		NodeMemoryMB: nodeMemoryMb,
		InstanceType: instanceType,
	}

	// Available storage pools
	// Return empty list for LXC as storage pools for LXC instances are not currently supported
	if instanceType != enum.InstanceTypeEnumLXC {
		for _, pool := range *pools {
			if !pool.Available {
				continue
			}
			result.AvailableStoragePools = append(result.AvailableStoragePools, getDeployInstanceContextStoragePool{
				Id:          pool.Id,
				Name:        pool.Name,
				AvailableGB: pool.Usage.AvailableGB,
				Type:        pool.Type.String(),
			})
		}
	}

	// Networks (vswitches * port groups)
	vswitchMap := make(map[string]api.VswitchList)
	for _, vswitch := range vswitches {
		vswitchMap[vswitch.Id] = vswitch
	}
	for _, portGroup := range portGroups {
		vswitch, ok := vswitchMap[portGroup.VswitchId]
		if !ok {
			continue
		}
		result.Networks = append(result.Networks, getDeployInstanceContextNetwork{
			VswitchId:              vswitch.Id,
			VswitchName:            vswitch.Name,
			VswitchHasUplinks:      vswitch.UplinkCount > 0,
			PortGroupName:          portGroup.Name,
			PortGroupConfigSummary: portGroup.Config.String(),
		})
	}

	// Available existing volumes (not attached to an instance)
	for _, volume := range *volumes {
		if volume.Status == enum.VolumeStatusEnumAvailable && !volume.Attached {
			result.AvailableVolumes = append(result.AvailableVolumes, getDeployInstanceContextVolume{
				Id:     volume.Id,
				Name:   volume.Name,
				SizeGB: volume.Size,
			})
		}
	}

	// Images for specified instance type
	for _, image := range *images {
		if image.Type == instanceType {
			result.Images = append(result.Images, image.Name)
		}
	}

	// Prefer first 64-bit KVM-supported architecture
	for arch, caps := range *capabilities {
		if caps.WordSize == 64 && caps.KVMSupported {
			result.RecommendedArchitecture = arch
			break
		}
	}

	return mcp.NewToolResultJSON(&result)
}
