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
	"context"
)

type RunHealthcheckArg struct{}
type RunHealthcheckResponse struct {
	Id            string `json:"id"`
	ClusterIdHash string `json:"cluster_id_hash"`
	Healthy       bool   `json:"healthy"`
	Time          int64  `json:"time"`
}

func RunHealthcheck(ctx context.Context, c *Client, arg *RunHealthcheckArg) (*RunHealthcheckResponse, *APIError) {
	path := "/v1/healthcheck"

	var resp RunHealthcheckResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}
	return &resp, nil
}
