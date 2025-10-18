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

	"github.com/PextraCloud/pce-mcp/internal/session"
	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const nodesHelpText = `\n\nNodes are servers that provide the compute resources within Pextra CloudEnvironment (PCE). They are organized in clusters, and host instances (virtual machines or containers) that run workloads.` + hierarchyHelpText

func GetNodeById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_node_by_id",
		mcp.WithDescription(fmt.Sprintf("Retrieve detailed information about a specific node%s", nodesHelpText)),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Node By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetNodeById
}

func handleGetNodeById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	node, getErr := api.GetNodeById(ctx, client, &api.GetNodeByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(node)
}

func GetCurrentNode() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_current_node",
		mcp.WithDescription("Retrieve detailed information about the current node"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Current Node",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	), handleGetCurrentNode
}

func handleGetCurrentNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := session.GetSession("sessionId")
	if err != nil {
		fmt.Println("Error retrieving session:", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get node ID through healthcheck endpoint
	health, healthErr := api.RunHealthcheck(ctx, client, &api.RunHealthcheckArg{})
	if healthErr != nil {
		return mcp.NewToolResultError(healthErr.Error()), nil
	}
	// This should never happen
	if !health.Healthy {
		return mcp.NewToolResultError("Node is not healthy"), nil
	}

	nodeId := health.Id
	node, getErr := api.GetNodeById(ctx, client, &api.GetNodeByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(node)
}

func GetNodeHardwareById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_node_hardware_by_id",
		mcp.WithDescription("Retrieve detailed hardware information (CPU, memory, disks, USB devices) about a specific node."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Node Hardware By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetNodeHardwareById
}

func handleGetNodeHardwareById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	hardware, getErr := api.GetNodeHardwareById(ctx, client, &api.GetNodeHardwareByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(hardware)
}

func GetNodeLicenseById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_node_license_by_id",
		mcp.WithDescription("Retrieve license information about a specific node."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Node License By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
		mcp.WithBoolean("include_key",
			mcp.Description("Whether to include the license key in the response. This may expose sensitive information."),
			mcp.DefaultBool(false),
		),
	), handleGetNodeLicenseById
}

func handleGetNodeLicenseById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	includeKey, err := optionalParam[bool](req, "include_key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	license, getErr := api.GetNodeLicenseById(ctx, client, &api.GetNodeByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	// Redact license key if not explicitly requested
	if !includeKey {
		license.Key = "REDACTED"
	}
	return mcp.NewToolResultJSON(license)
}

func GetNodeStoragePoolsById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_node_storagepools_by_id",
		mcp.WithDescription("Retrieve storage pool information for a specific node."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Node Storage Pools By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetNodeStoragePoolsById
}

type getNodeStoragePoolsByIdResult struct {
	Pools *api.GetNodeStoragePoolsByIdResponse `json:"pools"`
}

func handleGetNodeStoragePoolsById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pools, getErr := api.GetNodeStoragePoolsById(ctx, client, &api.GetNodeStoragePoolsByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&getNodeStoragePoolsByIdResult{
		Pools: pools,
	})
}

func GetNodePciDevicesById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_node_pcidevices_by_id",
		mcp.WithDescription("Retrieve PCI device information for a specific node."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Node PCI Devices By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetNodePciDevicesById
}

type getNodePciDevicesByIdResult struct {
	Devices *api.GetNodePciDevicesByIdResponse `json:"devices"`
}

func handleGetNodePciDevicesById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	devices, getErr := api.GetNodePciDevicesById(ctx, client, &api.GetNodePciDevicesByIdArg{
		NodeId: nodeId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&getNodePciDevicesByIdResult{
		Devices: devices,
	})
}
