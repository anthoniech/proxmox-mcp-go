package mcp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterBackupTools(s *server.MCPServer, c *ProxmoxClient) { //nolint:funlen,gocognit
	s.AddTool(
		mcp.NewTool("backup_guest",
			mcp.WithDescription("Create a backup (vzdump) of a VM or container"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("VM/container ID"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Target storage for backup (e.g. local)"),
			),
			mcp.WithString("mode",
				mcp.Description("Backup mode: snapshot, suspend, or stop"),
			),
			mcp.WithString("compress",
				mcp.Description("Compression: zstd, lzo, gzip, or 0 for none"),
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

			data := url.Values{}
			data.Set("vmid", vmid)
			for _, p := range []string{"storage", "mode", "compress"} {
				if v := req.GetString(p, ""); v != "" {
					data.Set(p, v)
				}
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/vzdump", node), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_backups",
			mcp.WithDescription("List backup files on a storage"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Storage name (e.g. local)"),
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

			result, err := c.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content?content=backup", node, storage))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	s.AddTool(
		mcp.NewTool("restore_backup",
			mcp.WithDescription("Restore a VM from a backup archive"),
			mcp.WithString("node",
				mcp.Description("Node name"),
				mcp.Required(),
			),
			mcp.WithString("vmid",
				mcp.Description("Target VM ID for the restored VM"),
				mcp.Required(),
			),
			mcp.WithString("archive",
				mcp.Description("Backup archive path (e.g. local:backup/vzdump-qemu-100-2024_01_01-12_00_00.vma.zst)"),
				mcp.Required(),
			),
			mcp.WithString("storage",
				mcp.Description("Target storage for restored disks"),
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
			archive, err := req.RequireString("archive")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data := url.Values{}
			data.Set("vmid", vmid)
			data.Set("archive", archive)
			if v := req.GetString("storage", ""); v != "" {
				data.Set("storage", v)
			}

			result, err := c.Post(ctx, fmt.Sprintf("/nodes/%s/qemu", node), data)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
}
