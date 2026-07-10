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

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetMe() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_me",
		mcp.WithDescription("Get details of the current authenticated user"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get my user details",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithOutputSchema[api.GetUserSessionResponse](),
	), handleGetMe
}

func handleGetMe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	me, getErr := api.GetUserSession(ctx, client)
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(me)
}
