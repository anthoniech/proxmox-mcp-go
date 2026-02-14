package mcp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterStorageTools(s *server.MCPServer, c *ProxmoxClient) { //nolint:funlen
	s.AddTool(
		mcp.NewTool("list_storage",
			mcp.WithDescription("List storage pools on a node"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/storage", node))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_templates",
			mcp.WithDescription("List available container templates on a storage"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Storage name (default: local)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			storage := req.GetString("storage", "local")

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content?content=vztmpl", node, storage))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_isos",
			mcp.WithDescription("List available ISO images on a storage"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Storage name (default: local)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			storage := req.GetString("storage", "local")

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content?content=iso", node, storage))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("download_template",
			mcp.WithDescription("Download a container template from the Proxmox repository"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Target storage (e.g. local)"),
				mcp.Required(),
			),
			mcp.WithString("template",
				mcp.Description("Template name (e.g. debian-12-standard_12.2-1_amd64.tar.zst)"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			storage, err := req.RequireString("storage")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			template, err := req.RequireString("template")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data := url.Values{}
			data.Set("storage", storage)
			data.Set("template", template)

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/aplinfo", node), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
