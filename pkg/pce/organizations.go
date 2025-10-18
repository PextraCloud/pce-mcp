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

const organizationsHelpText = `\n\nOrganizations are the top-level entities within the Pextra CloudEnvironment (PCE) hierarchy.
They represent distinct tenants within the cloud, each with its own users, storage, network configurations, and compute resources.` + hierarchyHelpText

func ListOrganizations() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_organizations",
		mcp.WithDescription(fmt.Sprintf("Retrieve a list of all organizations accessible to the user%s", organizationsHelpText)),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Organizations",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	), handleListOrganizations
}

func handleListOrganizations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	orgs, listErr := api.ListOrganizations(ctx, client, &api.ListOrganizationsArg{})
	if listErr != nil {
		return mcp.NewToolResultError(listErr.Error()), nil
	}

	return mcp.NewToolResultJSON(orgs)
}

func GetOrganizationById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_organization_by_id",
		mcp.WithDescription("Retrieve detailed information about a specific organization. This includes the datacenters, clusters, and nodes within the organization. Use this tool when asked for a tree/hierarchical view of infrastructure. Use this tool if a list of nodes, clusters, or datacenters is needed."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Organization By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Unique organization id (format: org-<xxx>)"),
		),
	), handleGetOrganizationById
}

func handleGetOrganizationById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	orgId, err := requiredParam[string](req, "organization_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var client *api.Client
	if client, err = session.GetSession("sessionId"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	org, getErr := api.GetOrganizationById(ctx, client, &api.GetOrganizationByIdArg{
		OrganizationId: orgId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(org)
}

func GetCurrentOrganization() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_current_organization",
		mcp.WithDescription("Retrieve detailed information about the organization associated with the current node. Use this tool when asked for a tree/hierarchical view of infrastructure for the current organization."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Current Organization",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	), handleGetCurrentOrganization
}

func handleGetCurrentOrganization(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Retrieve organization using organization ID from node
	organizationId := node.Node.OrganizationId
	org, orgErr := api.GetOrganizationById(ctx, client, &api.GetOrganizationByIdArg{
		OrganizationId: organizationId,
	})
	if orgErr != nil {
		return mcp.NewToolResultError(orgErr.Error()), nil
	}
	return mcp.NewToolResultJSON(org)
}

/*func ListOrganizationAuditLogsById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_organization_audit_logs_by_id",
		mcp.WithDescription("Retrieve audit logs for a specific organization. Audit logs provide a record of actions and events that have occurred within the organization, useful for tracking changes and ensuring compliance."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Organization Audit Logs By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Unique organization id (format: org-<xxx>)"),
		),
		mcp.WithNumber("entries",
			mcp.Min(0),
			mcp.Max(5000),
			mcp.DefaultNumber(50),
			mcp.Description("The number of audit log entries to retrieve. Default is 50."),
		),
		pageNum,
	), handleListOrganizationAuditLogsById
}

func ListOrganizationUserLockoutsById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_organization_user_lockouts_by_id",
		mcp.WithDescription("Retrieve a list of user lockouts for a specific organization. User lockouts occur when users are temporarily prevented from accessing their accounts due to multiple failed login attempts or security policies."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Organization User Lockouts By ID",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Unique organization id (format: org-<xxx>)"),
		),
		mcp.WithNumber("entries",
			mcp.Min(0),
			mcp.Max(5000),
			mcp.DefaultNumber(50),
			mcp.Description("The number of user lockout entries to retrieve. Default is 50."),
		),
		pageNum,
	), handleListOrganizationUserLockoutsById
}*/

func CreateOrganization() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("create_organization",
		mcp.WithDescription("Create a new organization within the Pextra CloudEnvironment (PCE). Organizations are top-level entities that represent distinct tenants within the cloud."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title: "Create Organization",
		}),
		mcp.WithString("name",
			mcp.Required(),
			mcp.MinLength(nameDefaultMinLength),
			mcp.MaxLength(nameDefaultMaxLength),
			mcp.Pattern(nameRegex(nameDefaultMinLength, nameDefaultMaxLength)), // redundant, but being explicit
			mcp.Description("The name of the new organization."),
		),
		mcp.WithString("description",
			mcp.MaxLength(descriptionDefaultMaxLength),
			mcp.Description("A brief description of the organization."),
		),
	), handleCreateOrganization
}

func handleCreateOrganization(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := requiredParam[string](req, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	description, err := optionalParam[string](req, "description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	org, createErr := api.CreateOrganization(ctx, client, &api.CreateOrganizationArg{
		Name:        name,
		Description: description,
	})
	if createErr != nil {
		return mcp.NewToolResultError(createErr.Error()), nil
	}

	return mcp.NewToolResultJSON(org)
}

func DeleteOrganizationById() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("delete_organization_by_id",
		mcp.WithDescription("Delete an existing organization. The organization must be empty of any datacenters (and consequently clusters and nodes) before it can be deleted."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Delete Organization By ID",
			DestructiveHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Unique organization id (format: org-<xxx>)"),
		),
		mcp.WithBoolean("are_you_sure",
			mcp.Required(),
			mcp.Description("A safety check to prevent accidental deletions. Must be set to true to proceed with deletion."),
		),
	), handleDeleteOrganizationById
}

func handleDeleteOrganizationById(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	orgId, err := requiredParam[string](req, "organization_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	areYouSure, err := requiredParam[bool](req, "are_you_sure")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !areYouSure {
		return mcp.NewToolResultError("Deletion not confirmed. Set 'are_you_sure' to true to proceed."), nil
	}

	client, err := session.GetSession("sessionId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, deleteErr := api.DeleteOrganizationById(ctx, client, &api.DeleteOrganizationByIdArg{
		OrganizationId: orgId,
	})
	if deleteErr != nil {
		return mcp.NewToolResultError(deleteErr.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Organization %s deleted successfully.", orgId)), nil
}
