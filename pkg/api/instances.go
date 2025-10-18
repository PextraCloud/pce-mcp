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

	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
)

type GetInstancesByIdArg struct {
	// Either `NodeId` or `ClusterId` must be provided
	NodeId    string
	ClusterId string
}
type GetInstancesByIdResponse = []InstanceList

func GetInstancesById(ctx context.Context, c *Client, arg *GetInstancesByIdArg) (*GetInstancesByIdResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "either node_id or cluster_id is required")
	}
	if arg.NodeId != "" && arg.ClusterId != "" {
		return nil, NewAPIError(400, "only one of node_id or cluster_id should be provided")
	}

	query := make(url.Values)
	if arg.NodeId != "" {
		query.Set("node_id", arg.NodeId)
	} else if arg.ClusterId != "" {
		query.Set("cluster_id", arg.ClusterId)
	} else {
		return nil, NewAPIError(400, "either node_id or cluster_id is required")
	}

	path := "/v1/instances"

	var resp GetInstancesByIdResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}

type PowerInstanceArg struct {
	InstanceId string
	NodeId     string
	Action     enum.InstancePowerAction
}
type PowerInstanceResponse struct {
	TaskId string `json:"task_id"`
}

func PowerInstance(ctx context.Context, c *Client, arg *PowerInstanceArg) (*PowerInstanceResponse, *APIError) {
	if arg == nil || arg.NodeId == "" || arg.InstanceId == "" {
		return nil, NewAPIError(400, "node_id and instance_id are required")
	}
	if !enum.InstancePowerAction.IsValid(arg.Action) {
		return nil, NewAPIError(400, "invalid action")
	}

	path := c.ExpandPath("/v1/instances/{instance_id}/power", map[string]string{"instance_id": arg.InstanceId})

	query := make(url.Values)
	query.Set("node_id", arg.NodeId)

	payload, err := json.Marshal(map[string]string{"action": string(arg.Action)})
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	var resp PowerInstanceResponse
	if apiErr := c.Post(ctx, path, query, bytes.NewReader(payload), &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
