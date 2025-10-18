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

func ListUsersInOrganizationById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_users_in_organization_by_id",
		mcp.WithDescription("List all users in a specific organization by its ID"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Users In Organization By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Unique organization id (format: org-<xxx>)"),
		),
	), handleListUsersInOrganizationById
}

type listUsersInOrganizationByIdResult struct {
	Users *api.ListUsersInOrganizationByIdResponse `json:"users"`
}

func handleListUsersInOrganizationById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	organizationId, err := requiredParam[string](req, "organization_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	users, getErr := api.ListUsersInOrganizationById(ctx, client, &api.ListUsersInOrganizationByIdArg{
		OrganizationId: organizationId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(&listUsersInOrganizationByIdResult{
		Users: users,
	})
}

func DeleteUserById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("delete_user_by_id",
		mcp.WithDescription("Delete a user by its ID"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Delete User By ID",
			DestructiveHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("Unique user id (format: user-<xxx>)"),
		),
	), handleDeleteUserById
}

func handleDeleteUserById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := requiredParam[string](req, "user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, deleteErr := api.DeleteUserById(ctx, client, &api.DeleteUserByIdArg{
		UserId: userId,
	})
	if deleteErr != nil {
		return mcp.NewToolResultError(deleteErr.Error()), nil
	}

	return mcp.NewToolResultText("User deleted successfully"), nil
}

func InvalidateUserSessionsById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("invalidate_user_sessions_by_id",
		mcp.WithDescription("Invalidate all sessions for a specific user by its ID"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Invalidate User Sessions By ID",
			DestructiveHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("Unique user id (format: user-<xxx>)"),
		),
	), handleInvalidateUserSessionsById
}

func handleInvalidateUserSessionsById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := requiredParam[string](req, "user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, invalidateErr := api.InvalidateUserSessionsById(ctx, client, &api.InvalidateUserSessionsByIdArg{
		UserId:            userId,
		InvalidateCurrent: false,
	})
	if invalidateErr != nil {
		return mcp.NewToolResultError(invalidateErr.Error()), nil
	}

	return mcp.NewToolResultText("User sessions invalidated successfully"), nil
}
