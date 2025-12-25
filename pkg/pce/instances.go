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
