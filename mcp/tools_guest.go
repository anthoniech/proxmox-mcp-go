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
		mcp.NewTool("get_guest_config",
			mcp.WithDescription("Get the configuration of a VM or container (disk layout, NIC config, boot order, etc.)"),
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

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/%s/%s/config", node, guestType, vmid))
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

	s.AddTool(
		mcp.NewTool("update_guest_config",
			mcp.WithDescription("Update the configuration of a VM or container (memory, CPU, network, etc.)"),
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
			mcp.WithString("memory",
				mcp.Description("Memory in MB"),
			),
			mcp.WithString("cores",
				mcp.Description("Number of CPU cores"),
			),
			mcp.WithString("sockets",
				mcp.Description("Number of CPU sockets"),
			),
			mcp.WithString("cpu",
				mcp.Description("CPU type (e.g. host, kvm64)"),
			),
			mcp.WithString("net0",
				mcp.Description("Network device configuration"),
			),
			mcp.WithString("name",
				mcp.Description("VM/container name"),
			),
			mcp.WithString("description",
				mcp.Description("Description/notes"),
			),
			mcp.WithString("boot",
				mcp.Description("Boot order (e.g. order=scsi0;net0)"),
			),
			mcp.WithString("onboot",
				mcp.Description("Start at boot: 1 or 0"),
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

			data := url.Values{}
			for _, key := range []string{"memory", "cores", "sockets", "cpu", "net0", "name", "description", "boot", "onboot"} {
				if v := req.GetString(key, ""); v != "" {
					data.Set(key, v)
				}
			}

			if len(data) == 0 {
				return mcp.NewToolResultError("at least one config field must be provided"), nil
			}

			result, err := c.Put(ctx, fmt.Sprintf("/nodes/%s/%s/%s/config", node, guestType, vmid), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("migrate_guest",
			mcp.WithDescription("Migrate a VM or container to another node"),
			mcp.WithString("node",
				mcp.Description("Source node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("type",
				mcp.Description("Guest type: qemu or lxc (default: qemu)"),
			),
			mcp.WithString("target",
				mcp.Description("Target node name"),
				mcp.Required(),
			),
			mcp.WithString("online",
				mcp.Description("Live migration: 1 or 0 (default: 0)"),
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
			target, err := req.RequireString("target")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType := req.GetString("type", "qemu")

			data := url.Values{}
			data.Set("target", target)
			if online := req.GetString("online", ""); online != "" {
				data.Set("online", online)
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/migrate", node, guestType, vmid), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("resize_guest_disk",
			mcp.WithDescription("Resize a disk of a VM or container"),
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
			mcp.WithString("disk",
				mcp.Description("Disk name (e.g. scsi0, virtio0, rootfs)"),
				mcp.Required(),
			),
			mcp.WithString("size",
				mcp.Description("New size or increment (e.g. +10G, 50G)"),
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
			disk, err := req.RequireString("disk")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			size, err := req.RequireString("size")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType := req.GetString("type", "qemu")

			data := url.Values{}
			data.Set("disk", disk)
			data.Set("size", size)

			result, err := c.Put(ctx, fmt.Sprintf("/nodes/%s/%s/%s/resize", node, guestType, vmid), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
