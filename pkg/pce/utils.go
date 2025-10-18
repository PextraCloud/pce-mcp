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
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
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
