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
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

type ListDatacentersArg struct {
	OrganizationId string
}
type ListDatacentersResponse = []DatacenterList

func ListDatacenters(ctx context.Context, c *Client, arg *ListDatacentersArg) (*ListDatacentersResponse, *APIError) {
	if arg == nil || arg.OrganizationId == "" {
		return nil, NewAPIError(400, "organization_id is required")
	}

	path := "/v1/datacenters"
	query := make(url.Values)
	query.Set("organization_id", arg.OrganizationId)

	var resp ListDatacentersResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type GetDatacenterArg struct {
	DatacenterId string
}
type GetDatacenterResponse struct {
	Datacenter DatacenterFull `json:"datacenter"`
	Clusters   []ClusterList  `json:"clusters"`
}

func GetDatacenter(ctx context.Context, c *Client, arg *GetDatacenterArg) (*GetDatacenterResponse, *APIError) {
	if arg == nil || arg.DatacenterId == "" {
		return nil, NewAPIError(400, "datacenter_id is required")
	}

	path := c.ExpandPath("/v1/datacenters/{datacenter_id}", map[string]string{
		"datacenter_id": arg.DatacenterId,
	})

	var resp GetDatacenterResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type CreateDatacenterArg struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	OrganizationId string `json:"organization_id"`
}
type CreateDatacenterResponse struct {
	Id string `json:"id"`
}

func CreateDatacenter(ctx context.Context, c *Client, arg *CreateDatacenterArg) (*CreateDatacenterResponse, *APIError) {
	if arg == nil || arg.Name == "" || arg.OrganizationId == "" {
		return nil, NewAPIError(400, "name and organization_id are required")
	}

	path := "/v1/datacenters"
	payload, err := json.Marshal(arg)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp CreateDatacenterResponse
	if apiErr := c.Post(ctx, path, nil, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type UpdateDatacenterArg struct {
	DatacenterId string              `json:"-"`
	Name         string              `json:"name,omitempty"`
	Description  string              `json:"description,omitempty"`
	Location     *DatacenterLocation `json:"location,omitempty"`
}
type UpdateDatacenterResponse = DatacenterFull

func UpdateDatacenter(ctx context.Context, c *Client, arg *UpdateDatacenterArg) (*UpdateDatacenterResponse, *APIError) {
	if arg == nil || arg.DatacenterId == "" {
		return nil, NewAPIError(400, "datacenter_id is required")
	}

	path := c.ExpandPath("/v1/datacenters/{datacenter_id}", map[string]string{
		"datacenter_id": arg.DatacenterId,
	})

	payload, err := json.Marshal(arg)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp UpdateDatacenterResponse
	if apiErr := c.Patch(ctx, path, nil, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type DestroyDatacenterArg struct {
	DatacenterId string `json:"-"`
}
type DestroyDatacenterResponse struct{}

func DestroyDatacenter(ctx context.Context, c *Client, arg *DestroyDatacenterArg) (*DestroyDatacenterResponse, *APIError) {
	if arg == nil || arg.DatacenterId == "" {
		return nil, NewAPIError(400, "datacenter_id is required")
	}

	path := c.ExpandPath("/v1/datacenters/{datacenter_id}", map[string]string{
		"datacenter_id": arg.DatacenterId,
	})

	var resp DestroyDatacenterResponse
	if apiErr := c.Delete(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
