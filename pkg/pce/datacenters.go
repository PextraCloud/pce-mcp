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

type ListDatacentersResult struct {
	Datacenters api.ListDatacentersResponse `json:"datacenters"`
}

func ListDatacenters() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_datacenters",
		mcp.WithDescription("List all datacenters in an organization"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List datacenters",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("organization_id",
			mcp.Description("If specified, the organization to list datacenters for (format: org-<xxx>), otherwise the current organization is used"),
		),
		mcp.WithOutputSchema[ListDatacentersResult](),
	), handleListDatacenters
}

func handleListDatacenters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	organizationId, _ := optionalParam[string](req, "organization_id")

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// If no organization ID is provided, use the current user's organization ID
	if organizationId == "" {
		// TODO: caching
		me, getErr := api.GetUserSession(ctx, client)
		if getErr != nil {
			return mcp.NewToolResultError(getErr.Error()), nil
		}
		organizationId = me.User.OrganizationId
	}

	datacenters, getErr := api.ListDatacenters(ctx, client, &api.ListDatacentersArg{
		OrganizationId: organizationId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(ListDatacentersResult{
		Datacenters: *datacenters,
	})
}

func GetDatacenter() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_datacenter",
		mcp.WithDescription("Get details of a specific datacenter, including its clusters and their details"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get datacenter",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("datacenter_id",
			mcp.Required(),
			mcp.Description("Unique datacenter id (format: dc-<xxx>)"),
		),
		mcp.WithOutputSchema[api.GetDatacenterResponse](),
	), handleGetDatacenter
}

func handleGetDatacenter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	datacenterId, err := requiredParam[string](req, "datacenter_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	datacenter, getErr := api.GetDatacenter(ctx, client, &api.GetDatacenterArg{
		DatacenterId: datacenterId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(datacenter)
}

func CreateDatacenter() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("create_datacenter",
		mcp.WithDescription("Create a new datacenter in an organization; to set the location, use the update_datacenter tool after creation"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title: "Create datacenter",
		}),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the new datacenter"),
		),
		mcp.WithString("description",
			mcp.Description("Optional description of the new datacenter"),
		),
		mcp.WithString("organization_id",
			mcp.Required(),
			mcp.Description("Organization to create the datacenter in (format: org-<xxx>); if not specified, the current organization is used"),
		),
		mcp.WithOutputSchema[api.CreateDatacenterResponse](),
	), handleCreateDatacenter
}

func handleCreateDatacenter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := requiredParam[string](req, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	description, _ := optionalParam[string](req, "description")

	organizationId, err := requiredParam[string](req, "organization_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	datacenter, createErr := api.CreateDatacenter(ctx, client, &api.CreateDatacenterArg{
		Name:           name,
		Description:    description,
		OrganizationId: organizationId,
	})
	if createErr != nil {
		return mcp.NewToolResultError(createErr.Error()), nil
	}

	return mcp.NewToolResultJSON(datacenter)
}

func UpdateDatacenter() (mcp.Tool, server.ToolHandlerFunc) {
	// NOTE: this tool is referred to by the "create_datacenter" tool as the way to set the location of a datacenter after creation
	return mcp.NewTool("update_datacenter",
		mcp.WithDescription("Update the name, description, and/or location of a datacenter"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title: "Update datacenter",
		}),
		mcp.WithString("datacenter_id",
			mcp.Required(),
			mcp.Description("Unique datacenter id (format: dc-<xxx>)"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the datacenter, if not specified, the name will not be changed"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the datacenter, if not specified, the description will not be changed"),
		),
		mcp.WithObject("location",
			mcp.Description("New location for the datacenter, if not specified, the location will not be changed. If specified, both latitude and longitude must be provided "),
			mcp.Properties(map[string]any{
				"latitude":  map[string]any{"type": "number", "description": "Latitude of the datacenter location"},
				"longitude": map[string]any{"type": "number", "description": "Longitude of the datacenter location"},
			}),
		),
	), handleUpdateDatacenter
}

func handleUpdateDatacenter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	datacenterId, err := requiredParam[string](req, "datacenter_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	name, _ := optionalParam[string](req, "name")
	description, _ := optionalParam[string](req, "description")
	location, _ := optionalParam[map[string]any](req, "location")

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var locationPtr *api.DatacenterLocation
	if location != nil {
		// Validate that both latitude and longitude are provided
		latVal, latOk := location["latitude"]
		lonVal, lonOk := location["longitude"]
		if !latOk || !lonOk {
			return mcp.NewToolResultError("both latitude and longitude must be provided when specifying location"), nil
		}

		// Validate that latitude and longitude can be cast to float64
		lat, latOk := latVal.(float64)
		lon, lonOk := lonVal.(float64)
		if !latOk || !lonOk {
			return mcp.NewToolResultError("latitude and longitude must be numbers"), nil
		}

		locationPtr = &api.DatacenterLocation{
			Latitude:  lat,
			Longitude: lon,
		}
	}

	_, updateErr := api.UpdateDatacenter(ctx, client, &api.UpdateDatacenterArg{
		DatacenterId: datacenterId,
		Name:         name,
		Description:  description,
		Location:     locationPtr,
	})
	if updateErr != nil {
		return mcp.NewToolResultError(updateErr.Error()), nil
	}

	return mcp.NewToolResultText("Datacenter updated successfully"), nil
}

func DestroyDatacenter() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("destroy_datacenter",
		mcp.WithDescription("Destroy a datacenter, this will fail if there are any clusters in the datacenter"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Destroy datacenter",
			DestructiveHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("datacenter_id",
			mcp.Required(),
			mcp.Description("Unique datacenter id (format: dc-<xxx>)"),
		),
		mcpToolOptionDestroyConfirmation,
	), handleDestroyDatacenter
}

func handleDestroyDatacenter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	datacenterId, err := requiredParam[string](req, "datacenter_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !confirmDestructiveAction(req) {
		return mcp.NewToolResultError("destructive action not confirmed"), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, destroyErr := api.DestroyDatacenter(ctx, client, &api.DestroyDatacenterArg{
		DatacenterId: datacenterId,
	})
	if destroyErr != nil {
		return mcp.NewToolResultError(destroyErr.Error()), nil
	}

	return mcp.NewToolResultText("Datacenter destroyed successfully"), nil
}
