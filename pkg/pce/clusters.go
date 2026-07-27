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
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetCluster() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_cluster",
		mcp.WithDescription("Get detailed information about a specific cluster, including its nodes and their details"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get cluster",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Description("If specified, the cluster to get information for (format: cls-<xxx>), otherwise the current cluster is used"),
		),
		mcp.WithOutputSchema[api.GetClusterByIdResponse](),
	), handleGetCluster
}

func handleGetCluster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type reqType struct {
		ClusterId string `json:"cluster_id"`
	}

	args := &reqType{}
	if err := req.BindArguments(args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	clusterId := args.ClusterId

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

	cluster, getErr := api.GetClusterById(ctx, client, &api.GetClusterByIdArg{
		ClusterId: clusterId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(cluster)
}

func GetClusterHardware() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_cluster_hardware",
		mcp.WithDescription("Get aggregated compute, memory, and storage capacity information for all nodes in a specific cluster"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get cluster hardware",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Description("If specified, the cluster to get hardware information for (format: cls-<xxx>), otherwise the current cluster is used"),
		),
		mcp.WithOutputSchema[api.GetClusterHardwareByIdResponse](),
	), handleGetClusterHardware
}

func handleGetClusterHardware(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type reqType struct {
		ClusterId string `json:"cluster_id"`
	}

	args := &reqType{}
	if err := req.BindArguments(args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	clusterId := args.ClusterId

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

	hardware, getErr := api.GetClusterHardwareById(ctx, client, &api.GetClusterHardwareByIdArg{
		ClusterId: clusterId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(hardware)
}

func GetClusterLicensingById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_cluster_licensing_by_id",
		mcp.WithDescription("Retrieve aggregated licensing information about all nodes in a specific cluster. Only nodes that have licensing information will be included in the response. There is no separate 'cluster license', licenses only exist for individual nodes."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Cluster License By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Required(),
			mcp.Description("Unique cluster id (format: cls-<xxx>)"),
		),
		mcp.WithOutputSchema[api.GetClusterLicensingByIdResponse](),
	), handleGetClusterLicensingById
}

func handleGetClusterLicensingById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterId, err := requiredParam[string](req, "cluster_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	license, getErr := api.GetClusterLicensingById(ctx, client, &api.GetClusterLicensingByIdArg{
		ClusterId: clusterId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}
	return mcp.NewToolResultJSON(license)
}
