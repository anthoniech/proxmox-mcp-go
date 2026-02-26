// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterGuestTools(s *server.MCPServer, c *ProxmoxClient) { //nolint:funlen,gocognit
	s.AddTool(
		mcp.NewTool("list_vms",
			mcp.WithDescription("List all QEMU virtual machines on a node"),
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
			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/qemu", node))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_containers",
			mcp.WithDescription("List all LXC containers on a node"),
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
			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/lxc", node))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_cluster_resources",
			mcp.WithDescription("List all VMs and containers across the entire cluster"),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := c.Get(ctx, "/cluster/resources?type=vm")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("start_guest",
			mcp.WithDescription("Start a VM or container"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("type",
				mcp.Description("Guest type: qemu or lxc (default: qemu)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err := req.RequireString("vmid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType := req.GetString("type", "qemu")

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/status/start", node, guestType, vmid), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("stop_guest",
			mcp.WithDescription("Stop, shutdown, or reboot a VM or container"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("type",
				mcp.Description("Guest type: qemu or lxc (default: qemu)"),
			),
			mcp.WithString("action",
				mcp.Description("Action: stop, shutdown, or reboot (default: shutdown)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err := req.RequireString("vmid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType := req.GetString("type", "qemu")
			action := req.GetString("action", "shutdown")

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/status/%s", node, guestType, vmid, action), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_next_id",
			mcp.WithDescription("Get the next available VMID in the cluster"),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := c.Get(ctx, "/cluster/nextid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
