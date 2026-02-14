package mcp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTaskTools(s *server.MCPServer, c *ProxmoxClient) {
	s.AddTool(
		mcp.NewTool("list_tasks",
			mcp.WithDescription("List recent tasks on a node"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("limit",
				mcp.Description("Max number of tasks to return (default: 10)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			limit := req.GetString("limit", "10")

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/tasks?limit=%s", node, url.QueryEscape(limit)))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_task_status",
			mcp.WithDescription("Get the status of a specific task by UPID"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("upid",
				mcp.Description("Task UPID"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			upid, err := req.RequireString("upid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/tasks/%s/status", node, url.PathEscape(upid)))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_task_log",
			mcp.WithDescription("Get the log output of a specific task"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("upid",
				mcp.Description("Task UPID"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			upid, err := req.RequireString("upid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/tasks/%s/log", node, url.PathEscape(upid)))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
