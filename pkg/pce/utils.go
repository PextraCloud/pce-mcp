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
	"strings"

	"github.com/PextraCloud/pce-mcp/internal/session"
	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const hierarchyHelpText = `\n\nHierarchy: [Organization -> Datacenters -> Clusters -> Nodes -> Instances]`

const nameDefaultMinLength = 3
const nameDefaultMaxLength = 64

const descriptionDefaultMaxLength = 512

func nameRegex(min, max int) string {
	return fmt.Sprintf("^[a-zA-Z0-9_-]{%d,%d}$", min, max)
}

var pageNum = mcp.WithNumber("page",
	mcp.Min(1),
	mcp.DefaultNumber(1),
	mcp.Description("The page number for paginated results. Default is 1."),
)

// From: https://github.com/github/github-mcp-server/blob/0188cc0041d86daec4080ef2e48de238919c7909/pkg/github/server.go#L68
// requiredParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request.
// 2. Checks if the parameter is of the expected type.
// 3. Checks if the parameter is not empty, i.e: non-zero value
func requiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.GetArguments()[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	val, ok := r.GetArguments()[p].(T)
	if !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if val == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	return val, nil
}

// From: https://github.com/github/github-mcp-server/blob/0188cc0041d86daec4080ef2e48de238919c7909/pkg/github/server.go#L106
// optionalParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, it checks if the parameter is of the expected type and returns it
func optionalParam[T any](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.GetArguments()[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.GetArguments()[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, r.GetArguments()[p])
	}

	return r.GetArguments()[p].(T), nil
}

// Same as optionalParam, but returns a pointer to the value instead of the value itself. This allows for distinguishing between a missing parameter and a parameter with a zero value.
func optionalParamPtr[T any](r mcp.CallToolRequest, p string) (*T, error) {
	// Check if the parameter is present in the request
	if _, ok := r.GetArguments()[p]; !ok {
		return nil, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.GetArguments()[p].(T); !ok {
		return nil, fmt.Errorf("parameter %s is not of type %T, is %T", p, *new(T), r.GetArguments()[p])
	}

	val := r.GetArguments()[p].(T)
	return &val, nil
}

func clientForRequest(ctx context.Context, req mcp.CallToolRequest) (*api.Client, error) {
	s := server.ClientSessionFromContext(ctx)
	if s == nil {
		return nil, fmt.Errorf("missing session context")
	}
	var authorization string
	if req.Header != nil {
		authorization = req.Header.Get("Authorization")
	}
	client, err := session.GetSession(s.SessionID())
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Update authorization header based on request `Authorization` header
	if authorization == "" {
		client.Headers.Del("Authorization")
	} else {
		client.Headers.Set("Authorization", authorization)
	}

	return client, nil
}

func mcpToolOptionStringFilter(name, description string) mcp.ToolOption {
	return mcp.WithString(name,
		mcp.Description(fmt.Sprintf("Optionally filter for matching %s, case-insensitive, partial match", description)),
	)
}

func applyStringFilter(value, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(filter))
}

func mcpToolOptionNumberFilter(name, description string) mcp.ToolOption {
	return mcp.WithObject(name,
		mcp.Description(fmt.Sprintf("Optionally filter for %s, using the following operators: gt (greater than), lt (less than), eq (equal to)", description)),
		mcp.Properties(map[string]any{
			"gt": map[string]any{"type": "number", "description": fmt.Sprintf("Filter for %s greater than this value", description)},
			"lt": map[string]any{"type": "number", "description": fmt.Sprintf("Filter for %s less than this value", description)},
			"eq": map[string]any{"type": "number", "description": fmt.Sprintf("Filter for %s equal to this value", description)},
		}),
	)
}

func convertFilterToInt(filter map[string]any) map[string]int {
	if filter == nil {
		return nil
	}

	intFilter := make(map[string]int)
	for key, value := range filter {
		// JSON numbers are unmarshaled as float64, so we need to convert them to int
		if floatVal, ok := value.(float64); ok {
			intFilter[key] = int(floatVal)
		}
	}
	return intFilter
}

func applyNumberFilter(value int, filter map[string]int) bool {
	if filter == nil {
		return true
	}

	if gt, ok := filter["gt"]; ok && value <= gt {
		return false
	}
	if lt, ok := filter["lt"]; ok && value >= lt {
		return false
	}
	if eq, ok := filter["eq"]; ok && value != eq {
		return false
	}

	return true
}

func mcpToolOptionBoolFilter(name, description string) mcp.ToolOption {
	return mcp.WithBoolean(name,
		mcp.Description(fmt.Sprintf("Optionally filter for %s, using either 'true' or 'false'", description)),
	)
}

func applyBoolFilter(value bool, filter *bool) bool {
	if filter == nil {
		return true
	}
	return value == *filter
}

var mcpToolOptionDestroyConfirmation = mcp.WithBoolean("are_you_sure",
	mcp.Required(),
	mcp.Description("Must be set to true to confirm the destruction; this action is irreversible"),
)

func confirmDestructiveAction(req mcp.CallToolRequest) bool {
	areYouSure, err := requiredParam[bool](req, "are_you_sure")
	if err != nil {
		return false
	}
	return areYouSure
}

func getCurrentNodeId(ctx context.Context, client *api.Client) (string, error) {
	health, err := api.RunHealthcheck(ctx, client, &api.RunHealthcheckArg{})
	if err != nil {
		return "", err
	}
	return health.Id, nil
}

type currentTreeIdsResult struct {
	NodeId         string
	ClusterId      string
	OrganizationId string
}

func getCurrentTreeIds(ctx context.Context, client *api.Client) (*currentTreeIdsResult, error) {
	// TODO: caching
	nodeId, err := getCurrentNodeId(ctx, client)
	if err != nil {
		return nil, err
	}

	node, getErr := api.GetNodeById(ctx, client, &api.GetNodeByIdArg{NodeId: nodeId})
	if getErr != nil {
		return nil, getErr
	}
	return &currentTreeIdsResult{
		NodeId:         nodeId,
		ClusterId:      node.Node.ClusterId,
		OrganizationId: node.Node.OrganizationId,
	}, nil
}
