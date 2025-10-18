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

func GetImages() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_images",
		mcp.WithDescription("Retrieve images for instance deployment on a specific node"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Images",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Unique node id (format: node-<xxx>)"),
		),
	), handleGetImages
}

type getImagesResult struct {
	Images *api.ListImagesByNodeResponse `json:"images"`
}

func handleGetImages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var client *api.Client
	if client, err = session.GetSession("sessionId"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	images, listErr := api.ListImagesByNode(ctx, client, &api.ListImagesByNodeArg{
		NodeId: nodeId,
	})
	if listErr != nil {
		return mcp.NewToolResultError(listErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&getImagesResult{
		Images: images,
	})
}
