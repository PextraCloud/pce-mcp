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
package api

import (
	"bytes"
	"context"
	"encoding/json"
)

type ListOrganizationsArg struct{}
type ListOrganizationsResponse = []OrganizationDetail

func ListOrganizations(ctx context.Context, c *Client, arg *ListOrganizationsArg) (*ListOrganizationsResponse, *APIError) {
	path := c.ExpandPath("/v1/organizations", nil)

	var resp ListOrganizationsResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetOrganizationByIdArg struct {
	OrganizationId string
}
type GetOrganizationByIdResponse = OrganizationDetail

func GetOrganizationById(ctx context.Context, c *Client, arg *GetOrganizationByIdArg) (*GetOrganizationByIdResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "organization_id is required")
	}

	path := c.ExpandPath("/v1/organizations/{organization_id}", map[string]string{"organization_id": arg.OrganizationId})

	var resp GetOrganizationByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type CreateOrganizationArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
type CreateOrganizationResponse = struct {
	Id string `json:"id"`
}

func CreateOrganization(ctx context.Context, c *Client, arg *CreateOrganizationArg) (*CreateOrganizationResponse, *APIError) {
	if arg == nil || arg.Name == "" {
		return nil, NewAPIError(400, "name is required")
	}

	path := c.ExpandPath("/v1/organizations", nil)

	payload, err := json.Marshal(map[string]string{"name": arg.Name, "description": arg.Description})
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp CreateOrganizationResponse
	if apiErr := c.Post(ctx, path, nil, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type DeleteOrganizationByIdArg struct {
	OrganizationId string
}
type DeleteOrganizationByIdResponse = struct{}

func DeleteOrganizationById(ctx context.Context, c *Client, arg *DeleteOrganizationByIdArg) (*DeleteOrganizationByIdResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "organization_id is required")
	}

	path := c.ExpandPath("/v1/organizations/{organization_id}", map[string]string{"organization_id": arg.OrganizationId})

	var resp DeleteOrganizationByIdResponse
	if apiErr := c.Delete(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
