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
	"net/url"
	"strconv"
)

type ListUsersInOrganizationByIdArg struct {
	OrganizationId string
}
type ListUsersInOrganizationByIdResponse = []UserList

func ListUsersInOrganizationById(ctx context.Context, c *Client, arg *ListUsersInOrganizationByIdArg) (*ListUsersInOrganizationByIdResponse, *APIError) {
	if arg == nil || arg.OrganizationId == "" {
		return nil, NewAPIError(400, "organization_id is required")
	}

	path := "/v1/users"
	query := make(url.Values)
	query.Set("organization_id", arg.OrganizationId)

	var resp ListUsersInOrganizationByIdResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type DeleteUserByIdArg struct {
	UserId string
}
type DeleteUserByIdResponse struct{}

func DeleteUserById(ctx context.Context, c *Client, arg *DeleteUserByIdArg) (*DeleteUserByIdResponse, *APIError) {
	if arg == nil || arg.UserId == "" {
		return nil, NewAPIError(400, "user_id is required")
	}

	path := c.ExpandPath("/v1/users/{user_id}", map[string]string{"user_id": arg.UserId})

	var resp DeleteUserByIdResponse
	if apiErr := c.Delete(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type InvalidateUserSessionsByIdArg struct {
	UserId            string
	InvalidateCurrent bool
}
type InvalidateUserSessionsByIdResponse struct{}

func InvalidateUserSessionsById(ctx context.Context, c *Client, arg *InvalidateUserSessionsByIdArg) (*InvalidateUserSessionsByIdResponse, *APIError) {
	if arg == nil || arg.UserId == "" {
		return nil, NewAPIError(400, "user_id is required")
	}

	path := c.ExpandPath("/v1/users/{user_id}/invalidate-sessions", map[string]string{"user_id": arg.UserId})

	payload, err := json.Marshal(map[string]string{"invalidate_current": strconv.FormatBool(arg.InvalidateCurrent)})
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp InvalidateUserSessionsByIdResponse
	if apiErr := c.Post(ctx, path, nil, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
