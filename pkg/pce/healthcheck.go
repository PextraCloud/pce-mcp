package pce

import (
	"context"

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetPCEHealthcheck() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pce_healthcheck",
		mcp.WithDescription("Check whether the current Pextra CloudEnvironment node is healthy and return basic node health information."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get PCE Healthcheck",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	), handleGetPCEHealthcheck
}

func handleGetPCEHealthcheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	health, healthErr := api.RunHealthcheck(ctx, client, &api.RunHealthcheckArg{})
	if healthErr != nil {
		return mcp.NewToolResultError(healthErr.Error()), nil
	}

	return mcp.NewToolResultJSON(health)
}
