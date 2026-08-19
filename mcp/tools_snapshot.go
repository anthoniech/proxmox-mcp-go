// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package mcp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//nolint:funlen,gocognit,gocyclo,cyclop // MCP tool registration with input validation
func RegisterSnapshotTools(
	s *server.MCPServer,
	c *ProxmoxClient,
) {
	s.AddTool(
		mcp.NewTool("list_snapshots",
			mcp.WithDescription("List all snapshots of a VM or container"),
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
			node, err = validatePathSegment("node", node)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err = parseVMID(vmid)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType, err := validateGuestType(req.GetString("type", "qemu"))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, guestType, vmid))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_snapshot",
			mcp.WithDescription("Create a snapshot of a VM or container"),
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
			mcp.WithString("snapname",
				mcp.Description("Snapshot name"),
				mcp.Required(),
			),
			mcp.WithString("description",
				mcp.Description("Snapshot description"),
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
			snapname, err := req.RequireString("snapname")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			node, err = validatePathSegment("node", node)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err = parseVMID(vmid)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType, err := validateGuestType(req.GetString("type", "qemu"))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data := url.Values{}
			data.Set("snapname", snapname)
			if desc := req.GetString("description", ""); desc != "" {
				data.Set("description", desc)
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, guestType, vmid), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("rollback_snapshot",
			mcp.WithDescription("Rollback a VM or container to a snapshot. Requires confirm=true to execute."),
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
			mcp.WithString("snapname",
				mcp.Description("Snapshot name to rollback to"),
				mcp.Required(),
			),
			mcp.WithString("confirm",
				mcp.Description("Must be set to 'true' to confirm this destructive operation"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			confirm := req.GetString("confirm", "")
			if confirm != confirmValue {
				return mcp.NewToolResultError(
					"destructive operation: set confirm='true' to rollback this snapshot",
				), nil
			}
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err := req.RequireString("vmid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			snapname, err := req.RequireString("snapname")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			node, err = validatePathSegment("node", node)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err = parseVMID(vmid)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			snapname, err = validatePathSegment("snapname", snapname)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType, err := validateGuestType(req.GetString("type", "qemu"))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			result, err := c.Post(
				ctx,
				fmt.Sprintf("/nodes/%s/%s/%s/snapshot/%s/rollback", node, guestType, vmid, snapname),
				nil,
			)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_snapshot",
			mcp.WithDescription("Delete a snapshot of a VM or container"),
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
			mcp.WithString("snapname",
				mcp.Description("Snapshot name to delete"),
				mcp.Required(),
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
			snapname, err := req.RequireString("snapname")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			node, err = validatePathSegment("node", node)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vmid, err = parseVMID(vmid)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			snapname, err = validatePathSegment("snapname", snapname)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType, err := validateGuestType(req.GetString("type", "qemu"))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			result, err := c.Delete(
				ctx,
				fmt.Sprintf("/nodes/%s/%s/%s/snapshot/%s", node, guestType, vmid, snapname),
				nil,
			)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
