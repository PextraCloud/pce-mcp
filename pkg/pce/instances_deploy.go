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
	type reqType struct {
		NodeId             string `json:"node_id"`
		InstanceTypeString string `json:"instance_type"`
	}

	args := &reqType{}
	if err := req.BindArguments(args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	instanceType, ok := enum.InstanceTypeEnumFromString(args.InstanceTypeString)
	if !ok {
		return mcp.NewToolResultError("instance_type not valid"), nil
	}

	nodeId := args.NodeId

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

type deployInstanceResult struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
	TaskIds []string `json:"task_ids"`
}

func DeployInstance() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("deploy_instance",
		mcp.WithDescription("Deploy a QEMU or LXC instance using choices returned by get_deploy_instance_context; uses recommended defaults unless overridden. Call only after showing a summary and receiving explicit confirmation."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Deploy Instance",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id",
			mcp.Description("Node to deploy the instance on (format: node-<xxx>); if not specified, the current node is used"),
		),
		mcp.WithInteger("count",
			mcp.Min(1),
			mcp.Max(16),
			mcp.Description("Number of instances to deploy. Defaults to 1."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the instance to deploy; if deploying multiple instances, names will be suffixed with a number (e.g., name-1, name-2)"),
			// API allows up to 255, but we will limit to 254 to allow for suffixing when deploying multiple instances
			mcp.MinLength(3),
			mcp.MaxLength(253),
			mcp.Pattern(nameRegex(nameDefaultMinLength, 253)),
		),
		mcp.WithString("architecture",
			mcp.Required(),
			mcp.Description("Architecture of the instance to deploy (e.g., x86_64, arm64); must match the architecture of the selected image"),
		),
		mcp.WithString("description",
			mcp.Description("Description of the instance to deploy"),
			mcp.MaxLength(512),
		),
		mcp.WithString("instance_type",
			mcp.Required(),
			mcp.Description("Type of instance to deploy, the image must match the instance type"),
			mcp.Enum("qemu", "lxc"),
		),
		mcp.WithString("image",
			mcp.Required(),
			mcp.Description("Name of the instance image to use"),
		),
		mcp.WithInteger("vcpus",
			mcp.Min(1),
			mcp.Description(fmt.Sprintf("Number of virtual CPUs to allocate; over-provisioning is allowed but may affect performance. Defaults to %d", deployInstanceDefaultCPUCores)),
		),
		mcp.WithInteger("memory_mb",
			mcp.Min(64),
			mcp.Description(fmt.Sprintf("Amount of memory in mebibytes (MiB) to allocate; over-provisioning is allowed but may affect performance. Defaults to %d MiB (%d GiB)", deployInstanceDefaultMemoryMB, deployInstanceDefaultMemoryMB/1024)),
		),
		mcp.WithBoolean("autostart",
			mcp.Description("Whether the instance should automatically start on node boot; this will also start the instance immediately after deployment if set to true. Defaults to false."),
		),
		mcp.WithArray("networks",
			mcp.Description("List of network interface devices to initially attach to the instance"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"vswitch_id": map[string]any{
						"type":        "string",
						"description": "ID of the vswitch to attach the instance to (format: svs-<xxx>); must be a valid vswitch on the specified node",
					},
					"port_group_name": map[string]any{
						"type":        "string",
						"description": "Name of the port group to attach the instance to; must be a valid port group on the specified vswitch",
					},
					"model": map[string]any{
						"type":        "string",
						"description": "Model of the network interface device; defaults to 'e1000'",
					},
				},
				"required": []string{"vswitch_id", "port_group_name"},
			}),
		),
		mcp.WithArray("existing_volumes",
			mcp.Description("List of existing volumes to attach to the instance (format: vol-<xxx>); these volumes must be available and not attached to any other instance. Cannot be specified when instance count > 1. LXC instances currently do not support attaching volumes, so this list will be ignored for LXC. Note: new_volumes and existing_volumes combined cannot exceed 16 volumes."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"volume_id": map[string]any{
						"type":        "string",
						"description": "ID of the existing volume to attach (format: vol-<xxx>); must be a valid volume on the specified node and not attached to any other instance",
					},
					"bus": map[string]any{
						"type":        "string",
						"description": "Bus type for the existing volume; defaults to 'virtio'",
						"enum":        enum.QEMUDiskBusEnumValues,
					},
				},
				"required": []string{"volume_id"},
			}),
		),
		mcp.WithArray("new_volumes",
			mcp.Description("List of new volumes to create and attach to the instance; each volume will be created on the specified storage pool with the specified size in gibibytes (GiB). LXC instances currently do not support attaching volumes, so this list will be ignored for LXC. Note: new_volumes and existing_volumes combined cannot exceed 16 volumes."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"size_gb": map[string]any{
						"type":        "number",
						"description": "Size of the new volume to create in gibibytes (GiB); must be at least 1 GiB and not exceed 90% of the available space on the specified storage pool",
						"minimum":     1,
					},
					"bus": map[string]any{
						"type":        "string",
						"description": "Bus type for the new volume; defaults to 'virtio'",
						"enum":        enum.QEMUDiskBusEnumValues,
					},
					"storage_pool_id": map[string]any{
						"type":        "string",
						"description": "ID of the storage pool to create the new volume on (format: pool-<xxx>); must be a valid storage pool on the specified node and have enough available space for the new volume",
					},
				},
				"required": []string{"size_gb", "storage_pool_id"},
			}),
		),
		mcp.WithOutputSchema[deployInstanceResult](),
	), handleDeployInstance
}

func handleDeployInstance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type reqNewVolume struct {
		SizeGB        float64              `json:"size_gb"`
		Bus           enum.QEMUDiskBusEnum `json:"bus"`
		StoragePoolId string               `json:"storage_pool_id"`
	}
	type reqExistingVolume struct {
		VolumeId string               `json:"volume_id"`
		Bus      enum.QEMUDiskBusEnum `json:"bus"`
	}
	type reqType struct {
		NodeId             string                           `json:"node_id"`
		Count              int                              `json:"count"`
		Name               string                           `json:"name"`
		Architecture       string                           `json:"architecture"`
		Description        string                           `json:"description"`
		InstanceTypeString string                           `json:"instance_type"`
		Image              string                           `json:"image"`
		Vcpus              int                              `json:"vcpus"`
		MemoryMB           int                              `json:"memory_mb"`
		Autostart          bool                             `json:"autostart"`
		Networks           []api.DeployInstanceV2NetworkArg `json:"networks"`
		ExistingVolumes    []reqExistingVolume              `json:"existing_volumes"`
		NewVolumes         []reqNewVolume                   `json:"new_volumes"`
	}
	args := &reqType{}
	if err := req.BindArguments(args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	instanceType, ok := enum.InstanceTypeEnumFromString(args.InstanceTypeString)
	if !ok {
		return mcp.NewToolResultError("instance_type not valid"), nil
	}

	if len(args.ExistingVolumes)+len(args.NewVolumes) > 16 {
		return mcp.NewToolResultError("total number of volumes (existing + new) cannot exceed 16"), nil
	}

	// Set defaults
	if args.Vcpus == 0 {
		args.Vcpus = deployInstanceDefaultCPUCores
	}
	if args.MemoryMB == 0 {
		args.MemoryMB = deployInstanceDefaultMemoryMB
	}
	if args.Count == 0 {
		args.Count = 1
	}

	// When deploying multiple instances, existing volumes cannot be specified as they can only be attached to one instance at a time,
	// deploying multiple instances with existing volumes would result in a conflict.
	// New volumes can be specified as they will be created for each instance.
	if args.Count > 1 && len(args.ExistingVolumes) > 0 {
		return mcp.NewToolResultError("cannot specify existing_volumes when deploying multiple instances"), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// If no node ID is provided, use the current node ID
	if args.NodeId == "" {
		currentIds, err := getCurrentTreeIds(ctx, client)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		args.NodeId = currentIds.NodeId
	}

	var taskIds []string
	var errors []string

	var lxcMetadata *api.DeployInstanceV2LXCMetadata
	var qemuMetadata *api.DeployInstanceV2QEMUInstanceMetadata

	// 0 -> 'sd(a)' through 25 -> 'sd(z)' ('a'+dev); we do not need to implement 'sdaa' etc. due to 16-disk limit
	// Start from sdb since image CD-ROM device is attached as 'sda'
	dev := 1

	switch instanceType {
	case enum.InstanceTypeEnumLXC:
		lxcMetadata = &api.DeployInstanceV2LXCMetadata{}
	case enum.InstanceTypeEnumQEMU:
		instanceNewVolumes := make([]api.DeployInstanceV2QEMUNewVolume, len(args.NewVolumes))
		instanceExistingVolumes := make([]api.DeployInstanceV2QEMUExistingVolume, len(args.ExistingVolumes))
		for i, vol := range args.NewVolumes {
			// Set default if not specified
			bus := vol.Bus
			if bus == "" {
				bus = enum.QEMUDiskBusVirtio
			}

			devString := fmt.Sprintf("sd%c", 'a'+dev)
			dev++
			instanceNewVolumes[i] = api.DeployInstanceV2QEMUNewVolume{
				DeployInstanceV2QEMUVolumeBase: api.DeployInstanceV2QEMUVolumeBase{
					Bus: bus,
					Dev: devString,
					// Default to qcow2
					Driver: enum.LibvirtDiskDriverQcow2,
				},
				Size:          vol.SizeGB,
				StoragePoolId: vol.StoragePoolId,
			}
		}
		for i, vol := range args.ExistingVolumes {
			// Set default if not specified
			bus := vol.Bus
			if bus == "" {
				bus = enum.QEMUDiskBusVirtio
			}

			devString := fmt.Sprintf("sd%c", 'a'+dev)
			dev++
			instanceExistingVolumes[i] = api.DeployInstanceV2QEMUExistingVolume{
				DeployInstanceV2QEMUVolumeBase: api.DeployInstanceV2QEMUVolumeBase{
					Bus: bus,
					Dev: devString,
					// TODO: this is faulty for existing volumes
					Driver: enum.LibvirtDiskDriverQcow2,
				},
				VolumeId: vol.VolumeId,
			}
		}

		qemuMetadata = &api.DeployInstanceV2QEMUInstanceMetadata{
			CPUModel:    "default",
			MachineType: deployInstanceDefaultQEMUMachineType,
			Firmware:    deployInstanceDefaultQEMUFirmware,
			SecureBoot: api.DeployInstanceV2QEMUSecureBoot{
				Enabled:            false,
				EnrollStandardKeys: false,
			},
			NewVolumes:      instanceNewVolumes,
			ExistingVolumes: instanceExistingVolumes,
		}
	default:
		// This should never happen
		return mcp.NewToolResultError("instance_type not valid"), nil
	}

	instanceNetworks := make([]api.DeployInstanceV2NetworkArg, len(args.Networks))
	for i, net := range args.Networks {
		// Set default if not specified
		model := net.Model
		if model == "" {
			model = "e1000"
		}
		instanceNetworks[i] = api.DeployInstanceV2NetworkArg{
			VSwitchId:     net.VSwitchId,
			PortGroupName: net.PortGroupName,
			Model:         model,
		}
	}

	for i := range args.Count {
		instanceName := args.Name
		if args.Count > 1 {
			instanceName = fmt.Sprintf("%s-%d", args.Name, i+1)
		}

		payload := &api.DeployInstanceV2Arg{
			Type:         instanceType,
			NodeId:       args.NodeId,
			Name:         instanceName,
			Description:  args.Description,
			Architecture: args.Architecture,
			Image:        args.Image,
			CPU: api.DeployInstanceV2CPUArg{
				Sockets: 1,
				Cores:   args.Vcpus,
				Threads: 1,
			},
			Memory:       args.MemoryMB,
			IMDSEnabled:  false,
			Networks:     instanceNetworks,
			Autostart:    args.Autostart,
			LXCMetadata:  lxcMetadata,
			QEMUMetadata: qemuMetadata,
		}
		res, err := api.DeployInstanceV2(ctx, client, payload)
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			taskIds = append(taskIds, res.TaskId)
		}
	}

	// Return array of task IDs
	numSuccess := args.Count - len(errors)
	message := fmt.Sprintf("Successfully initiated deployment of %d/%d instance(s) of type '%s' with image '%s' on node '%s'.", numSuccess, args.Count, args.InstanceTypeString, args.Image, args.NodeId)
	if numSuccess == 0 {
		message = fmt.Sprintf("Failed to deploy any instances of type '%s' with image '%s' on node '%s'. See errors for details.", args.InstanceTypeString, args.Image, args.NodeId)
	} else if len(errors) > 0 {
		message += fmt.Sprintf(" %d instance(s) failed to deploy. See errors for details.", len(errors))
	}

	return mcp.NewToolResultJSON(&deployInstanceResult{
		Message: message,
		TaskIds: taskIds,
		Errors:  errors,
	})
}
