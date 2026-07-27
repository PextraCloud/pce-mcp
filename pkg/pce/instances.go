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

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type getInstancesInNodeOrClusterResult struct {
	Instances *api.GetInstancesByIdResponse `json:"instances"`
}

type powerInstanceResult struct {
	Message string `json:"message"`
	TaskId  string `json:"task_id"`
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
		mcp.WithOutputSchema[powerInstanceResult](),
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

	return mcp.NewToolResultJSON(&powerInstanceResult{
		Message: "Power action initiated successfully",
		TaskId:  res.TaskId,
	})
}

func SearchInstances() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("search_instances_in_cluster",
		mcp.WithDescription("List instances in a specific cluster with optional filtering based on name, vcpus, memory, autostart status, and node id; also use when you want to find instances in a cluster or node"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Search instances in cluster",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Description("Cluster to search for instances in (format: cls-<xxx>); if not specified, the current cluster is used"),
		),
		mcp.WithString("node_id",
			mcp.Description("Node to search for instances in (format: node-<xxx>); if not specified, all nodes in the specified cluster are searched, takes precedence over cluster_id"),
		),
		mcpToolOptionStringFilter("name", "instance names"),
		mcpToolOptionNumberFilter("vcpus", "number of vCPUs"),
		mcpToolOptionNumberFilter("memory", "memory size in MB"),
		mcpToolOptionBoolFilter("autostart", "whether the instance is set to autostart on boot"),
		mcp.WithOutputSchema[getInstancesInNodeOrClusterResult](),
	), handleSearchInstances
}

func handleSearchInstances(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type reqType struct {
		ClusterId string         `json:"cluster_id"`
		NodeId    string         `json:"node_id"`
		Name      string         `json:"name"`
		Vcpus     map[string]any `json:"vcpus"`
		Memory    map[string]any `json:"memory"`
		Autostart *bool          `json:"autostart,omitempty"`
	}

	args := &reqType{}
	if err := req.BindArguments(args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	clusterId := args.ClusterId
	nodeId := args.NodeId
	nameFilter := args.Name
	vcpusFilterAny := args.Vcpus
	memoryFilterAny := args.Memory
	autostartFilter := args.Autostart
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
