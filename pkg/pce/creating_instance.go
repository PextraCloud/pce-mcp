package pce

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	defaultCPUCores       = 1
	defaultMemoryMB       = 2048
	defaultQEMUStorageGB  = 0.0
	defaultQEMUFirmware   = "efi"
	defaultQEMUNetwork    = "virtio"
	defaultQEMUVolumeBus  = "virtio"
	defaultQEMUVolumeDev  = "/dev/sdb"
	defaultQEMUVolumeType = "qcow2"
)

type deployImageOption struct {
	Name          string `json:"name"`
	SizeMB        int64  `json:"size_mb"`
	StoragePoolId string `json:"storage_pool_id"`
}

type deployStoragePoolOption struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	AvailableGB float64 `json:"available_gb"`
}

type deployVolumeOption struct {
	Id            string  `json:"id"`
	Name          string  `json:"name"`
	StoragePoolId string  `json:"storage_pool_id"`
	SizeGB        float64 `json:"size_gb"`
}

type deployNetworkOption struct {
	VSwitchId     string `json:"vswitch_id"`
	VSwitchName   string `json:"vswitch_name"`
	PortGroupName string `json:"port_group_name"`
}

type deployQEMUOptions struct {
	Architectures []string `json:"architectures"`
	CPUModels     []string `json:"cpu_models"`
	MachineTypes  []string `json:"machine_types"`
}

type deployInstanceRecommendations struct {
	CPUCores       int     `json:"cpu_cores"`
	MemoryMB       int     `json:"memory_mb"`
	ExtraStorageGB float64 `json:"extra_storage_gb"`
	StoragePoolId  string  `json:"storage_pool_id,omitempty"`
	Architecture   string  `json:"architecture"`
	CPUModel       string  `json:"cpu_model,omitempty"`
	MachineType    string  `json:"machine_type,omitempty"`
	Firmware       string  `json:"firmware,omitempty"`
	VSwitchId      string  `json:"vswitch_id,omitempty"`
	PortGroupName  string  `json:"port_group_name,omitempty"`
}

type getDeployInstanceContextResult struct {
	NodeId          string                        `json:"node_id"`
	InstanceType    string                        `json:"instance_type"`
	Images          []deployImageOption           `json:"images"`
	StoragePools    []deployStoragePoolOption     `json:"storage_pools"`
	ExistingVolumes []deployVolumeOption          `json:"unattached_existing_volumes"`
	Networks        []deployNetworkOption         `json:"networks"`
	QEMUOptions     *deployQEMUOptions            `json:"qemu_options,omitempty"`
	Recommended     deployInstanceRecommendations `json:"recommended"`
	AskUserFor      []string                      `json:"ask_user_for"`
	OptionalChoices []string                      `json:"optional_customizations"`
	NextStep        string                        `json:"next_step"`
}

func GetDeployInstanceContext() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_deploy_instance_context",
		mcp.WithDescription("Gather valid deployment images and optional customization choices for a QEMU or LXC instance. By default, ask only for name, instance type, and image. CPU, memory, storage, and network are optional customizations. Do not deploy from this tool."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Deploy Instance Context",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Node on which the instance will be deployed (format: node-<xxx>)"),
		),
		mcp.WithString("instance_type",
			mcp.Required(),
			mcp.Enum("qemu", "lxc"),
			mcp.Description("Type of instance to deploy"),
		),
	), handleGetDeployInstanceContext
}

func handleGetDeployInstanceContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceTypeName, err := requiredParam[string](req, "instance_type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceType, err := deployInstanceType(instanceTypeName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
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
	vswitches, apiErr := api.RetrieveStandaloneVSwitches(ctx, client, &api.RetrieveStandaloneVSwitchesArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	portGroups, apiErr := api.RetrieveStandalonePortGroups(ctx, client, &api.RetrieveStandalonePortGroupsArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	capabilities, apiErr := api.RetrieveNodeVirtualizationCapabilities(ctx, client, &api.RetrieveNodeVirtualizationCapabilitiesArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}

	architecture, capability, err := recommendedCapability(*capabilities)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := getDeployInstanceContextResult{
		NodeId:          nodeId,
		InstanceType:    strings.ToLower(instanceTypeName),
		Images:          make([]deployImageOption, 0),
		StoragePools:    make([]deployStoragePoolOption, 0),
		ExistingVolumes: make([]deployVolumeOption, 0),
		Networks:        make([]deployNetworkOption, 0),
		Recommended: deployInstanceRecommendations{
			CPUCores:       defaultCPUCores,
			MemoryMB:       defaultMemoryMB,
			ExtraStorageGB: 0,
			Architecture:   architecture,
		},
		AskUserFor:      []string{"name", "image"},
		OptionalChoices: []string{"cpu_cores", "memory_mb", "extra_storage_gb", "network"},
		NextStep:        "Use the defaults unless the user asks to customize them. Show a summary, obtain explicit confirmation, then call deploy_instance.",
	}

	for _, image := range *images {
		if image.Type == instanceType {
			result.Images = append(result.Images, deployImageOption{
				Name:          image.Name,
				SizeMB:        image.SizeMB,
				StoragePoolId: image.StoragePoolId,
			})
		}
	}

	for _, pool := range *pools {
		if !pool.Initialized || !pool.Available {
			continue
		}
		result.StoragePools = append(result.StoragePools, deployStoragePoolOption{
			Id:          pool.Id,
			Name:        pool.Name,
			AvailableGB: pool.Usage.AvailableGB,
		})
	}

	for _, volume := range *volumes {
		if !volume.Attached {
			result.ExistingVolumes = append(result.ExistingVolumes, deployVolumeOption{
				Id:            volume.Id,
				Name:          volume.Name,
				StoragePoolId: volume.StoragePoolId,
				SizeGB:        volume.Size,
			})
		}
	}

	vswitchNames := make(map[string]string, len(*vswitches))
	for _, vswitch := range *vswitches {
		vswitchNames[vswitch.Id] = vswitch.Name
	}
	for _, portGroup := range *portGroups {
		result.Networks = append(result.Networks, deployNetworkOption{
			VSwitchId:     portGroup.VSwitchId,
			VSwitchName:   vswitchNames[portGroup.VSwitchId],
			PortGroupName: portGroup.Name,
		})
	}
	if instanceType == enum.InstanceTypeEnumQEMU {
		cpuModel := recommendedCPUModel(capability.CPUModels)
		machineType := capability.DefaultMachine
		if machineType == "" && len(capability.Machines) > 0 {
			machineType = capability.Machines[0].Name
		}
		result.Recommended.ExtraStorageGB = defaultQEMUStorageGB
		result.Recommended.CPUModel = cpuModel
		result.Recommended.MachineType = machineType
		result.Recommended.Firmware = defaultQEMUFirmware
		result.QEMUOptions = &deployQEMUOptions{
			Architectures: capabilityArchitectures(*capabilities),
			CPUModels:     capabilityCPUModels(capability),
			MachineTypes:  capabilityMachineTypes(capability),
		}
	}

	return mcp.NewToolResultJSON(&result)
}

type deployInstanceResult struct {
	Message      string `json:"message"`
	TaskId       string `json:"task_id"`
	InstanceType string `json:"instance_type"`
	Name         string `json:"name"`
}

func DeployInstance() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("deploy_instance",
		mcp.WithDescription("Deploy a QEMU or LXC instance using choices returned by get_deploy_instance_context. Use recommended defaults unless the user overrides them. Call only after showing a summary and receiving explicit confirmation."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Deploy Instance",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Target node id")),
		mcp.WithString("instance_type", mcp.Required(), mcp.Enum("qemu", "lxc"), mcp.Description("Instance type")),
		mcp.WithString("name", mcp.Required(), mcp.MinLength(3), mcp.MaxLength(255), mcp.Pattern("^[a-zA-Z0-9_-]+$"), mcp.Description("Instance name")),
		mcp.WithString("image", mcp.Required(), mcp.Description("Image name selected from deployment context")),
		mcp.WithString("vswitch_id", mcp.Description("Exact vSwitch id returned by deployment context. Omit it to resolve the vSwitch from port_group_name automatically; never use placeholders such as 'any'.")),
		mcp.WithString("port_group_name", mcp.Description("Optional port group selected from deployment context. Omit it to deploy with no network interfaces.")),
		mcp.WithNumber("cpu_cores", mcp.Min(1), mcp.DefaultNumber(defaultCPUCores), mcp.Description("CPU cores; Pextra default is 1")),
		mcp.WithNumber("memory_mb", mcp.Min(64), mcp.DefaultNumber(defaultMemoryMB), mcp.Description("Memory in MB; recommended default is 2048")),
		mcp.WithNumber("extra_storage_gb", mcp.Min(0), mcp.DefaultNumber(defaultQEMUStorageGB), mcp.Description("Optional new QEMU disk size in GiB. Pextra default is 0 (no volumes). LXC cannot add a new volume during deployment.")),
		mcp.WithString("storage_pool_id", mcp.Description("Storage pool from the context recommendation. This is chosen automatically if omitted.")),
		mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true only after the user explicitly confirms the deployment summary")),
	), handleDeployInstance
}

func handleDeployInstance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceTypeName, err := requiredParam[string](req, "instance_type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceType, err := deployInstanceType(instanceTypeName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := requiredParam[string](req, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	image, err := requiredParam[string](req, "image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	vswitchId, err := optionalParam[string](req, "vswitch_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if strings.EqualFold(vswitchId, "any") {
		vswitchId = ""
	}
	portGroupName, err := optionalParam[string](req, "port_group_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if portGroupName == "" && vswitchId != "" {
		return mcp.NewToolResultError("port_group_name is required when vswitch_id is provided"), nil
	}
	confirmed, err := req.RequireBool("confirm")
	if err != nil || !confirmed {
		return mcp.NewToolResultError("deployment requires explicit user confirmation"), nil
	}

	cpuCores := req.GetInt("cpu_cores", defaultCPUCores)
	memoryMB := req.GetInt("memory_mb", defaultMemoryMB)
	extraStorageGB := defaultQEMUStorageGB
	if _, ok := req.GetArguments()["extra_storage_gb"]; ok {
		extraStorageGB = req.GetFloat("extra_storage_gb", extraStorageGB)
	}
	if instanceType == enum.InstanceTypeEnumLXC && extraStorageGB > 0 {
		return mcp.NewToolResultError("extra storage cannot be created during LXC deployment; use 0 GiB"), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if portGroupName != "" {
		var apiErr *api.APIError
		vswitchId, apiErr = resolveStandaloneVSwitchId(ctx, client, nodeId, portGroupName, vswitchId)
		if apiErr != nil {
			return mcp.NewToolResultError(apiErr.Error()), nil
		}
	}
	capabilities, apiErr := api.RetrieveNodeVirtualizationCapabilities(ctx, client, &api.RetrieveNodeVirtualizationCapabilitiesArg{NodeId: nodeId})
	if apiErr != nil {
		return mcp.NewToolResultError(apiErr.Error()), nil
	}
	architecture, capability, err := recommendedCapability(*capabilities)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	deployArg := api.DeployInstanceV2Arg{
		Type:         instanceType,
		NodeId:       nodeId,
		Name:         name,
		Architecture: architecture,
		Image:        image,
		CPU: api.DeployInstanceCPU{
			Sockets: 1,
			Cores:   cpuCores,
			Threads: 1,
		},
		Memory:      memoryMB,
		IMDSEnabled: false,
		Networks:    []api.DeployInstanceNetwork{},
		Autostart:   true,
	}
	if portGroupName != "" {
		deployArg.Networks = append(deployArg.Networks, api.DeployInstanceNetwork{
			VSwitchId:     vswitchId,
			PortGroupName: portGroupName,
		})
	}

	if instanceType == enum.InstanceTypeEnumLXC {
		deployArg.LXCMetadata = &api.DeployLXCInstanceMetadata{Init: "/sbin/init"}
	} else {
		machineType := capability.DefaultMachine
		if machineType == "" && len(capability.Machines) > 0 {
			machineType = capability.Machines[0].Name
		}
		if len(deployArg.Networks) > 0 {
			deployArg.Networks[0].Model = defaultQEMUNetwork
		}
		deployArg.QEMUMetadata = &api.DeployQEMUInstanceMetadata{
			CPUModel:        recommendedCPUModel(capability.CPUModels),
			MachineType:     machineType,
			Firmware:        defaultQEMUFirmware,
			SecureBoot:      api.DeployQEMUSecureBoot{Enabled: false},
			NewVolumes:      []api.DeployQEMUNewVolume{},
			ExistingVolumes: []api.DeployQEMUExistingVolume{},
		}
		if extraStorageGB > 0 {
			storagePoolId, _ := optionalParam[string](req, "storage_pool_id")
			if storagePoolId == "" {
				pools, getErr := api.GetNodeStoragePoolsById(ctx, client, &api.GetNodeStoragePoolsByIdArg{NodeId: nodeId})
				if getErr != nil {
					return mcp.NewToolResultError(getErr.Error()), nil
				}
				storagePoolId = largestAvailablePool(*pools)
			}
			if storagePoolId == "" {
				return mcp.NewToolResultError("no available storage pool was found for the new QEMU volume"), nil
			}
			deployArg.QEMUMetadata.NewVolumes = append(deployArg.QEMUMetadata.NewVolumes, api.DeployQEMUNewVolume{
				Backup:        true,
				Bus:           defaultQEMUVolumeBus,
				Device:        defaultQEMUVolumeDev,
				Driver:        defaultQEMUVolumeType,
				StoragePoolId: storagePoolId,
				Size:          extraStorageGB,
			})
		}
	}

	response, deployErr := api.DeployInstanceV2(ctx, client, &deployArg)
	if deployErr != nil {
		return mcp.NewToolResultError(deployErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&deployInstanceResult{
		Message:      "Instance deployment started successfully",
		TaskId:       response.TaskId,
		InstanceType: strings.ToLower(instanceTypeName),
		Name:         name,
	})
}

func resolveStandaloneVSwitchId(
	ctx context.Context,
	client *api.Client,
	nodeId string,
	portGroupName string,
	requestedVSwitchId string,
) (string, *api.APIError) {
	portGroups, apiErr := api.RetrieveStandalonePortGroups(ctx, client, &api.RetrieveStandalonePortGroupsArg{NodeId: nodeId})
	if apiErr != nil {
		return "", apiErr
	}

	matches := make([]string, 0, 1)
	for _, portGroup := range *portGroups {
		if portGroup.Name != portGroupName {
			continue
		}
		if requestedVSwitchId != "" && portGroup.VSwitchId != requestedVSwitchId {
			continue
		}
		matches = append(matches, portGroup.VSwitchId)
	}

	if len(matches) == 0 {
		if requestedVSwitchId != "" {
			return "", api.NewAPIError(400, "the selected port group does not belong to the requested vSwitch on this node")
		}
		return "", api.NewAPIError(400, "the selected port group was not found on this node")
	}
	if len(matches) > 1 && requestedVSwitchId == "" {
		return "", api.NewAPIError(400, "multiple vSwitches contain this port group; provide the exact vswitch_id returned by deployment context")
	}

	return matches[0], nil
}

func deployInstanceType(value string) (enum.InstanceTypeEnum, error) {
	switch strings.ToLower(value) {
	case "qemu":
		return enum.InstanceTypeEnumQEMU, nil
	case "lxc":
		return enum.InstanceTypeEnumLXC, nil
	default:
		return 0, fmt.Errorf("instance_type must be qemu or lxc")
	}
}

func recommendedCapability(capabilities api.RetrieveNodeVirtualizationCapabilitiesResponse) (string, api.NodeVirtualizationCapability, error) {
	keys := make([]string, 0, len(capabilities))
	for key := range capabilities {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, preferred := range []string{"x86_64", "amd64"} {
		for _, key := range keys {
			capability := capabilities[key]
			if key == preferred || capability.Architecture == preferred {
				return capabilityArchitecture(key, capability), capability, nil
			}
		}
	}
	if len(keys) == 0 {
		return "", api.NodeVirtualizationCapability{}, fmt.Errorf("the node returned no virtualization capabilities")
	}
	capability := capabilities[keys[0]]
	return capabilityArchitecture(keys[0], capability), capability, nil
}

func capabilityArchitecture(key string, capability api.NodeVirtualizationCapability) string {
	if capability.Architecture != "" {
		return capability.Architecture
	}
	return key
}

func recommendedCPUModel(models []api.NodeCPUModel) string {
	for _, model := range models {
		if model.Name == "default" {
			return model.Name
		}
	}
	for _, model := range models {
		if model.Name == "host" {
			return model.Name
		}
	}
	for _, model := range models {
		if !model.Experimental {
			return model.Name
		}
	}
	if len(models) > 0 {
		return models[0].Name
	}
	return "default"
}

func capabilityArchitectures(capabilities api.RetrieveNodeVirtualizationCapabilitiesResponse) []string {
	values := make([]string, 0, len(capabilities))
	for key, capability := range capabilities {
		values = append(values, capabilityArchitecture(key, capability))
	}
	sort.Strings(values)
	return values
}

func capabilityCPUModels(capability api.NodeVirtualizationCapability) []string {
	values := make([]string, 0, len(capability.CPUModels))
	for _, model := range capability.CPUModels {
		if !model.Experimental {
			values = append(values, model.Name)
		}
	}
	return values
}

func capabilityMachineTypes(capability api.NodeVirtualizationCapability) []string {
	values := make([]string, 0, len(capability.Machines))
	for _, machine := range capability.Machines {
		values = append(values, machine.Name)
	}
	return values
}

func largestAvailablePool(pools api.GetNodeStoragePoolsByIdResponse) string {
	var selected string
	var available float64
	for _, pool := range pools {
		if pool.Initialized && pool.Available && (selected == "" || pool.Usage.AvailableGB > available) {
			selected = pool.Id
			available = pool.Usage.AvailableGB
		}
	}
	return selected
}

func recommendedPoolAvailable(pools api.GetNodeStoragePoolsByIdResponse, poolId string) float64 {
	for _, pool := range pools {
		if pool.Id == poolId {
			return pool.Usage.AvailableGB
		}
	}
	return 0
}
