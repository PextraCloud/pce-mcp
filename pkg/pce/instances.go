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
package pce

import (
	"context"
	"fmt"

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const instancesHelpText = `\n\nInstances are virtual machines (QEMU/KVM) or containers (LXC) that run on nodes within clusters.
They utilize the compute resources of the nodes to perform various tasks and services.` + hierarchyHelpText

type getInstancesInNodeOrClusterResult struct {
	Instances *api.GetInstancesByIdResponse `json:"instances"`
}

func GetInstancesInCluster() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_instances_in_cluster",
		mcp.WithDescription(fmt.Sprintf("Retrieve instances deployed on all nodes within a specific cluster%s", instancesHelpText)),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Instances in Cluster",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Required(),
			mcp.Description("Unique cluster id (format: cls-<xxx>)"),
		),
	), handleGetInstancesInCluster
}

func handleGetInstancesInCluster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterId, err := requiredParam[string](req, "cluster_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	instances, getErr := api.GetInstancesById(ctx, client, &api.GetInstancesByIdArg{
		ClusterId: clusterId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&getInstancesInNodeOrClusterResult{
		Instances: instances,
	})
}

func GetInstancesInNode() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_instances_in_node",
		mcp.WithDescription(fmt.Sprintf("Retrieve instances deployed on a specific node%s", instancesHelpText)),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Instances in Node",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetInstancesInNode
}

func handleGetInstancesInNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	instances, getErr := api.GetInstancesById(ctx, client, &api.GetInstancesByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&getInstancesInNodeOrClusterResult{
		Instances: instances,
	})
}

func PowerInstance() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("power_instance",
		mcp.WithDescription("Perform a power action on a specific instance (start, stop, restart, kill)"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Power Instance",
			ReadOnlyHint: mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("Unique instance id (format: inst-<xxx>)"),
		),
		mcp.WithString("action",
			mcp.Enum(
				string(enum.InstancePowerActionStart),
				string(enum.InstancePowerActionStop),
				string(enum.InstancePowerActionRestart),
				string(enum.InstancePowerActionKill),
			),
			mcp.Required(),
			mcp.Description("Power action to perform. start = power on, stop = graceful shutdown (may do nothing if the guest OS does not support it), restart = graceful reboot, kill = immediate power off, like pulling the power plug"),
		),
	), handlePowerInstance
}

func handlePowerInstance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceId, err := requiredParam[string](req, "instance_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Using `string` instead of  `enum.InstancePowerAction` since `requiredParam`` does not support enum types
	action, err := requiredParam[string](req, "action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	res, powerErr := api.PowerInstance(ctx, client, &api.PowerInstanceArg{
		NodeId:     nodeId,
		InstanceId: instanceId,
		Action:     enum.InstancePowerAction(action),
	})
	if powerErr != nil {
		return mcp.NewToolResultError(powerErr.Error()), nil
	}

	return mcp.NewToolResultJSON(struct {
		Message string `json:"message"`
		TaskId  string `json:"task_id"`
	}{
		Message: "Power action initiated successfully",
		TaskId:  res.TaskId,
	})
}

func SearchInstances() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("search_instances_in_cluster",
		mcp.WithDescription("Search for instances in a specific cluster based on name, instance type, vcpus, and memory; returns all instances in the cluster if no filters are specified"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Search instances in cluster",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Description("Cluster to search for instances in (format: cls-<xxx>); if not specified, the current cluster is used"),
		),
		mcp.WithString("node_id",
			mcp.Description("Node to search for instances in (format: node-<xxx>); if not specified, all nodes in the specified cluster are searched"),
		),
		mcpToolOptionStringFilter("name", "instance names"),
		mcpToolOptionNumberFilter("vcpus", "number of vCPUs"),
		mcpToolOptionNumberFilter("memory", "memory size in MB"),
		mcpToolOptionBoolFilter("autostart", "whether the instance is set to autostart on boot"),
		mcp.WithOutputSchema[getInstancesInNodeOrClusterResult](),
	), handleSearchInstances
}

func handleSearchInstances(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterId, _ := optionalParam[string](req, "cluster_id")
	nodeId, _ := optionalParam[string](req, "node_id")
	nameFilter, _ := optionalParam[string](req, "name")
	vcpusFilterAny, _ := optionalParam[map[string]any](req, "vcpus")
	memoryFilterAny, _ := optionalParam[map[string]any](req, "memory")
	autostartFilter, _ := optionalParamPtr[bool](req, "autostart")

	vcpusFilter := convertFilterToInt(vcpusFilterAny)
	memoryFilter := convertFilterToInt(memoryFilterAny)

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// If no cluster ID is provided, use the current cluster ID
	if clusterId == "" {
		currentIds, err := getCurrentTreeIds(ctx, client)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		clusterId = currentIds.ClusterId
	}

	// If both specified, nodeId takes precedence over clusterId
	searchArg := &api.GetInstancesByIdArg{}
	if nodeId != "" {
		searchArg.NodeId = nodeId
	} else {
		// clusterId is guaranteed to be non-empty here due to above use of currentTreeIds if clusterId is not provided
		searchArg.ClusterId = clusterId
	}

	instances, getErr := api.GetInstancesById(ctx, client, searchArg)
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	filteredInstances := filterInstances(instances, nameFilter, vcpusFilter, memoryFilter, autostartFilter)
	return mcp.NewToolResultJSON(&getInstancesInNodeOrClusterResult{
		Instances: filteredInstances,
	})
}

// Perform filtering here since the PCE API does not support such filters (yet)
func filterInstances(instances *api.GetInstancesByIdResponse, nameFilter string, vcpusFilter map[string]int, memoryFilter map[string]int, autostartFilter *bool) *api.GetInstancesByIdResponse {
	var filtered api.GetInstancesByIdResponse
	for _, instance := range *instances {
		if !applyStringFilter(instance.Name, nameFilter) {
			continue
		}
		if !applyNumberFilter(instance.Vcpus, vcpusFilter) {
			continue
		}
		if !applyNumberFilter(instance.Memory, memoryFilter) {
			continue
		}
		if !applyBoolFilter(instance.Autostart, autostartFilter) {
			continue
		}

		filtered = append(filtered, instance)
	}

	return &filtered
}
