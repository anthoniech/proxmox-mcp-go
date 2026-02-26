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

func RegisterCreateTools(s *server.MCPServer, c *ProxmoxClient) { //nolint:funlen,gocognit
	s.AddTool(
		mcp.NewTool("create_vm",
			mcp.WithDescription("Create a new QEMU virtual machine"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("VM ID (optional, auto-assigned if empty)"),
			),
			mcp.WithString("name",
				mcp.Description("VM name"),
			),
			mcp.WithString("memory",
				mcp.Description("Memory in MB (e.g. 2048)"),
			),
			mcp.WithString("cores",
				mcp.Description("Number of CPU cores"),
			),
			mcp.WithString("sockets",
				mcp.Description("Number of CPU sockets"),
			),
			mcp.WithString("cpu",
				mcp.Description("CPU type (e.g. host)"),
			),
			mcp.WithString("net0",
				mcp.Description("Network config (e.g. virtio,bridge=vmbr0)"),
			),
			mcp.WithString("scsi0",
				mcp.Description("SCSI disk (e.g. local-lvm:32)"),
			),
			mcp.WithString("ide2",
				mcp.Description("IDE device, typically CD-ROM (e.g. local:iso/ubuntu.iso,media=cdrom)"),
			),
			mcp.WithString("ostype",
				mcp.Description("OS type (e.g. l26 for Linux, win11 for Windows)"),
			),
			mcp.WithString("boot",
				mcp.Description("Boot order (e.g. order=scsi0;ide2;net0)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data := url.Values{}
			optionalParams := []string{
				"vmid",
				"name",
				"memory",
				"cores",
				"sockets",
				"cpu",
				"net0",
				"scsi0",
				"ide2",
				"ostype",
				"boot",
			}
			for _, p := range optionalParams {
				if v := req.GetString(p, ""); v != "" {
					data.Set(p, v)
				}
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/qemu", node), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_container",
			mcp.WithDescription("Create a new LXC container"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("Container ID (optional, auto-assigned if empty)"),
			),
			mcp.WithString("hostname",
				mcp.Description("Container hostname"),
			),
			mcp.WithString("ostemplate",
				mcp.Description("OS template (e.g. local:vztmpl/debian-12-standard_12.2-1_amd64.tar.zst)"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Storage for rootfs (e.g. local-lvm)"),
			),
			mcp.WithString("rootfs",
				mcp.Description("Root filesystem (e.g. local-lvm:8)"),
			),
			mcp.WithString("memory",
				mcp.Description("Memory in MB"),
			),
			mcp.WithString("swap",
				mcp.Description("Swap in MB"),
			),
			mcp.WithString("cores",
				mcp.Description("Number of CPU cores"),
			),
			mcp.WithString("net0",
				mcp.Description("Network config (e.g. name=eth0,bridge=vmbr0,ip=dhcp)"),
			),
			mcp.WithString("password",
				mcp.Description("Root password"),
			),
			mcp.WithString("ssh_public_keys",
				mcp.Description("SSH public keys (URL encoded)"),
			),
			mcp.WithString("unprivileged",
				mcp.Description("Unprivileged container (1 or 0)"),
			),
			mcp.WithString("start",
				mcp.Description("Start after creation (1 or 0)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ostemplate, err := req.RequireString("ostemplate")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data := url.Values{}
			data.Set("ostemplate", ostemplate)

			optionalParams := []string{
				"vmid",
				"hostname",
				"storage",
				"rootfs",
				"memory",
				"swap",
				"cores",
				"net0",
				"password",
				"unprivileged",
				"start",
			}
			for _, p := range optionalParams {
				if v := req.GetString(p, ""); v != "" {
					data.Set(p, v)
				}
			}
			if v := req.GetString("ssh_public_keys", ""); v != "" {
				data.Set("ssh-public-keys", v)
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/lxc", node), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("clone_guest",
			mcp.WithDescription("Clone an existing VM or container"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("Source VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("type",
				mcp.Description("Guest type: qemu or lxc (default: qemu)"),
			),
			mcp.WithString("newid",
				mcp.Description("New VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("New name (or hostname for LXC)"),
			),
			mcp.WithString("full",
				mcp.Description("Full clone (1) or linked clone (0)"),
			),
			mcp.WithString("storage",
				mcp.Description("Target storage for the clone"),
			),
			mcp.WithString("target",
				mcp.Description("Target node for cross-node cloning"),
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
			newid, err := req.RequireString("newid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			guestType := req.GetString("type", "qemu")

			data := url.Values{}
			data.Set("newid", newid)

			for _, p := range []string{"name", "full", "storage", "target"} {
				if v := req.GetString(p, ""); v != "" {
					data.Set(p, v)
				}
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/clone", node, guestType, vmid), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_guest",
			mcp.WithDescription("Delete a VM or container (must be stopped)"),
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
			mcp.WithString("purge",
				mcp.Description("Purge from all configurations (1 or 0)"),
			),
			mcp.WithString("destroy_unreferenced_disks",
				mcp.Description("Destroy unreferenced disks (1 or 0)"),
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

			params := url.Values{}
			if v := req.GetString("purge", ""); v != "" {
				params.Set("purge", v)
			}
			if v := req.GetString("destroy_unreferenced_disks", ""); v != "" {
				params.Set("destroy-unreferenced-disks", v)
			}

			result, err := c.Delete(ctx, fmt.Sprintf("/nodes/%s/%s/%s", node, guestType, vmid), params)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("convert_to_template",
			mcp.WithDescription("Convert a VM or container to a template"),
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

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/%s/%s/template", node, guestType, vmid), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
