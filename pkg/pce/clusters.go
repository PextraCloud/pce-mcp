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

	"github.com/PextraCloud/pce-mcp/internal/session"
	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetClusterHardwareById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_cluster_hardware_by_id",
		mcp.WithDescription("Retrieve aggregated hardware information about all nodes in a specific cluster"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Cluster Hardware By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("cluster_id",
			mcp.Required(),
			mcp.Description("Unique cluster id (format: cls-<xxx>)"),
		),
	), handleGetClusterHardwareById
}

func handleGetClusterHardwareById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterId, err := requiredParam[string](req, "cluster_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
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
	), handleGetClusterLicensingById
}

func handleGetClusterLicensingById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterId, err := requiredParam[string](req, "cluster_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
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
